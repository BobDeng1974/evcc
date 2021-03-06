package core

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/andig/evcc/api"
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
	sync.Mutex
	Name        string
	Mode        api.ChargeMode
	Charger     api.Charger
	GridMeter   api.Meter // home usage meter
	PVMeter     api.Meter // pv generation meter
	ChargeMeter api.Meter // charger usage meter
	MinCurrent  int64     // PV mode: start current	Min+PV mode: min current
	MaxCurrent  int64
	Voltage     float64
	Phases      float64

	// state variables
	isCharging        bool
	chargeStartEnergy float64
	chargeStartTime   time.Time
	chargedEnergy     float64
	chargedDuration   time.Duration
}

// NewLoadPoint creates a LoadPoint with sane defaults
func NewLoadPoint(name string, charger api.Charger) *LoadPoint {
	return &LoadPoint{
		Name:            name,
		Phases:          1,
		Voltage:         230, // V
		MinCurrent:      5,   // A
		MaxCurrent:      16,  // A
		Mode:            api.ModeNow,
		Charger:         charger,
		chargedDuration: 0,
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

// CurrentChargeMode returns current charge mode
func (lp *LoadPoint) CurrentChargeMode() api.ChargeMode {
	lp.Lock()
	defer lp.Unlock()

	return lp.Mode
}

// ChargeMode updates charge mode
func (lp *LoadPoint) ChargeMode(mode api.ChargeMode) error {
	Logger.Printf("%s set charge mode: %s", lp.Name, string(mode))

	// check if charger is controllable
	_, chargerControllable := lp.Charger.(api.ChargeController)

	// disable charger if enabled
	if mode == api.ModeOff {
		if err := lp.chargerEnable(false); err != nil {
			return err
		}

		lp.Lock()
		defer lp.Unlock()
		lp.Mode = mode

		// async from http call
		go lp.stopCharging()

		return nil
	}

	// enable charger if disabled
	if err := lp.chargerEnable(true); err != nil {
		return err
	}

	// remaining modes require GridMeter
	if mode == api.ModeMinPV || mode == api.ModePV {
		if lp.GridMeter == nil || !chargerControllable {
			return errors.New("invalid charge mode: " + string(mode))
		}
	}

	lp.Lock()
	defer lp.Unlock()
	lp.Mode = mode

	// async from http call
	go lp.startCharging()

	return nil
}

// startCharging resets charge energy counter and starts charge timer
func (lp *LoadPoint) startCharging() {
	lp.Lock()
	if lp.isCharging {
		lp.Unlock()
		return
	}
	lp.isCharging = true
	lp.chargeStartTime = time.Now()
	lp.chargeStartEnergy = 0
	lp.Unlock()

	// get starting energy amount
	if m, ok := lp.ChargeMeter.(api.MeterEnergy); ok {
		if f, err := m.TotalEnergy(); err == nil {
			lp.Lock()
			defer lp.Unlock()
			lp.chargeStartEnergy = f
			Logger.Printf("%s charge start energy: %.0f", lp.Name, lp.chargeStartEnergy)
		} else {
			Logger.Printf("%s charge meter error: %s", lp.Name, err)
		}
	}
}

func (lp *LoadPoint) stopCharging() {
	lp.Lock()
	if !lp.isCharging {
		lp.Unlock()
		return
	}
	lp.isCharging = false
	lp.chargedDuration = time.Now().Sub(lp.chargeStartTime)
	lp.Unlock()

	// get end energy amount
	if m, ok := lp.ChargeMeter.(api.MeterEnergy); ok {
		if f, err := m.TotalEnergy(); err == nil {
			lp.Lock()
			defer lp.Unlock()
			lp.chargedEnergy = f - lp.chargeStartEnergy
			Logger.Printf("%s charged energy final: %.0f", lp.Name, lp.chargedEnergy)
		} else {
			Logger.Printf("%s charge meter error: %s", lp.Name, err)
		}
	}
}

// ChargeDuration returns for how long the charge cycle has been running
func (lp *LoadPoint) ChargeDuration() time.Duration {
	lp.Lock()
	defer lp.Unlock()

	// return previous if not currently charging
	if !lp.isCharging {
		return lp.chargedDuration
	}

	return time.Now().Sub(lp.chargeStartTime)
}

// ChargedEnergy returns energy consumption since charge start
func (lp *LoadPoint) ChargedEnergy() (float64, error) {
	if lp.ChargeMeter == nil {
		return 0, fmt.Errorf("%s no charge meter assigned", lp.Name)
	}

	// return previous if not currently charging
	lp.Lock()
	if !lp.isCharging {
		defer lp.Unlock()
		return lp.chargedEnergy, nil
	}
	lp.Unlock()

	// get starting energy amount
	if m, ok := lp.ChargeMeter.(api.MeterEnergy); ok {
		if f, err := m.TotalEnergy(); err == nil {
			lp.Lock()
			defer lp.Unlock()
			chargedEnergy := f - lp.chargeStartEnergy
			Logger.Printf("%s charged energy sofar: %.0f", lp.Name, chargedEnergy)

			return chargedEnergy, nil
		} else {
			Logger.Printf("%s charge meter error: %s", lp.Name, err)
		}
	}

	return 0, fmt.Errorf("%s charge meter does not support measuring energy", lp.Name)
}

// updateChargerEnabled checks charger enabled state
func (lp *LoadPoint) updateChargerEnabled() (bool, api.ChargeMode) {
	// check charger status
	enabled, err := lp.Charger.Enabled()
	if err != nil {
		log.Printf("%s charger error: %v", lp.Name, err)
		return false, api.ModeOff
	}
	Logger.Printf("%s charger enabled: %v", lp.Name, enabled)

	// set mode=off if charger not enabled
	if !enabled {
		lp.stopCharging()

		lp.Lock()
		lp.Mode = api.ModeOff
		lp.Unlock()
	}

	lp.Lock()
	defer lp.Unlock()

	return enabled, lp.Mode
}

func (lp *LoadPoint) updateCarConnected() bool {
	// abort if no vehicle connected
	status, err := lp.Charger.Status()
	if err != nil {
		log.Printf("%s charger error: %v", lp.Name, err)
		return false
	}
	Logger.Printf("%s charger status: %s", lp.Name, status)

	connected := status == api.StatusC
	if !connected {
		lp.stopCharging()
	}

	return connected
}

// Update reevaluates meters and charger state
func (lp *LoadPoint) Update() {
	// check if charging is enabled
	enabled, mode := lp.updateChargerEnabled()
	Logger.Printf("%s charge mode: %s", lp.Name, mode)
	if !enabled || mode == api.ModeOff {
		return
	}

	// check if car is connected
	if connected := lp.updateCarConnected(); !connected {
		return
	}

	// start tracking time and energy
	lp.startCharging()

	// abort if dumb charge controller
	if _, chargeController := lp.Charger.(api.ChargeController); !chargeController {
		log.Printf("%s no charge controller assigned", lp.Name)
		return
	}

	// execute loading strategy
	var err error
	switch mode {
	case api.ModeNow:
		err = lp.ApplyModeNow()
	case api.ModeMinPV, api.ModePV:
		err = lp.ApplyModePV(mode)
	}

	if err != nil {
		Logger.Printf("%s error: %v", lp.Name, err)
		return
	}
}

// ApplyModeNow sets "now" charger mode
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

// ApplyModePV sets "minpv" or "pv" load modes
func (lp *LoadPoint) ApplyModePV(mode api.ChargeMode) error {
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
		switch mode {
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
