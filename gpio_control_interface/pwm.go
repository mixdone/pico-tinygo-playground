package main

import "machine"

type pwmDevice interface {
	Configure(machine.PWMConfig) error
	Channel(machine.Pin) (uint8, error)
	Set(channel uint8, value uint32)
	Top() uint32
}

var pwmPeripheral = [totalPins]pwmDevice{
	machine.PWM0, // GP0
	machine.PWM0, // GP1
	machine.PWM1, // GP2
	machine.PWM1, // GP3
	machine.PWM3, // GP6
	machine.PWM3, // GP7
	machine.PWM4, // GP8
	machine.PWM4, // GP9
	machine.PWM5, // GP10
	machine.PWM5, // GP11
	machine.PWM6, // GP12
	machine.PWM6, // GP13
	machine.PWM7, // GP14
	machine.PWM7, // GP15
	machine.PWM0, // GP16 (конфликт с GP0)
	machine.PWM0, // GP17 (конфликт с GP1)
	machine.PWM2, // GP21
	machine.PWM3, // GP22 (конфликт с GP6)
	machine.PWM5, // GP26 (конфликт с GP10)
	machine.PWM5, // GP27 (конфликт с GP11)
	machine.PWM6, // GP28 (конфликт с GP12)
}

// Какие пины могут быть PWM независимо.
// false — пин конфликтует с другим (разница 16), PWM ему не даём.
var pwmCapable = [totalPins]bool{
	true, true, true, true, // GP0, GP1, GP2, GP3
	true, true, true, true, // GP6, GP7, GP8, GP9
	true, true, true, true, // GP10, GP11, GP12, GP13
	true, true, // GP14, GP15
	false, false, // GP16, GP17 — конфликт
	true,                // GP21
	false,               // GP22 — конфликт
	false, false, false, // GP26, GP27, GP28 — конфликт
}

// Состояние PWM на каждый пин
var (
	pwmChannels   [totalPins]uint8
	pwmConfigured [totalPins]bool
)

const pwmFreqHz = 1000

// Настроить PWM на пине idx (вызывается при переключении в MODE_PWM)
func pwmSetup(idx int) {
	if !pwmCapable[idx] {
		return
	}

	pwm := pwmPeripheral[idx]

	// Configure можно вызывать повторно — переустановит период
	pwm.Configure(machine.PWMConfig{
		Period: 1e9 / pwmFreqHz,
	})

	ch, err := pwm.Channel(pinNumbers[idx])
	if err != nil {
		// Пин не поддерживает PWM — оставляем как INPUT
		pinNumbers[idx].Configure(machine.PinConfig{Mode: machine.PinInputPullup})
		pwmConfigured[idx] = false
		return
	}

	pwmChannels[idx] = ch
	pwmConfigured[idx] = true

	// Стартовая скважность — 50%
	pwm.Set(ch, pwm.Top()/2)
}

// Установить скважность 0..100
func pwmSetDuty(idx int, percent uint8) {
	if !pwmConfigured[idx] {
		return
	}
	pwm := pwmPeripheral[idx]
	ch := pwmChannels[idx]
	pwm.Set(ch, pwm.Top()*uint32(percent)/100)
}

// Отключить PWM на пине (при переходе в IN/OUT)
func pwmDisable(idx int) {
	if !pwmConfigured[idx] {
		return
	}
	pwm := pwmPeripheral[idx]
	ch := pwmChannels[idx]
	pwm.Set(ch, 0) // скважность 0 — фактически выключено
	pwmConfigured[idx] = false
}
