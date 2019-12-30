package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/andig/evcc/api"
	"github.com/andig/evcc/core"
	"github.com/andig/evcc/provider"
	"github.com/andig/evcc/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	url = "127.1:7070"
)

var (
	cfgFile    string
	mq         *provider.MqttClient
	loadPoints []*core.LoadPoint
	clientPush = make(chan server.SocketValue)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "evcc",
	Short: "EV Charge Controller",
	// Long:  "Easily read and distribute data from ModBus meters and grid inverters",
	Run: run,
}

type CompositeCharger struct {
	api.Charger
	api.ChargeController
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile,
		"config", "c",
		"",
		"Config file (default is $HOME/evcc.yaml)",
	)
	rootCmd.PersistentFlags().BoolP(
		"help", "h",
		false,
		"Help for "+rootCmd.Name(),
	)
	rootCmd.PersistentFlags().BoolP(
		"verbose", "v",
		false,
		"Verbose mode",
	)

	// bind command line options
	// if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
	// 	log.Fatal(err)
	// }
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name "mbmd" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")    // optionally look for config in the working directory
		viper.AddConfigPath("/etc") // path to look for the config file in

		viper.SetConfigName("evcc")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		// using config file
		cfgFile = viper.ConfigFileUsed()
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		// parsing failed - exit
		fmt.Println(err)
		os.Exit(1)
	} else {
		// not using config file
		cfgFile = ""
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func updateLoadPoints() {
	for _, lp := range loadPoints {
		go lp.Update()
	}
}

func observeLoadPoint(lp *core.LoadPoint) {
	meters := map[string]api.Meter{
		"grid":   lp.GridMeter,
		"pv":     lp.PVMeter,
		"charge": lp.ChargeMeter,
	}
	for name, meter := range meters {
		if f, err := meter.CurrentPower(); err == nil {
			key := name + "Power"
			clientPush <- server.SocketValue{Key: key, Val: f}
		} else {
			log.Printf("%s update %s meter failed: %v", lp.Name, name, err)
		}
	}

	if lp.ChargeMeter != nil {
		if f, err := lp.ChargedEnergy(); err == nil {
			clientPush <- server.SocketValue{Key: "chargeEnergy", Val: f}
		} else {
			log.Printf("%s update charge meter failed: %v", lp.Name, err)
		}
	}

	clientPush <- server.SocketValue{Key: "mode", Val: string(lp.CurrentChargeMode())}
}

func observeLoadPoints() {
	var wg sync.WaitGroup
	for _, lp := range loadPoints {
		wg.Add(1)
		go func(lp *core.LoadPoint) {
			observeLoadPoint(lp)
			wg.Done()
		}(lp)
	}
	wg.Wait()
}

func logEnabled() bool {
	env := strings.TrimSpace(os.Getenv("ULM_DEBUG"))
	return env != "" && env != "0"
}

func clientID() string {
	pid := os.Getpid()
	return fmt.Sprintf("ulm-%d", pid)
}

func configureLoadPoint(lp *core.LoadPoint, lpc LoadPointConfig) {
	if lpc.Mode != "" {
		lp.Mode = api.ChargeMode(lpc.Mode)
	}
	if lpc.MinCurrent > 0 {
		lp.MinCurrent = lpc.MinCurrent
	}
	if lpc.MaxCurrent > 0 {
		lp.MaxCurrent = lpc.MaxCurrent
	}
	if lpc.Voltage > 0 {
		lp.Voltage = lpc.Voltage
	}
	if lpc.Phases > 0 {
		lp.Phases = lpc.Phases
	}
}

func loadConfig(conf Config) {
	if viper.Get("mqtt") != nil {
		mq = provider.NewMqttClient(conf.Mqtt.Broker, conf.Mqtt.User, conf.Mqtt.Password, clientID(), true, 1)
	}

	meters := make(map[string]api.Meter)
	for _, mc := range conf.Meters {
		var p api.FloatProvider

		switch mc.Type {
		case "mqtt":
			p = mq.FloatProvider(mc.Power)
		case "exec":
			exec := &provider.Exec{}
			p = exec.FloatProvider(mc.Power)
		default:
			log.Fatalf("invalid meter type %s", mc.Type)
		}

		m := core.NewMeter(p)
		meters[mc.Name] = m
	}

	chargers := make(map[string]api.Charger)
	for _, cc := range conf.Chargers {
		var c api.Charger

		switch cc.Type {
		case "wallbe":
			c = provider.NewWallbe(cc.URI)
		case "configurable":
			status := stringProvider(cc.Status)
			actualCurrent := intProvider(cc.ActualCurrent)
			enable := boolSetter("enable", cc.Enable)
			enabled := boolProvider(cc.Enabled)
			c = core.NewCharger(
				status,
				actualCurrent,
				enabled,
				enable,
			)

			// if chargecontroller specified build composite charger
			if cc.MaxCurrent != nil {
				c = &CompositeCharger{
					c,
					core.NewChargeController(
						intSetter("current", cc.MaxCurrent),
					),
				}
			}
		default:
			log.Fatalf("invalid charger type %s", cc.Type)
		}

		chargers[cc.Name] = c
	}

	for _, lpc := range conf.LoadPoints {
		lp := core.NewLoadPoint(
			lpc.Name,
			chargers[lpc.Charger],
		)

		// assign meters
		for _, m := range []struct {
			key   string
			meter *api.Meter
		}{
			{lpc.GridMeter, &lp.GridMeter},
			{lpc.ChargeMeter, &lp.ChargeMeter},
			{lpc.PVMeter, &lp.PVMeter},
		} {
			if m.key != "" {
				if impl, ok := meters[m.key]; ok {
					*m.meter = impl
				} else {
					log.Fatalf("invalid meter %s", m.key)
				}
			}
		}

		// assign remaing config
		configureLoadPoint(lp, lpc)
		loadPoints = append(loadPoints, lp)
	}
}

func run(cmd *cobra.Command, args []string) {
	if true || logEnabled() {
		logger := log.New(os.Stdout, "", log.LstdFlags)
		core.Logger = logger
	}

	if cfgFile != "" {
		var conf Config
		if err := viper.UnmarshalExact(&conf); err != nil {
			log.Fatalf("config: failed parsing config file %s: %v", cfgFile, err)
		}
		loadConfig(conf)
	} else {
		log.Fatal("missing evcc config")
	}

	lp := loadPoints[0]
	log.Printf("%+v", lp)

	// create webserver
	hub := server.NewSocketHub()
	httpd := server.NewHttpd(url, lp, hub)

	// start broadcasting values
	go hub.Run(clientPush)

	// push updates
	go func() {
		for range time.Tick(2 * time.Second) {
			observeLoadPoints()
		}
	}()

	go func() {
		updateLoadPoints()
		for range time.Tick(5 * time.Second) {
			core.Logger.Printf("---")
			updateLoadPoints()
		}
	}()

	log.Fatal(httpd.ListenAndServe())
}
