package api

import "context"

// Setters update typed device data.
// They are used to abstract the underlying device implementation.

type (
	FloatSetter  func(context.Context, float64) error
	IntSetter    func(context.Context, int64) error
	StringSetter func(context.Context, string) error
	BoolSetter   func(context.Context, bool) error
)
