package server

import (
	"context"
	"log"

	"github.com/andig/ulm/api"
)

// Observer allows to intercept provider functions and re-use their values
type Observer struct {
	values   chan<- SocketValue
	observed []interface{}
}

// NewObserver creates new observer
func NewObserver(c chan<- SocketValue) *Observer {
	return &Observer{
		values:   c,
		observed: make([]interface{}, 0),
	}
}

// Observe wraps a provider function, allowing to intercept the returned value
// and relaying it to the observer's channel. The wrapped function is added to the
// observer's list of observed functions.
func (m *Observer) Observe(key string, fun interface{}) {
	switch typed := fun.(type) {
	case api.FloatProvider:
		m.observed = append(m.observed, m.FloatValue(key, typed))
	// case api.IntProvider:
	// 	m.observed = append(m.observed, m.IntValue(key, typed))
	// case api.StringProvider:
	// 	m.observed = append(m.observed, m.StringValue(key, typed))
	default:
		panic("observer: invalid type")
	}
}

// Update updates all observed values
func (m *Observer) Update(ctx context.Context) {
	for _, fun := range m.observed {
		switch typed := fun.(type) {
		case api.FloatProvider:
			if _, err := typed(ctx); err != nil {
				log.Printf("observer: %v", err)
			}
		case api.IntProvider:
			if _, err := typed(ctx); err != nil {
				log.Printf("observer: %v", err)
			}
		case api.StringProvider:
			if _, err := typed(ctx); err != nil {
				log.Printf("observer: %v", err)
			}
		}
	}
}

// FloatValue returns a wrapped api provider
func (m *Observer) FloatValue(key string, fun api.FloatProvider) api.FloatProvider {
	return func(ctx context.Context) (float64, error) {
		val, err := fun(ctx)

		// send to socket channel if not error
		if err == nil {
			m.values <- SocketValue{
				Key: key,
				Val: val,
			}
		}

		return val, err
	}
}
