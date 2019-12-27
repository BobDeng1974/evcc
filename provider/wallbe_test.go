package provider

import (
	"testing"

	"github.com/andig/ulm/api"
)

func TestWallbe(t *testing.T) {
	var c api.Charger
	c = NewWallbe("192.168.0.8:502")

	if _, ok := c.(api.ChargeController); !ok {
		t.Error("not a charge controller")
	}
}
