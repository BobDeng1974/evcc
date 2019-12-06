package exec

import (
	"time"

	"github.com/andig/ulm"
)

type charger struct {
	status  string
	enable  string
	timeout time.Duration
}

// NewCharger creates a new exec charger
func NewCharger(status string, enable string, timeout time.Duration) ulm.Charger {
	return &charger{
		status:  status,
		enable:  enable,
		timeout: timeout,
	}
}

func (m *charger) Status() (ulm.ChargeStatus, error) {
	s, err := execWithStringResult(contextWithTimeout(m.timeout), m.status)
	if err != nil {
		return ulm.StatusNone, err
	}

	return ulm.ChargeStatus(s), nil
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
