package core

import (
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/andig/ulm/api"
)

// LoadPoint is responsible for controlling charge depending on
// SoC needs and power availability.
//
// Power availability is goverened by this equation (positiv sign signals
// consumption, negative sign is grid production):
//
//    HAges = HArest + EV
//
// therefore
//
//    HArest = HAges - EVist
//    EVsoll = -HArest
type LoadPoint struct {
	Name       string
	Mode       api.ChargeMode
	GridMeter  api.Meter // UsageMeter api.Meter
	Charger    api.Charger
	MinCurrent int64 // PV mode: start current	Min+PV mode: min current
	MaxCurrent int64
	Voltage    float64
	Phases     float64
}

// NewLoadPoint creates a LoadPoint with sane defaults
func NewLoadPoint(name string, charger api.Charger, meter api.Meter) *LoadPoint {
	return &LoadPoint{
		Phases:     1,
		Voltage:    230, // V
		MinCurrent: 5,   // A
		MaxCurrent: 16,  // A
		Mode:       api.ModeNow,
		Charger:    charger,
		GridMeter:  meter,
	}
}

// setTargetCurrent guards setting current against changing to identical value
// and violating MaxCurrent
func (lp *LoadPoint) setTargetCurrent(chargeCurrent, targetChargeCurrent int64) error {
	if targetChargeCurrent > lp.MaxCurrent {
		targetChargeCurrent = lp.MaxCurrent
		Logger.Printf("%s limit charge current: %dA", lp.Name, targetChargeCurrent)
	}

	if chargeCurrent != targetChargeCurrent {
		if err := lp.Charger.(api.ChargeController).MaxCurrent(targetChargeCurrent); err != nil {
			return fmt.Errorf("charge controller error: %v", err)
		}
	}

	return nil
}

// chargerEnable switches charger on/off if status
func (lp *LoadPoint) chargerEnable(enable bool) error {
	// get enabled state
	enabled, err := lp.Charger.Enabled()
	if err != nil {
		return err
	}

	// state change required?
	if enable != enabled {
		return lp.Charger.Enable(enable)
	}

	return nil
}

func (lp *LoadPoint) CurrentChargeMode() api.ChargeMode {
	return lp.Mode
}

func (lp *LoadPoint) ChargeMode(mode api.ChargeMode) error {
	Logger.Printf("%s set charge mode: %s", lp.Name, string(mode))

	// check if charger is controllable
	_, chargerControllable := lp.Charger.(api.ChargeController)

	// disable charger if enabled
	if mode == api.ModeOff {
		return lp.chargerEnable(false)
	}

	// both modes require GridMeter
	if mode == api.ModeMinPV || mode == api.ModePV {
		if lp.GridMeter == nil || !chargerControllable {
			return errors.New("invalid charge mode: " + string(mode))
		}
	}

	// enable charger if disabled
	if err := lp.chargerEnable(true); err != nil {
		return err
	}

	lp.Mode = mode
	return nil
}

func (lp *LoadPoint) Update() {
	if lp.Charger == nil {
		panic(fmt.Sprintf("%s no charger assigned", lp.Name))
	}

	Logger.Printf("%s charge mode: %s", lp.Name, lp.Mode)

	status, err := lp.Charger.Status()
	if err != nil {
		log.Printf("%s charger error: %v", lp.Name, err)
		return
	}
	Logger.Printf("%s charger status: %s", lp.Name, status)

	// no vehicle connected
	if status != api.StatusC {
		return
	}

	enabled, err := lp.Charger.Enabled()
	if err != nil {
		log.Printf("%s charger error: %v", lp.Name, err)
		return
	}
	Logger.Printf("%s charger enabled: %v", lp.Name, enabled)
	if !enabled {
		return
	}

	if _, chargeController := lp.Charger.(api.ChargeController); !chargeController {
		log.Printf("%s no charge controller assigned", lp.Name)
		return
	}

	// execute loading strategy
	switch lp.Mode {
	case api.ModeNow:
		err = lp.ApplyModeNow()
	case api.ModeMinPV, api.ModePV:
		err = lp.ApplyModePV()
	}

	if err != nil {
		Logger.Printf("%s error: %v", lp.Name, err)
		return
	}
}

func (lp *LoadPoint) ApplyModeNow() error {
	// get grid power
	if lp.GridMeter != nil {
		gridPower, err := lp.GridMeter.CurrentPower()
		if err != nil {
			log.Printf("%s meter error: %v", lp.Name, err)
			return err
		}
		Logger.Printf("%s grid meter power: %.0fW", lp.Name, gridPower)
	}

	// get charger current
	chargeCurrent, err := lp.Charger.ActualCurrent()
	if err != nil {
		Logger.Printf("%s charger error: %v", lp.Name, err)
		return err
	}
	Logger.Printf("%s charge current: %dA", lp.Name, chargeCurrent)

	// get max charge current
	targetChargeCurrent := lp.MaxCurrent
	Logger.Printf("%s max charge current: %dA", lp.Name, targetChargeCurrent)

	// set max charge current
	if err := lp.setTargetCurrent(chargeCurrent, targetChargeCurrent); err != nil {
		return err
	}

	return nil
}

func (lp *LoadPoint) ApplyModePV() error {
	// get grid power
	gridPower, err := lp.GridMeter.CurrentPower()
	if err != nil {
		log.Printf("%s meter error: %v", lp.Name, err)
		return err
	}
	Logger.Printf("%s grid meter power: %.0fW", lp.Name, gridPower)

	// get charger current
	chargeCurrent, err := lp.Charger.ActualCurrent()
	if err != nil {
		Logger.Printf("%s charger error: %v", lp.Name, err)
		return err
	}
	Logger.Printf("%s charge current: %dA", lp.Name, chargeCurrent)

	// get charge power
	chargePower := CurrentToPower(float64(chargeCurrent), lp.Voltage, lp.Phases)
	Logger.Printf("%s charge power: %.0fW", lp.Name, chargePower)

	// -2500w = -1500w - 1000w
	haNetPower := gridPower - chargePower
	Logger.Printf("%s home power: %.0fW", lp.Name, haNetPower)

	// maxChargePower = 2500w
	maxChargePower := -haNetPower
	Logger.Printf("%s max charge power: %.0fW", lp.Name, maxChargePower)

	// get max charge current
	f := PowerToCurrent(maxChargePower, lp.Voltage, lp.Phases)
	targetChargeCurrent := int64(math.Max(0, f))
	Logger.Printf("%s max charge current: %dA", lp.Name, targetChargeCurrent)

	if targetChargeCurrent < lp.MinCurrent {
		switch lp.Mode {
		case api.ModeMinPV:
			targetChargeCurrent = lp.MinCurrent
			minPower := CurrentToPower(float64(targetChargeCurrent), lp.Voltage, lp.Phases)
			Logger.Printf("%s override charge power: %.0fW", lp.Name, minPower)
		case api.ModePV:
			targetChargeCurrent = 0
			Logger.Printf("%s override charge power: 0W", lp.Name)
		}
		Logger.Printf("%s max charge current: %dA", lp.Name, targetChargeCurrent)
	}

	// set max charge current
	if err := lp.setTargetCurrent(chargeCurrent, targetChargeCurrent); err != nil {
		return err
	}

	return nil
}
