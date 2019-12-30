package core

import (
	"context"

	"github.com/andig/evcc/api"
)

type MeterEnergy struct {
	totalEnergyP api.FloatProvider
}

// NewMeterEnergy creates a new charger
func NewMeterEnergy(totalEnergyP api.FloatProvider) api.MeterEnergy {
	return &MeterEnergy{
		totalEnergyP: totalEnergyP,
	}
}

func (m *MeterEnergy) TotalEnergy() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return m.totalEnergyP(ctx)
}
