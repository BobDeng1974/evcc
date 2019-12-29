package core

import (
	"context"
	"time"

	"github.com/andig/evcc/api"
)

const (
	timeout = 1 * time.Second
)

type Charger struct {
	statusP        api.StringProvider
	actualCurrentP api.IntProvider
	enabledP       api.BoolProvider
	enableS        api.BoolSetter
}

// NewCharger creates a new charger
func NewCharger(
	statusP api.StringProvider,
	actualCurrentP api.IntProvider,
	enabledP api.BoolProvider,
	enableS api.BoolSetter,
) api.Charger {
	return &Charger{
		statusP:        statusP,
		actualCurrentP: actualCurrentP,
		enabledP:       enabledP,
		enableS:        enableS,
	}
}

func (m *Charger) Status() (api.ChargeStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	s, err := m.statusP(ctx)
	if err != nil {
		return api.StatusNone, err
	}

	return api.ChargeStatus(s), nil
}

func (m *Charger) ActualCurrent() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return m.actualCurrentP(ctx)
}

func (m *Charger) Enabled() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return m.enabledP(ctx)
}

func (m *Charger) Enable(enable bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return m.enableS(ctx, enable)
}
