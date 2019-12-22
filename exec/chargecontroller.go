package exec

import (
	"time"

	"github.com/andig/ulm/api"
)

type chargecontroller struct {
	maxCurrentCmd string
	timeout       time.Duration
}

// NewChargeController creates a new exec chargecontroller
func NewChargeController(maxCurrent string, timeout time.Duration) api.ChargeController {
	return &chargecontroller{
		maxCurrentCmd: maxCurrent,
		timeout:       timeout,
	}
}

func (m *chargecontroller) MaxCurrent(current int) error {
	cmd, err := replaceFormatted(m.maxCurrentCmd, map[string]interface{}{
		"current": current,
	})
	if err != nil {
		return err
	}

	_, err = execWithStringResult(contextWithTimeout(m.timeout), cmd)
	return err
}
