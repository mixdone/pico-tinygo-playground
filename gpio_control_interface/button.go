package main

import (
	"machine"
	"time"
)

var pinKey = machine.GP18

type buttonState int

const (
	IDLE buttonState = iota
	DEBOUNCING
	PRESSED
)

var (
	btnState   buttonState = IDLE
	btnTimer   time.Time
	btnPressed = false
)

func handleButton() {
	now := time.Now()
	pressed := !pinKey.Get()

	switch btnState {
	case IDLE:
		if pressed {
			btnState = DEBOUNCING
			btnTimer = now
		}
	case DEBOUNCING:
		if !pressed {
			btnState = IDLE
		} else if now.Sub(btnTimer) >= 20*time.Millisecond {
			btnPressed = true
			btnState = PRESSED
		}
	case PRESSED:
		if !pressed {
			btnState = IDLE
		}
	}
}
