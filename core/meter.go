package core

import (
	"context"

	"github.com/andig/evcc/api"
)

type Meter struct {
	currentPowerP api.FloatProvider
}

// NewMeter creates a new charger
func NewMeter(currentPowerP api.FloatProvider) api.Meter {
	return &Meter{
		currentPowerP: currentPowerP,
	}
}

func (m *Meter) CurrentPower() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return m.currentPowerP(ctx)
}
