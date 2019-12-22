package exec

import (
	"testing"

	"github.com/andig/ulm/api"
)

func TestNewChargeController(t *testing.T) {
	var m api.ChargeController = NewChargeController("", 0)
	_ = m
}

func TestChargeControllerFail(t *testing.T) {
	m := &chargecontroller{
		maxCurrentCmd: "/bin/bash -c false",
	}

	if err := m.MaxCurrent(1); err == nil {
		t.Error(err)
	}
}
