package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/andig/evcc/api"
	"github.com/andig/evcc/core"
	"github.com/andig/evcc/provider"
	"github.com/spf13/viper"
)

// compositeCharger combines Charger and ChargeController
type compositeCharger struct {
	api.Charger
	api.ChargeController
}

// compositeCharger combines Meter and MeterEnergy
type compositeMeter struct {
	api.Meter
	api.MeterEnergy
}

// MQTT singleton
var mq *provider.MqttClient

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

func clientID() string {
	pid := os.Getpid()
	return fmt.Sprintf("evcc-%d", pid)
}

func configureLoadPoint(lp *core.LoadPoint, lpc loadPointConfig) {
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

func configureMeters(conf config) (meters map[string]api.Meter) {
	meters = make(map[string]api.Meter)
	for _, mc := range conf.Meters {
		m := core.NewMeter(
			floatProvider(mc.Power),
		)

		if mc.Energy != nil {
			m = &compositeMeter{
				m,
				core.NewMeterEnergy(floatProvider(mc.Energy)),
			}
		}
		meters[mc.Name] = m
	}
	return
}

func configureChargers(conf config) (chargers map[string]api.Charger) {
	chargers = make(map[string]api.Charger)
	for _, cc := range conf.Chargers {
		var c api.Charger

		switch cc.Type {
		case "wallbe":
			c = provider.NewWallbe(cc.URI)

		case "configurable":
			c = core.NewCharger(
				stringProvider(cc.Status),
				intProvider(cc.ActualCurrent),
				boolProvider(cc.Enabled),
				boolSetter("enable", cc.Enable),
			)

			// if chargecontroller specified build composite charger
			if cc.MaxCurrent != nil {
				c = &compositeCharger{
					c,
					core.NewChargeController(
						intSetter("current", cc.MaxCurrent),
					),
				}
			}
		default:
			log.Fatalf("invalid charger type '%s'", cc.Type)
		}

		chargers[cc.Name] = c
	}
	return
}

func loadConfig(conf config) {
	if viper.Get("mqtt") != nil {
		mq = provider.NewMqttClient(conf.Mqtt.Broker, conf.Mqtt.User, conf.Mqtt.Password, clientID(), true, 1)
	}

	meters := configureMeters(conf)
	chargers := configureChargers(conf)

	for _, lpc := range conf.LoadPoints {
		charger, ok := chargers[lpc.Charger]
		if !ok {
			log.Fatalf("invalid charger '%s'", lpc.Charger)
		}
		lp := core.NewLoadPoint(
			lpc.Name,
			charger,
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
					log.Fatalf("invalid meter '%s'", m.key)
				}
			}
		}

		// assign remaing config
		configureLoadPoint(lp, lpc)
		loadPoints = append(loadPoints, lp)
	}
}
