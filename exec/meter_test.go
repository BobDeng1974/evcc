package exec

import (
	"testing"

	"github.com/andig/ulm/api"
)

func TestNewMeter(t *testing.T) {
	var m api.Meter = NewMeter("", 0)
	_ = m
}

func TestOutErr(t *testing.T) {
	m := &meter{
		script: "/bin/bash -c \"echo -n 1; echo 1>&2 2\"",
	}

	if p, err := m.CurrentPower(); p != 12 || err != nil {
		t.Error(p, err)
	}
}
func TestMeterFail(t *testing.T) {
	m := &meter{
		script: "/bin/bash -c false",
	}

	if _, err := m.CurrentPower(); err == nil {
		t.Error(err)
	}
}
