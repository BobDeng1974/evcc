package core

func CurrentToPower(current, voltage, phases float64) float64 {
	return phases * current * voltage
}

func PowerToCurrent(power, voltage, phases float64) float64 {
	return power / (phases * voltage)
}
