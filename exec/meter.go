package exec

import (
	"time"

	"github.com/andig/ulm/api"
)

type meter struct {
	currentPowerCmd string
	timeout         time.Duration
}

// NewMeter creates a new exec meter
func NewMeter(currentPowerCmd string, timeout time.Duration) api.Meter {
	return &meter{
		currentPowerCmd: currentPowerCmd,
		timeout:         timeout,
	}
}

func (m *meter) CurrentPower() (float64, error) {
	return execWithFloatResult(contextWithTimeout(m.timeout), m.currentPowerCmd)
}
