package main

import (
	"github.com/andig/ulm/api"
	"math"
	"math/rand"
)

type meter struct{}

func (m *meter) CurrentPower() (float64, error) {
	f := math.Round(30*rand.Float64()-25) * 100
	return f, nil
}

type charger struct {
	current int
}

func (c *charger) Enable(charge bool) error {
	return nil
}

func (c *charger) Status() (api.ChargeStatus, error) {
	return api.StatusC, nil
}

func (c *charger) Enabled() (bool, error) {
	return true, nil
}

func (c *charger) MaxCurrent(max int) error {
	c.current = max
	return nil
}

func (c *charger) ActualCurrent() (int, error) {
	// i := rand.Int31n(17)
	return c.current, nil
}
