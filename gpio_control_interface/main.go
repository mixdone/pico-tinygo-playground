package main

import (
	"machine"
	"time"
)

func main() {
	setupHardware()

	for i := 0; i < totalPins; i++ {
		pinStates[i].mode = MODE_INPUT
		pinStates[i].value = 1
		applyPinMode(i)
	}

	redraw()

	prev := readEncoder()

	for {
		cur := readEncoder()
		if cur != prev {
			step := stepTable[prev][cur]
			if step != 0 {
				onEncoder(int(step))
				redraw()
			}
			prev = cur
		}

		handleButton()
		if btnPressed {
			btnPressed = false
			onButton()
			redraw()
		}

		pollInputs()
		if needsRedraw {
			needsRedraw = false
			redraw()
		}
	}
}

func setupHardware() {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400_000,
		SDA:       machine.GP4,
		SCL:       machine.GP5,
	})

	pinA.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinB.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinKey.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	time.Sleep(time.Second)
	initDisplay()
}
