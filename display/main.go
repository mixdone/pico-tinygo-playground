package main

import (
	"machine"
	"time"
)

var (
	dataPin  = machine.GP0
	latchPin = machine.GP1
	clockPin = machine.GP2
)

var digitPins = []machine.Pin{
	machine.GP3,
	machine.GP4,
	machine.GP5,
	machine.GP6,
}

var buttonPins = []machine.Pin{
	machine.GP7,
	machine.GP8,
	machine.GP9,
	machine.GP10,
}

var digits = []int{0, 0, 0, 0}

var lastButtonStates = []bool{false, false, false, false}
var lastChangeTime = []time.Time{
	time.Now(), time.Now(), time.Now(), time.Now(),
}
var handled = []bool{false, false, false, false}

var numMap = []byte{
	0b00111111,
	0b00000110,
	0b01011011,
	0b01001111,
	0b01100110,
	0b01101101,
	0b01111101,
	0b00000111,
	0b01111111,
	0b01101111,
}

func main() {
	dataPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	latchPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	clockPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	for _, pin := range digitPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin.High()
	}

	for i, pin := range buttonPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
		lastButtonStates[i] = pin.Get()
	}

	for {
		for i := 0; i < 4; i++ {
			state := buttonPins[i].Get()

			if state != lastButtonStates[i] {
				lastChangeTime[i] = time.Now()
				lastButtonStates[i] = state
				handled[i] = false
				continue
			}

			if state && !handled[i] &&
				time.Since(lastChangeTime[i]) >= 20*time.Millisecond {
				digits[i]++
				if digits[i] > 9 {
					digits[i] = 0
				}
				handled[i] = true
			}
		}
		refreshDisplay()
	}
}

func sendToRegister(b byte) {
	latchPin.Low()
	for i := 0; i < 8; i++ {
		bit := (b >> (7 - uint(i))) & 1
		if bit == 1 {
			dataPin.High()
		} else {
			dataPin.Low()
		}
		clockPin.High()
		clockPin.Low()
	}
	latchPin.High()
}

func refreshDisplay() {
	for i := 0; i < 4; i++ {
		for _, pin := range digitPins {
			pin.High()
		}

		sendToRegister(0x00)
		sendToRegister(numMap[digits[i]])
		digitPins[i].Low()
		time.Sleep(500 * time.Microsecond)
	}
}
