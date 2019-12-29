package provider

import (
	"encoding/binary"
	"time"

	"github.com/andig/evcc/api"
	"github.com/grid-x/modbus"
)

const (
	slaveID = 255
	timeout = 1 * time.Second
)

type Wallbe struct {
	client modbus.Client
}

// NewWallbe creates a Wallbe charger
func NewWallbe(conn string) api.Charger {
	handler := modbus.NewTCPClientHandler(conn)
	client := modbus.NewClient(handler)

	handler.SlaveID = slaveID
	handler.Timeout = timeout

	return &Wallbe{
		client: client,
	}
}

func (m *Wallbe) Status() (api.ChargeStatus, error) {
	b, err := m.client.ReadInputRegisters(100, 1)
	if err != nil {
		return api.StatusNone, nil
	}

	return api.ChargeStatus(string(b)), nil
}

func (m *Wallbe) ActualCurrent() (int64, error) {
	b, err := m.client.ReadHoldingRegisters(300, 1)
	if err != nil {
		return 0, nil
	}

	u := binary.BigEndian.Uint16(b)
	return int64(u), nil
}

func (m *Wallbe) MaxCurrent(current int64) error {
	_, err := m.client.WriteSingleRegister(528, uint16(current))
	return err
}

func (m *Wallbe) Enabled() (bool, error) {
	b, err := m.client.ReadCoils(400, 1)
	if err != nil {
		return false, nil
	}

	u := binary.BigEndian.Uint16(b)
	return u == 1, nil
}

func (m *Wallbe) Enable(enable bool) error {
	var u uint16
	if enable {
		u = 1
	}
	_, err := m.client.WriteSingleCoil(400, u)
	return err
}
