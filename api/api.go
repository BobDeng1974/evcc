package api

//go:generate mockgen -destination mock_api/api.go github.com/andig/ulm/api Charger,ChargeController,Meter

// Strategy calculates desired power based on input parameters
type Strategy interface {
	DesiredPower() (float64, error)
}

// Meter is able to provide current power at metering point
type Meter interface {
	CurrentPower() (float64, error)
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
	ActualCurrent() (int, error)
}

// ChargeController provides controlling of the charger's max allowed power
type ChargeController interface {
	MaxCurrent(current int) error
}

type ChargeMode string

const (
	ModeNow   ChargeMode = "now"
	ModeMinPV ChargeMode = "minpv"
	ModePV    ChargeMode = "pv"
)

type LoadPoint interface {
	Update()
	CurrentChargeMode() ChargeMode
	ChargeMode(mode ChargeMode) error
}
