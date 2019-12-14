package core

import (
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/andig/ulm/api"
)

type LoadPoint struct {
	Name       string
	Mode       api.ChargeMode
	GridMeter  api.Meter
	UsageMeter api.Meter
	Charger    api.Charger
	Strategy   api.Strategy
	MinCurrent float64
	MaxCurrent float64
	Phases     float64
	Debug      bool
}

func (lp *LoadPoint) CurrentChargeMode() api.ChargeMode {
	return lp.Mode
}

func (lp *LoadPoint) SetChargeMode(mode api.ChargeMode) error {
	if lp.Debug {
		log.Printf("%s set charge mode: %s", lp.Name, string(mode))
	}

	switch mode {
	// both modes require GridMeter
	case api.ModeMinPV, api.ModePV:
		if lp.GridMeter == nil {
			return errors.New("invalid charge mode: " + string(mode))
		}
	}

	lp.Mode = mode
	return nil
}

func (lp *LoadPoint) Update() {
	if lp.Charger == nil {
		return
	}

	s, err := lp.Charger.Status()
	if err != nil {
		log.Printf("%s charger error: %v", lp.Name, err)
		return
	}

	if lp.Debug {
		log.Printf("%s charger status: %s", lp.Name, s)
	}

	// vehicle connected
	if s == api.StatusC || s == api.StatusD {
		enabled, err := lp.Charger.Enabled()
		if err != nil {
			log.Printf("%s charger error: %v", lp.Name, err)
			return
		}

		if lp.Debug {
			log.Printf("%s charger enabled: %v", lp.Name, enabled)
		}

		if enabled {
			if err := lp.ApplyStrategy(); err != nil {
				log.Printf("%s charger error: %v", lp.Name, err)
				return
			}
		}
	}
}

func (lp *LoadPoint) ApplyStrategy() error {
	var gridpower, maxPower float64

	// get grid power
	if lp.GridMeter != nil {
		var err error
		gridpower, err = lp.GridMeter.CurrentPower()
		if err != nil {
			log.Printf("%s meter error: %v", lp.Name, err)
			return err
		}
	}

	if lp.Debug {
		log.Printf("%s grid meter power: %.0f", lp.Name, gridpower)
	}

	// negative gridpower means excess production
	availablepower := math.Max(0, -gridpower)

	switch lp.Mode {
	case api.ModeMinPV:
		maxPower = math.Max(250, availablepower)
	case api.ModePV:
		maxPower = availablepower
	}

	if lp.Debug {
		log.Printf("%s charge power: %.0f", lp.Name, availablepower)
	}

	if charger, ok := lp.Charger.(api.ChargeController); ok {
		if err := charger.MaxPower(maxPower); err != nil {
			return fmt.Errorf("charge controller error: %v", err)
		}
	} else {
		log.Printf("%s has no charge controller", lp.Name)
	}

	// enable charging if not already
	on, err := lp.Charger.Enabled()
	if err != nil {
		return err
	}
	if !on {
		if err := lp.Charger.Enable(true); err != nil {
			return err
		}
	}

	return nil
}
