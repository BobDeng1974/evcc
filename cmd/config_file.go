package cmd

import "github.com/andig/evcc/api"

type config struct {
	URI        string
	Mqtt       mqttConfig
	Meters     []meterConfig
	Chargers   []chargerConfig
	LoadPoints []loadPointConfig
}

type mqttConfig struct {
	Broker   string
	User     string
	Password string
}

type meterConfig struct {
	Name   string
	Type   string
	Power  *providerConfig
	Energy *providerConfig
}

type providerConfig struct {
	Type  string
	Topic string
	Cmd   string
}

type chargerConfig struct {
	Name string
	Type string

	// wallbe charger
	URI string

	// composite charger
	Status        *providerConfig // Charger
	ActualCurrent *providerConfig // Charger
	MaxCurrent    *providerConfig // ChargeController
	Enable        *providerConfig // Charger
	Enabled       *providerConfig // Charger
}

type loadPointConfig struct {
	Name        string
	Charger     string // api.Charger
	GridMeter   string // api.Meter
	PVMeter     string // api.Meter
	ChargeMeter string // api.Meter
	Mode        api.ChargeMode
	MinCurrent  int64
	MaxCurrent  int64
	Voltage     float64
	Phases      float64
}
