package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/andig/ulm/api"
	"github.com/andig/ulm/core"
	"github.com/andig/ulm/provider"
	"github.com/andig/ulm/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	url = "127.1:7070"
)

var (
	cfgFile    string
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

type CompositeCharger struct {
	api.Charger
	api.ChargeController
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
		log.Println(conf)
	}

	// mqtt provider
	mq := provider.NewMqttClient("nas.fritz.box:1883", "", "", clientID(), true, 1)

	// charger
	exec := provider.Exec{}
	charger := &CompositeCharger{
		core.NewCharger(
			exec.StringProvider("/bin/bash -c 'echo C'"),
			exec.IntProvider("/bin/bash -c 'echo $((RANDOM % 32))'"),
			exec.BoolProvider("/bin/bash -c 'echo true'"),
			exec.BoolSetter("enable", "/bin/bash -c 'echo true'"),
		),
		core.NewChargeController(
			exec.IntSetter("current", "/bin/bash -c 'echo $((RANDOM % 1000))'"),
		),
	}

	// meters
	gridMeter := core.NewMeter(mq.FloatValue("mbmd/sdm1-1/Power"))
	pvMeter := core.NewMeter(mq.FloatValue("mbmd/sdm1-2/Power"))

	// loadpoint
	lp := core.NewLoadPoint("lp1", charger)
	lp.Phases = 2      // Audi
	lp.Voltage = 230   // V
	lp.MinCurrent = 0  // A
	lp.MaxCurrent = 16 // A
	lp.GridMeter = gridMeter
	lp.PVMeter = pvMeter
	lp.ChargeMeter = pvMeter
	loadPoints = append(loadPoints, lp)

	if err := lp.ChargeMode(api.ModePV); err != nil {
		log.Fatal(err)
	}

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
