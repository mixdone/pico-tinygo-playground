package main

import (
	"machine"
	"time"
)

var (
	spi      = machine.SPI0
	latchPin = machine.GP1
)

var spiConfig = machine.SPIConfig{
	Frequency: 10_000_000,
	Mode:      0,
	SCK:       machine.GP2,
	SDO:       machine.GP3,
}

var digitPins = []machine.Pin{
	machine.GP0,
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

type buttonState int

const (
	IDLE buttonState = iota
	DEBOUNCING
)

var (
	btnState   = [4]buttonState{IDLE, IDLE, IDLE, IDLE}
	btnPending = [4]bool{false, false, false, false}
	btnTimer   = [4]time.Time{}
)

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
	spi.Configure(spiConfig)

	latchPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	latchPin.High()

	for _, pin := range digitPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin.High()
	}

	for i, pin := range buttonPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
		idx := i
		pin.SetInterrupt(machine.PinRising, func(p machine.Pin) {
			btnPending[idx] = true
		})
	}

	for {
		handleButtons()
		refreshDisplay()
	}
}

func handleButtons() {
	now := time.Now()

	for i := 0; i < 4; i++ {
		if btnPending[i] {
			buttonPins[i].SetInterrupt(0, nil)
			btnState[i] = DEBOUNCING
			btnTimer[i] = now
			btnPending[i] = false
		}

		if btnState[i] == DEBOUNCING {
			if now.Sub(btnTimer[i]) >= 20*time.Millisecond {
				if buttonPins[i].Get() {
					digits[i]++
					if digits[i] > 9 {
						digits[i] = 0
					}
				}
				btnState[i] = IDLE

				idx := i
				buttonPins[i].SetInterrupt(machine.PinRising, func(p machine.Pin) {
					btnPending[idx] = true
				})
			}
		}
	}
}

func sendToRegister(b byte) {
	latchPin.Low()
	spi.Tx([]byte{b}, nil)
	latchPin.High()
}

func refreshDisplay() {
	for i := 0; i < 4; i++ {
		for _, pin := range digitPins {
			pin.High()
		}

		sendToRegister(numMap[digits[i]])
		digitPins[i].Low()
		time.Sleep(500 * time.Microsecond)
	}
}
