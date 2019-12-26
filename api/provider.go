package api

import "context"

type FloatProvider func(context.Context) (float64, error)

type IntProvider func(context.Context) (int64, error)

type StringProvider func(context.Context) (string, error)
