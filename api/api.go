package api

//go:generate mockgen -destination mock_api/api.go github.com/andig/evcc/api Charger,ChargeController,Meter

// Meter is able to provide current power at metering point
type Meter interface {
	CurrentPower() (float64, error)
}

// MeterEnergy is able to provide current power at metering point
type MeterEnergy interface {
	TotalEnergy() (float64, error)
}

// ChargeStatus is the EVSE models charging status from A to F
type ChargeStatus string

const (
	StatusNone ChargeStatus = ""
	StatusA    ChargeStatus = "A" // Fzg. angeschlossen: nein    Laden möglich: nein
	StatusB    ChargeStatus = "B" // Fzg. angeschlossen:   ja    Laden möglich: nein
	StatusC    ChargeStatus = "C" // Fzg. angeschlossen:   ja    Laden möglich:   ja
	StatusD    ChargeStatus = "D" // Fzg. angeschlossen:   ja    Laden möglich:   ja
	StatusE    ChargeStatus = "E" // Fzg. angeschlossen:   ja    Laden möglich: nein
	StatusF    ChargeStatus = "F" // Fzg. angeschlossen:   ja    Laden möglich: nein
)

// Charger is able to provide current charging status and to enable/disabler charging
type Charger interface {
	Status() (ChargeStatus, error)
	Enabled() (bool, error)
	Enable(enable bool) error
	ActualCurrent() (int64, error)
}

// ChargeController provides controlling of the charger's max allowed power
type ChargeController interface {
	MaxCurrent(current int64) error
}

// ChargeMode are charge modes modeled after OpenWB
type ChargeMode string

const (
	ModeOff   ChargeMode = "off"
	ModeNow   ChargeMode = "now"
	ModeMinPV ChargeMode = "minpv"
	ModePV    ChargeMode = "pv"
)

// LoadPoint ties charger and meter together and contains the controller logic
type LoadPoint interface {
	Update()
	CurrentChargeMode() ChargeMode
	ChargeMode(mode ChargeMode) error
}
