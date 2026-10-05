package main

// --- FSM интерфейса ---

type uiState int

const (
	UI_NAV uiState = iota
	UI_MODE
	UI_VALUE
)

var ui = UI_NAV

var uiNext = [3]uiState{UI_MODE, UI_VALUE, UI_NAV}

// --- Скролл и курсор ---

const visibleRows = 6

var (
	cursor    = 0
	scrollTop = 0
)

func moveCursor(delta int) {
	cursor += delta

	if cursor < 0 {
		cursor = totalPins - 1
	}
	if cursor >= totalPins {
		cursor = 0
	}

	ensureCursorVisible()
}

func ensureCursorVisible() {
	if cursor < scrollTop {
		scrollTop = cursor
	}
	if cursor >= scrollTop+visibleRows {
		scrollTop = cursor - visibleRows + 1
	}
}

// --- Диспетчеры ---

func onEncoder(delta int) {
	switch ui {
	case UI_NAV:
		moveCursor(delta)
	case UI_MODE:
		cycleModeOf(cursor, delta)
	case UI_VALUE:
		changeValueOf(cursor, delta)
	}
}

func onButton() {
	ui = uiNext[ui]
}

// --- Изменение режима и значения ---

var modeNext = [3]pinMode{MODE_OUTPUT, MODE_PWM, MODE_INPUT}
var modePrev = [3]pinMode{MODE_PWM, MODE_INPUT, MODE_OUTPUT}

func cycleModeOf(idx int, delta int) {
	st := &pinStates[idx]

	// Пробуем следующий режим
	if delta > 0 {
		st.mode = modeNext[st.mode]
	} else {
		st.mode = modePrev[st.mode]
	}

	if st.mode == MODE_PWM && !pwmCapable[idx] {
		if delta > 0 {
			st.mode = MODE_INPUT
		} else {
			st.mode = MODE_OUTPUT
		}
	}

	applyPinMode(idx)
}

func changeValueOf(idx int, delta int) {
	st := &pinStates[idx]

	switch st.mode {
	case MODE_INPUT:
		return

	case MODE_OUTPUT:
		if delta > 0 {
			st.value = 1
			pinNumbers[idx].High()
		} else {
			st.value = 0
			pinNumbers[idx].Low()
		}

	case MODE_PWM:
		v := int(st.value) + delta*5
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		st.value = uint8(v)
		pwmSetDuty(idx, st.value)
	}
}

// --- Отрисовка ---

const (
	colCursor = 0
	colName   = 12
	colMode   = 42
	colModeX  = 48
	colValue  = 66
	colValueX = 72
)

func redraw() {
	clearBuffer()
	drawInterface()
	sendFramebuffer()
}

func drawInterface() {
	drawString(0, 0, "GPIO CONTROL")
	drawHLine(0, 8, 128)

	for row := 0; row < visibleRows; row++ {
		idx := scrollTop + row
		if idx >= totalPins {
			break
		}
		y := uint8(11 + row*9)
		drawPinRow(y, idx)
	}

	drawScrollIndicator()
}

func drawPinRow(y uint8, idx int) {
	isCursor := (idx == cursor)

	if isCursor && ui == UI_NAV {
		drawString(colCursor, y, ">")
	} else {
		drawString(colCursor, y, " ")
	}

	drawPinName(colName, y, idx)

	if isCursor && ui == UI_MODE {
		drawString(colMode, y, ">")
	} else {
		drawString(colMode, y, " ")
	}

	switch pinStates[idx].mode {
	case MODE_INPUT:
		drawString(colModeX, y, "IN ")
	case MODE_OUTPUT:
		drawString(colModeX, y, "OUT")
	case MODE_PWM:
		drawString(colModeX, y, "PWM")
	}

	if isCursor && ui == UI_VALUE {
		drawString(colValue, y, ">")
	} else {
		drawString(colValue, y, " ")
	}

	drawValue(colValueX, y, pinStates[idx])
}

func drawPinName(x, y uint8, idx int) {
	num := pinNums[idx]

	drawChar(x, y, 'G')
	drawChar(x+6, y, 'P')

	if num >= 10 {
		drawChar(x+12, y, byte('0'+num/10))
		drawChar(x+18, y, byte('0'+num%10))
		drawChar(x+24, y, ' ')
	} else {
		drawChar(x+12, y, byte('0'+num))
		drawChar(x+18, y, ' ')
		drawChar(x+24, y, ' ')
	}
}

func drawValue(x, y uint8, p pinState) {
	switch p.mode {
	case MODE_INPUT, MODE_OUTPUT:
		if p.value == 0 {
			drawChar(x, y, '0')
		} else {
			drawChar(x, y, '1')
		}
	case MODE_PWM:
		v := p.value
		if v >= 100 {
			drawString(x, y, "100%")
		} else if v >= 10 {
			drawChar(x, y, byte('0'+v/10))
			drawChar(x+6, y, byte('0'+v%10))
			drawChar(x+12, y, '%')
		} else {
			drawChar(x, y, byte('0'+v))
			drawChar(x+6, y, '%')
		}
	}
}

func drawScrollIndicator() {
	barHeight := uint8(6)
	trackTop := uint8(10)
	trackBottom := uint8(63)
	trackHeight := trackBottom - trackTop

	maxScroll := totalPins - visibleRows
	if maxScroll <= 0 {
		return
	}

	pos := uint8(0)
	if visibleRows > 1 {
		pos = uint8((int(trackHeight-barHeight) *
			(cursor - scrollTop)) / (visibleRows - 1))
	}

	for i := uint8(0); i < barHeight; i++ {
		setPixel(126, trackTop+pos+i, true)
		setPixel(127, trackTop+pos+i, true)
	}
}
