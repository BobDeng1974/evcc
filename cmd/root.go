package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andig/evcc/api"
	"github.com/andig/evcc/core"
	"github.com/andig/evcc/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	Run:   run,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP(
		"uri", "u",
		"0.0.0.0:8080",
		"Listen address",
	)
	viper.BindPFlag("uri", rootCmd.PersistentFlags().Lookup("uri"))

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
		lp.Update()
	}
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d:%02d", int64(d.Hours()), int64(d.Minutes())%60, int64(d.Seconds())%60)
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

	clientPush <- server.SocketValue{Key: "chargeDuration", Val: formatDuration(lp.ChargeDuration())}
	clientPush <- server.SocketValue{Key: "mode", Val: string(lp.CurrentChargeMode())}

	if f, err := lp.ChargedEnergy(); err == nil {
		clientPush <- server.SocketValue{Key: "chargedEnergy", Val: f}
	} else {
		log.Printf("%s update charge meter failed: %v", lp.Name, err)
	}

	if f, err := lp.Charger.ActualCurrent(); err == nil {
		clientPush <- server.SocketValue{Key: "chargeCurrent", Val: f}
	} else {
		log.Printf("%s update charger current failed: %v", lp.Name, err)
	}
}

func run(cmd *cobra.Command, args []string) {
	if true {
		logger := log.New(os.Stdout, "", log.LstdFlags)
		core.Logger = logger
	}

	if cfgFile != "" {
		var conf config
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
	httpd := server.NewHttpd(viper.GetString("uri"), lp, hub)

	// start broadcasting values
	go hub.Run(clientPush)

	// push updates
	go func() {
		for range time.Tick(1 * time.Second) {
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
