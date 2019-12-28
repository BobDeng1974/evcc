package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/andig/ulm/api"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	publishTimeout = 2 * time.Second
	valueTimeout   = 10 * time.Second
)

// MqttClient is a paho publisher
type MqttClient struct {
	Client  mqtt.Client
	qos     byte
	verbose bool
}

// NewMqttClient creates new publisher for paho
func NewMqttClient(
	broker string,
	user string,
	password string,
	clientID string,
	cleanSession bool,
	qos byte,
) *MqttClient {
	log.Printf("mqtt: connecting %s at %s", clientID, broker)

	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetUsername(user)
	options.SetPassword(password)
	options.SetClientID(clientID)
	options.SetCleanSession(cleanSession)
	options.SetAutoReconnect(true)

	client := mqtt.NewClient(options)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt: error connecting: %s", token.Error())
	}
	log.Println("mqtt: connected")

	return &MqttClient{
		Client: client,
		qos:    qos,
	}
}

// Listen listens to topic and relays to publisher
func (m *MqttClient) Listen(topic string, callback func(string)) {
	token := m.Client.Subscribe(topic, m.qos, func(c mqtt.Client, msg mqtt.Message) {
		s := string(msg.Payload())
		if len(s) > 0 {
			callback(s)
		}
	})
	m.WaitForToken(token)
}

// FloatValue parses float64 from MQTT topic and returns cached value
func (m *MqttClient) FloatValue(topic string) api.FloatProvider {
	var mux sync.Mutex // guards following values
	var ts time.Time
	var val float64
	var err error

	// listen
	m.Listen(topic, func(s string) {
		mux.Lock()
		defer mux.Unlock()

		if val, err = strconv.ParseFloat(s, 64); err == nil {
			// log.Printf("mqtt: recv %s value '%.2f'", topic, val)
			ts = time.Now()
		} else {
			log.Printf("mqtt: invalid value '%s'", s)
		}
	})

	// return func to access cached value
	return func(ctx context.Context) (float64, error) {
		mux.Lock()
		defer mux.Unlock()

		// cached value unless outdated
		if ts.Add(valueTimeout).After(time.Now()) {
			return val, err
		}

		// value outdated
		return val, fmt.Errorf("mqtt: value outdated for %s", topic)
	}
}

// IntValue parses int64 from MQTT topic and returns cached value
func (m *MqttClient) IntValue(topic string) api.IntProvider {
	var mux sync.Mutex // guards following values
	var ts time.Time
	var val int64
	var err error

	// listen
	m.Listen(topic, func(s string) {
		mux.Lock()
		defer mux.Unlock()

		if val, err = strconv.ParseInt(s, 10, 64); err == nil {
			ts = time.Now()
		} else {
			log.Printf("mqtt: invalid value '%s'", s)
		}
	})

	// return func to access cached value
	return func(ctx context.Context) (int64, error) {
		mux.Lock()
		defer mux.Unlock()

		// cached value unless outdated
		if ts.Add(valueTimeout).After(time.Now()) {
			return val, err
		}

		// value outdated
		return val, fmt.Errorf("mqtt: value outdated %s", topic)
	}
}

// WaitForToken synchronously waits until token operation completed
func (m *MqttClient) WaitForToken(token mqtt.Token) {
	if token.WaitTimeout(publishTimeout) {
		if token.Error() != nil {
			log.Printf("mqtt: error: %s", token.Error())
		}
	} else if m.verbose {
		log.Println("mqtt: timeout")
	}
}
