package main

import "machine"

var (
	pinA = machine.GP20
	pinB = machine.GP19
)

var stepTable = [4][4]int8{
	{0, -1, +1, 0},
	{+1, 0, 0, -1},
	{-1, 0, 0, +1},
	{0, +1, -1, 0},
}

func readEncoder() int {
	state := 0
	if pinA.Get() {
		state |= 2
	}
	if pinB.Get() {
		state |= 1
	}
	return state
}
