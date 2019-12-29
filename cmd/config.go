package cmd

type Config struct {
	Mqtt       MqttConfig
	Meters     []MeterConfig
	Chargers   []ChargerConfig
	Loadpoints []LoadpointConfig
}

type MqttConfig struct {
	Broker   string
	User     string
	Password string
	Qos      int
}

type MeterConfig struct {
	Name   string
	Type   string
	Power  string
	Energy string
}

type ChargerConfig struct {
	Name string
	Type string
	URI  string
}
type LoadpointConfig struct {
	Name        string
	Charger     string
	GridMeter   string
	PVMeter     string
	ChargeMeter string
}
