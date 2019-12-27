package api

import "context"

// Providers return typed device data.
// They are used to abstract the underlying device implementation.

type FloatProvider func(context.Context) (float64, error)
type IntProvider func(context.Context) (int64, error)
type StringProvider func(context.Context) (string, error)
type BoolProvider func(context.Context) (bool, error)
