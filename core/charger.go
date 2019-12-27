package core

import (
	"context"
	"time"

	"github.com/andig/ulm/api"
)

const (
	timeout = 1 * time.Second
)

type charger struct {
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
	return &charger{
		statusP:        statusP,
		actualCurrentP: actualCurrentP,
		enabledP:       enabledP,
		enableS:        enableS,
	}
}

func (m *charger) Status() (api.ChargeStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	s, err := m.statusP(ctx)
	if err != nil {
		return api.StatusNone, err
	}

	return api.ChargeStatus(s), nil
}

func (m *charger) ActualCurrent() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return m.actualCurrentP(ctx)
}

func (m *charger) Enabled() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return m.enabledP(ctx)
}

func (m *charger) Enable(enable bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return m.enableS(ctx, enable)
}
