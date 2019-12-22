package exec

import (
	"testing"

	"github.com/andig/ulm/api"
)

func TestNewCharger(t *testing.T) {
	var m api.Charger = NewCharger("", "", "", "", 0)
	_ = m
}

func TestChargerSuccessStatus(t *testing.T) {
	m := &charger{
		statusCmd: "/bin/bash -c true",
	}
	if _, err := m.Status(); err != nil {
		t.Error(err)
	}
}
func TestChargerSuccessCurrent(t *testing.T) {
	m := &charger{
		currentCmd: "/bin/bash -c 'echo 1'",
	}
	if r, err := m.ActualCurrent(); r != 1 || err != nil {
		t.Error(err)
	}
}
func TestChargerSuccessEnable(t *testing.T) {
	m := &charger{
		enableCmd: "/bin/bash -c true",
	}
	if err := m.Enable(true); err != nil {
		t.Error(err)
	}
}
func TestChargerSuccessEnabled(t *testing.T) {
	m := &charger{
		enabledCmd: "/bin/bash -c 'echo true'",
	}
	if b, err := m.Enabled(); !b || err != nil {
		t.Error(err)
	}
}

func TestChargerFailStatus(t *testing.T) {
	m := &charger{
		statusCmd: "/bin/bash -c 'false'",
	}
	if _, err := m.Status(); err == nil {
		t.Error(err)
	}
}
func TestChargerFailCurrent(t *testing.T) {
	m := &charger{
		currentCmd: "/bin/bash -c false",
	}
	if _, err := m.ActualCurrent(); err == nil {
		t.Error(err)
	}
}
func TestChargerFailEnable(t *testing.T) {
	m := &charger{
		enableCmd: "/bin/bash -c false",
	}
	if err := m.Enable(true); err == nil {
		t.Error(err)
	}
}
func TestChargerFailEnabled(t *testing.T) {
	m := &charger{
		enabledCmd: "/bin/bash -c false",
	}
	if _, err := m.Enabled(); err == nil {
		t.Error(err)
	}
}
