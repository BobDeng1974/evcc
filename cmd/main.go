package main

import (
	"log"
	"strconv"

	"github.com/andig/ulm"
	"github.com/kballard/go-shellquote"
)

type meter struct{}

func (m *meter) CurrentPower() (float64, error) {
	return 500, nil
}

type charger struct{}

func (c *charger) Enable(charge bool) error {
	return nil
}

func (c *charger) Status() (ulm.ChargeStatus, error) {
	return ulm.StatusA, nil
}

func (c *charger) MaxPower(max float64) error {
	return nil
}

func control(c ulm.Charger) {
	if c, ok := c.(ulm.ChargeController); ok {
		log.Println("Maxpower")
		c.MaxPower(1)
	} else {
		log.Println("no Maxpower")
	}
}

func main() {
	m := &meter{}
	log.Println(m.CurrentPower())

	c := &charger{}
	// c.MaxPower(5)
	c.Enable(true)

	voltage := 230
	phases := 3
	current := 16
	power := phases * current * voltage
	log.Println(power)

	log.Println(strconv.ParseFloat(string([]byte{}), 64))

	control(c)

	a, _ := shellquote.Split("/bin/sh -c \"foo bar\"")
	log.Printf("%d %+v", len(a), a)
}
