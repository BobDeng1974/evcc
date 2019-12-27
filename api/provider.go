package api

import "context"

// Providers return typed device data.
// They are used to abstract the underlying device implementation.

type (
	FloatProvider  func(context.Context) (float64, error)
	IntProvider    func(context.Context) (int64, error)
	StringProvider func(context.Context) (string, error)
	BoolProvider   func(context.Context) (bool, error)
)
