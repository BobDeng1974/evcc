package exec

import (
	"time"

	"github.com/andig/ulm/api"
)

type charger struct {
	statusCmd  string
	currentCmd string
	enabledCmd string
	enableCmd  string
	timeout    time.Duration
}

// NewCharger creates a new exec charger
func NewCharger(statusCmd, currentCmd, enabledCmd, enableCmd string, timeout time.Duration) api.Charger {
	return &charger{
		statusCmd:  statusCmd,
		currentCmd: currentCmd,
		enabledCmd: enabledCmd,
		enableCmd:  enableCmd,
		timeout:    timeout,
	}
}

func (m *charger) Status() (api.ChargeStatus, error) {
	s, err := execWithStringResult(contextWithTimeout(m.timeout), m.statusCmd)
	if err != nil {
		return api.StatusNone, err
	}

	return api.ChargeStatus(s), nil
}

func (m *charger) ActualCurrent() (int64, error) {
	f, err := execWithFloatResult(contextWithTimeout(m.timeout), m.currentCmd)
	return int64(f), err
}

func (m *charger) Enabled() (bool, error) {
	s, err := execWithStringResult(contextWithTimeout(m.timeout), m.enabledCmd)
	if err != nil {
		return false, err
	}

	return truish(s), nil
}

func (m *charger) Enable(enable bool) error {
	cmd, err := replaceFormatted(m.enableCmd, map[string]interface{}{
		"enable": enable,
	})
	if err != nil {
		return err
	}

	_, err = execWithStringResult(contextWithTimeout(m.timeout), cmd)
	return err
}
