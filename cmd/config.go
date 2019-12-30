package cmd

import "github.com/andig/evcc/api"

type Config struct {
	URI        string
	Mqtt       MqttConfig
	Meters     []MeterConfig
	Chargers   []ChargerConfig
	LoadPoints []LoadPointConfig
}

type MqttConfig struct {
	Broker   string
	User     string
	Password string
}

type MeterConfig struct {
	Name   string
	Type   string
	Power  string
	Energy string
}

type ProviderConfig struct {
	Type  string
	Topic string
	Cmd   string
}

type ChargerConfig struct {
	Name string
	Type string

	// wallbe charger
	URI string

	// composite charger
	Status        *ProviderConfig // Charger
	ActualCurrent *ProviderConfig // Charger
	MaxCurrent    *ProviderConfig // ChargeController
	Enable        *ProviderConfig // Charger
	Enabled       *ProviderConfig // Charger
}

type LoadPointConfig struct {
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
