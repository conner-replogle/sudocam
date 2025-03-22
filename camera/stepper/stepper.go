package stepper

import (
	"fmt"

	"os"
	"time"
)

type Stepper struct {
	pins       [4]int16
	valueFiles []*os.File // Store file pointers for each pin
	steps      int
	direction  bool
}

// Define the sequence for a full step drive
var sequence = [][4]int{
	{0, 0, 0, 1},
	{0, 0, 1, 1},
	{0, 0, 1, 0},
	{0, 1, 1, 0},
	{0, 1, 0, 0},
	{1, 1, 0, 0},
	{1, 0, 0, 0},
	{1, 0, 0, 1},
}

func NewStepper(pins [4]int16) *Stepper {
	return &Stepper{
		pins:       pins,
		valueFiles: make([]*os.File, 4), // Initialize the slice
		steps:      0,
		direction:  true,
	}
}

func (s *Stepper) Initialize() {
	exportFile, err := os.OpenFile("/sys/class/gpio/export", os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer exportFile.Close()

	for i, pin := range s.pins {
		_, err = exportFile.WriteString(fmt.Sprint(pin))
		if err != nil {
			fmt.Printf("Error exporting pin %d: %v (may already be exported)\n", pin, err)
		}

		directionFile, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", pin), os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		_, err = directionFile.WriteString("out")
		if err != nil {
			panic(err)
		}
		directionFile.Close()

		// Open and store the value file
		valueFile, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", pin), os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		s.valueFiles[i] = valueFile
	}
}

// SetPinValue sets the GPIO pin value (0 or 1)
func (s *Stepper) SetPinValue(pin int16, value int) error {
	// Find the index of the pin
	index := -1
	for i, p := range s.pins {
		if p == pin {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("pin %d not found in stepper pins", pin)
	}

	// Use the stored file pointer
	_, err := s.valueFiles[index].WriteString(fmt.Sprint(value))
	if err != nil {
		return err
	}

	s.valueFiles[index].Sync()
	return nil
}

func (s *Stepper) step() {
	step := sequence[s.steps]
	for i, pin := range s.pins {
		s.SetPinValue(pin, step[i])
	}
	s.setDirection()
}

func (s *Stepper) setDirection() {
	if s.direction {
		s.steps++
	} else {
		s.steps--
	}

	if s.steps > 7 {
		s.steps = 0
	}
	if s.steps < 0 {
		s.steps = 7
	}
}

// StepForward moves the stepper motor one step forward
func (s *Stepper) Step(xw int,delay time.Duration) {
	if xw < 0 {
		s.direction = false
		xw = -xw
	} else {
		s.direction = true
	}


	for range xw {
		s.step()
		time.Sleep(delay)
		

	}
}
