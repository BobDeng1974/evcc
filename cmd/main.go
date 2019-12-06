package main

import (
	"log"
	"time"

	"github.com/andig/ulm"
)

type meter struct{}

func (m *meter) CurrentPower() (float64, error) {
	return 500, nil
}

type charger struct{}

func (c *charger) Enable(charge bool) error {
	return nil
}

func (c *charger) Status() (ulm.ChargeStatus, error) {
	return ulm.StatusC, nil
}

func (c *charger) Enabled() (bool, error) {
	return true, nil
}

func (c *charger) MaxPower(max float64) error {
	return nil
}

func control(c ulm.Charger) {
	if c, ok := c.(ulm.ChargeController); ok {
		log.Println("Maxpower")
		c.MaxPower(1)
	} else {
		log.Println("no Maxpower")
	}
}

type LoadPoint struct {
	name     string
	meter    ulm.Meter
	charger  ulm.Charger
	strategy ulm.Strategy
}

var loadPoints []LoadPoint

func (lp *LoadPoint) Update() {
	if lp.charger == nil {
		return
	}

	debug := ulm.LogEnabled()

	s, err := lp.charger.Status()
	if err != nil {
		log.Printf("%s charger error: %v", lp.name, err)
		return
	}

	if debug {
		log.Printf("%s charger status: %s", lp.name, s)
	}

	// vehicle connected
	if s == ulm.StatusC || s == ulm.StatusD {
		enabled, err := lp.charger.Enabled()
		if err != nil {
			log.Printf("%s charger error: %v", lp.name, err)
			return
		}

		if debug {
			log.Printf("%s charger enabled: %v", lp.name, enabled)
		}

		if enabled {
			if err := lp.ApplyStrategy(); err != nil {
				log.Printf("%s charger error: %v", lp.name, err)
				return
			}
		} else {
			if err := lp.charger.Enable(true); err != nil {
				log.Printf("%s charger error: %v", lp.name, err)
				return
			}
		}
	}
}

func (lp *LoadPoint) ApplyStrategy() error {
	if lp.meter == nil {
		return nil
	}

	debug := ulm.LogEnabled()

	power, err := lp.meter.CurrentPower()
	if err != nil {
		log.Printf("%s meter error: %v", lp.name, err)
		return err
	}

	if debug {
		log.Printf("%s meter power: %.0f", lp.name, power)
	}

	if charger, ok := lp.charger.(ulm.ChargeController); ok {
		maxPower := -power - 250 // apply margin

		if err := charger.MaxPower(maxPower); err != nil {
			log.Printf("%s charge controller error: %v", lp.name, err)
			return err
		}
	}

	return nil
}

func main() {
	m := &meter{}
	log.Println(m.CurrentPower())

	c := &charger{}
	// c.MaxPower(5)
	c.Enable(true)

	voltage := 230
	phases := 3
	current := 16
	power := phases * current * voltage
	log.Println(power)

	loadPoints = append(loadPoints, LoadPoint{
		name:     "lp1",
		meter:    m,
		charger:  c,
		strategy: nil,
	})

	for range time.Tick(time.Second) {
		for _, lp := range loadPoints {
			go lp.Update()
		}
	}
}
