package exec

import (
	"time"

	"github.com/andig/ulm/api"
)

type meter struct {
	script  string
	timeout time.Duration
}

// NewMeter creates a new exec meter
func NewMeter(script string, timeout time.Duration) api.Meter {
	return &meter{
		script:  script,
		timeout: timeout,
	}
}

func (m *meter) CurrentPower() (float64, error) {
	return execWithFloatResult(contextWithTimeout(m.timeout), m.script)
}
