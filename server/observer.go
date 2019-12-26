package server

import (
	"context"
	"log"

	"github.com/andig/ulm/api"
)

type Observer struct {
	values   chan<- SocketValue
	observed []interface{}
}

func NewObserver(c chan<- SocketValue) *Observer {
	return &Observer{
		values:   c,
		observed: make([]interface{}, 0),
	}
}

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
