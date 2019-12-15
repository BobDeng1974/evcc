package main

import (
	"github.com/andig/ulm/api"
)

type meter struct{}

func (m *meter) CurrentPower() (float64, error) {
	return 500, nil
}

type charger struct{}

func (c *charger) Enable(charge bool) error {
	return nil
}

func (c *charger) Status() (api.ChargeStatus, error) {
	return api.StatusC, nil
}

func (c *charger) Enabled() (bool, error) {
	return true, nil
}

func (c *charger) MaxPower(max float64) error {
	return nil
}

func (c *charger) ActualCurrent() (float64, error) {
	return 0, nil
}
