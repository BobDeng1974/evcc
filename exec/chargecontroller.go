package exec

import (
	"time"

	"github.com/andig/ulm/api"
)

type chargecontroller struct {
	cmd     string
	timeout time.Duration
}

// NewChargeController creates a new exec chargecontroller
func NewChargeController(cmd string, timeout time.Duration) api.ChargeController {
	return &chargecontroller{
		cmd:     cmd,
		timeout: timeout,
	}
}

func (m *chargecontroller) MaxCurrent(current int) error {
	cmd, err := replaceFormatted(m.cmd, map[string]interface{}{
		"current": current,
	})
	if err != nil {
		return err
	}

	_, err = execWithStringResult(contextWithTimeout(m.timeout), cmd)
	return err
}
