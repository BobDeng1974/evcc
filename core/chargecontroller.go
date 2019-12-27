package core

import (
	"context"

	"github.com/andig/ulm/api"
)

type ChargeController struct {
	maxCurrentS api.IntSetter
}

// NewChargeController creates a new charge controller
func NewChargeController(maxCurrentS api.IntSetter) api.ChargeController {
	return &ChargeController{
		maxCurrentS: maxCurrentS,
	}
}

func (m *ChargeController) MaxCurrent(current int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return m.maxCurrentS(ctx, current)
}
