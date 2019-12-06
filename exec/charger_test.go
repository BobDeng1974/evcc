package exec

import (
	"testing"

	"github.com/andig/ulm"
)

func TestNewCharger(t *testing.T) {
	var m ulm.Charger = NewCharger("", "", 0)
	_ = m
}

func TestChargerFail(t *testing.T) {
	m := &charger{
		status: "/bin/bash -c false",
		enable: "/bin/bash -c false",
	}

	if _, err := m.Status(); err == nil {
		t.Error(err)
	}
	if err := m.Enable(true); err == nil {
		t.Error(err)
	}
}
