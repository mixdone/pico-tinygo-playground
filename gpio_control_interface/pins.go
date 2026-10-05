package main

import "machine"

type pinMode uint8

const (
	MODE_INPUT pinMode = iota
	MODE_OUTPUT
	MODE_PWM
)

type pinState struct {
	mode  pinMode
	value uint8
}

const totalPins = 21

var pinNumbers = [totalPins]machine.Pin{
	machine.GP0, machine.GP1, machine.GP2, machine.GP3,
	machine.GP6, machine.GP7, machine.GP8, machine.GP9,
	machine.GP10, machine.GP11, machine.GP12, machine.GP13,
	machine.GP14, machine.GP15, machine.GP16, machine.GP17,
	machine.GP21, machine.GP22,
	machine.GP26, machine.GP27, machine.GP28,
}

var pinNums = [totalPins]int{
	0, 1, 2, 3,
	6, 7, 8, 9,
	10, 11, 12, 13,
	14, 15, 16, 17,
	21, 22,
	26, 27, 28,
}

var pinStates [totalPins]pinState

var needsRedraw = false

func applyPinMode(idx int) {
	p := pinNumbers[idx]
	st := &pinStates[idx]

	switch st.mode {
	case MODE_INPUT:
		pwmDisable(idx)
		p.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
		st.value = 1

	case MODE_OUTPUT:
		pwmDisable(idx)
		p.Configure(machine.PinConfig{Mode: machine.PinOutput})
		p.Low()
		st.value = 0

	case MODE_PWM:
		pwmSetup(idx)
		if pwmConfigured[idx] {
			st.value = 50
		} else {
			// Пин не поддерживает PWM — оставляем INPUT
			p.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
			st.value = 1
		}
	}
}

func pollInputs() {
	for i := 0; i < totalPins; i++ {
		if pinStates[i].mode != MODE_INPUT {
			continue
		}
		v := uint8(0)
		if pinNumbers[i].Get() {
			v = 1
		}
		if pinStates[i].value != v {
			pinStates[i].value = v
			needsRedraw = true
		}
	}
}
