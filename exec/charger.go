package exec

import (
	"time"

	"github.com/andig/ulm/api"
)

type charger struct {
	status  string
	current string
	enable  string
	timeout time.Duration
}

// NewCharger creates a new exec charger
func NewCharger(status, current, enable string, timeout time.Duration) api.Charger {
	return &charger{
		status:  status,
		current: current,
		enable:  enable,
		timeout: timeout,
	}
}

func (m *charger) Status() (api.ChargeStatus, error) {
	s, err := execWithStringResult(contextWithTimeout(m.timeout), m.status)
	if err != nil {
		return api.StatusNone, err
	}

	return api.ChargeStatus(s), nil
}

func (m *charger) ActualCurrent() (float64, error) {
	return execWithFloatResult(contextWithTimeout(m.timeout), m.current)
}

func (m *charger) Enabled() (bool, error) {
	s, err := execWithStringResult(contextWithTimeout(m.timeout), m.status)
	if err != nil {
		return false, err
	}

	return truish(s), nil
}

func (m *charger) Enable(enable bool) error {
	cmd, err := replaceFormatted(m.status, map[string]interface{}{
		"enable": enable,
	})
	if err != nil {
		return err
	}

	_, err = execWithStringResult(contextWithTimeout(m.timeout), cmd)
	return err
}
