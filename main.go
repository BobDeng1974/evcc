package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/andig/ulm/api"
	"github.com/andig/ulm/core"
	"github.com/andig/ulm/exec"
	"github.com/andig/ulm/provider"
	"github.com/andig/ulm/server"
)

const (
	url     = "127.1:7070"
	timeout = 1 * time.Second
)

type charger struct {
	api.Charger
	api.ChargeController
}

var loadPoints []*core.LoadPoint

func updateLoadPoints() {
	for _, lp := range loadPoints {
		go lp.Update()
	}
}

func logEnabled() bool {
	env := strings.TrimSpace(os.Getenv("ULM_DEBUG"))
	return env != "" && env != "0"
}

func clientID() string {
	pid := os.Getpid()
	return fmt.Sprintf("ulm-%d", pid)
}

func chargeModeObserver(lp *core.LoadPoint) api.StringProvider {
	return func(ctx context.Context) (string, error) {
		return string(lp.CurrentChargeMode()), nil
	}
}

func chargedEnergyObserver(lp *core.LoadPoint) api.FloatProvider {
	return func(ctx context.Context) (float64, error) {
		return lp.ChargedEnergy()
	}
}

func main() {
	if true || logEnabled() {
		logger := log.New(os.Stdout, "", log.LstdFlags)
		core.Logger = logger
		exec.Logger = logger
	}

	m := exec.NewMeter("/bin/bash -c 'echo $((RANDOM % 1000))'", timeout)
	c := &charger{
		exec.NewCharger(
			"/bin/bash -c 'echo C'",
			"/bin/bash -c 'echo $((RANDOM % 32))'",
			"/bin/bash -c 'echo true'",
			"/bin/bash -c 'echo true'",
			timeout,
		),
		exec.NewChargeController(
			"/bin/bash -c 'echo $((RANDOM % 1000))'",
			timeout,
		),
	}

	// create loadpoint
	lp := core.NewLoadPoint("lp1", c)
	lp.GridMeter = m
	lp.Phases = 2
	lp.Voltage = 230   // V
	lp.MinCurrent = 5  // A
	lp.MaxCurrent = 16 // A
	lp.ChargeMode(api.ModePV)
	lp.Validate()

	loadPoints = append(loadPoints, lp)

	// create webserver
	hub := server.NewSocketHub()
	httpd := server.NewHttpd(url, lp, hub)

	// start broadcasting values
	socketChan := make(chan server.SocketValue)
	go hub.Run(socketChan)

	// observe meters
	mq := provider.NewMqttClient("nas.fritz.box:1883", "", "", clientID(), true, 1)
	observer := server.NewObserver(socketChan)
	observer.Observe("gridPower", mq.FloatValue("mbmd/sdm1-1/Power"))
	observer.Observe("pvPower", mq.FloatValue("mbmd/sdm1-2/Power"))
	observer.Observe("mode", chargeModeObserver(lp))
	observer.Observe("socEnergy", chargedEnergyObserver(lp))

	// push updates
	go func() {
		for range time.Tick(time.Second) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			observer.Update(ctx)
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
