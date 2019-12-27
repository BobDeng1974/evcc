package main

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
)

const (
	url     = "127.1:7070"
	timeout = 1 * time.Second
)

type CompositeCharger struct {
	api.Charger
	api.ChargeController
}

var (
	loadPoints []*core.LoadPoint
	clientPush = make(chan server.SocketValue)
)

func updateLoadPoints() {
	for _, lp := range loadPoints {
		go lp.Update()
	}
}

func observeLoadPoint(lp *core.LoadPoint) {
	if lp.GridMeter != nil {
		if f, err := lp.GridMeter.CurrentPower(); err == nil {
			clientPush <- server.SocketValue{Key: "gridPower", Val: f}
		} else {
			log.Printf("%s update grid meter failed: %v", lp.Name, err)
		}
	}

	if lp.PVMeter != nil {
		if f, err := lp.PVMeter.CurrentPower(); err == nil {
			clientPush <- server.SocketValue{Key: "pvPower", Val: f}
		} else {
			log.Printf("%s update pv meter failed: %v", lp.Name, err)
		}
	}

	if lp.ChargeMeter != nil {
		if f, err := lp.ChargedEnergy(); err == nil {
			clientPush <- server.SocketValue{Key: "socEnergy", Val: f}
		} else {
			log.Printf("%s update soc meter failed: %v", lp.Name, err)
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

func main() {
	if true || logEnabled() {
		logger := log.New(os.Stdout, "", log.LstdFlags)
		core.Logger = logger
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
	lp.ChargeMode(api.ModePV)

	loadPoints = append(loadPoints, lp)

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
