package stepper

import "time"

var tiltPins = [4]int16{41, 40, 33, 32}

type MovementManager struct {
	tiltStepper *Stepper
}

func NewMovementManager() *MovementManager {
	tiltStepper := NewStepper(tiltPins)
	tiltStepper.Initialize()
	return &MovementManager{
		tiltStepper,
	}
}

func (m *MovementManager) MoveTilt(steps int) {
	m.tiltStepper.Step(steps, 800*time.Microsecond)
}

