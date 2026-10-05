package main

import (
	"machine"
	"time"
)

// --- SSD1306 ---

const (
	oledAddr = 0x3C
	oledCmd  = 0x00
	oledData = 0x40
)

var framebuffer [8][128]byte

var initCmds = []byte{
	0xAE,
	0xD5, 0x80,
	0xA8, 0x3F,
	0xD3, 0x00,
	0x40,
	0x8D, 0x14,
	0x20, 0x00,
	0xA1,
	0xC8,
	0xDA, 0x12,
	0x81, 0xCF,
	0xD9, 0xF1,
	0xDB, 0x40,
	0xA4,
	0xA6,
	0xAF,
}

// --- Пины и их состояние ---

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

// --- Конфигурация списка пинов ---

// Все свободные GPIO, кроме занятых:
//
//	GP4, GP5   — I2C (OLED)
//	GP18/19/20 — энкодер
const totalPins = 21
const visibleRows = 6

var pinNumbers = [totalPins]machine.Pin{
	machine.GP0, machine.GP1, machine.GP2, machine.GP3,
	machine.GP6, machine.GP7, machine.GP8, machine.GP9,
	machine.GP10, machine.GP11, machine.GP12, machine.GP13,
	machine.GP14, machine.GP15, machine.GP16, machine.GP17,
	machine.GP21, machine.GP22,
	machine.GP26, machine.GP27, machine.GP28,
}

// Номера GPIO в том же порядке, что pinNumbers — для отображения
var pinNums = [totalPins]int{
	0, 1, 2, 3,
	6, 7, 8, 9,
	10, 11, 12, 13,
	14, 15, 16, 17,
	21, 22,
	26, 27, 28,
}

var pinStates = [totalPins]pinState{
	{MODE_OUTPUT, 1},
	{MODE_INPUT, 0},
	{MODE_PWM, 50},
	{MODE_OUTPUT, 0},
	{MODE_INPUT, 1},
	{MODE_OUTPUT, 0},
	{MODE_INPUT, 0},
	{MODE_PWM, 25},
	{MODE_OUTPUT, 1},
	{MODE_INPUT, 1},
	{MODE_PWM, 75},
	{MODE_OUTPUT, 0},
	{MODE_INPUT, 0},
	{MODE_OUTPUT, 1},
	{MODE_PWM, 50},
	{MODE_INPUT, 0},
	{MODE_OUTPUT, 0},
	{MODE_INPUT, 1},
	{MODE_PWM, 10},
	{MODE_OUTPUT, 0},
	{MODE_INPUT, 1},
}

var (
	cursor    = 0
	scrollTop = 0
)

// --- Энкодер ---

var (
	pinA   = machine.GP20
	pinB   = machine.GP19
	pinKey = machine.GP18
)

var stepTable = [4][4]int8{
	{0, -1, +1, 0},
	{+1, 0, 0, -1},
	{-1, 0, 0, +1},
	{0, +1, -1, 0},
}

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

// --- main ---

func main() {
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

	redraw()

	prev := readEncoder()

	for {
		cur := readEncoder()
		if cur != prev {
			step := stepTable[prev][cur]
			if step != 0 {
				moveCursor(int(step))
				redraw()
			}
			prev = cur
		}

		handleButton()
		if btnPressed {
			btnPressed = false
			cycleMode()
			redraw()
		}
	}
}

// --- Скролл и курсор ---

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

func cycleMode() {
	switch pinStates[cursor].mode {
	case MODE_INPUT:
		pinStates[cursor].mode = MODE_OUTPUT
		pinStates[cursor].value = 0
	case MODE_OUTPUT:
		pinStates[cursor].mode = MODE_PWM
		pinStates[cursor].value = 50
	case MODE_PWM:
		pinStates[cursor].mode = MODE_INPUT
		pinStates[cursor].value = 0
	}
}

// --- Отрисовка ---

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
	if idx == cursor {
		drawString(0, y, ">")
	} else {
		drawString(0, y, " ")
	}

	drawPinName(12, y, idx)

	switch pinStates[idx].mode {
	case MODE_INPUT:
		drawString(48, y, "IN ")
	case MODE_OUTPUT:
		drawString(48, y, "OUT")
	case MODE_PWM:
		drawString(48, y, "PWM")
	}

	drawValue(72, y, pinStates[idx])
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

// drawScrollIndicator — тонкая полоска справа, показывает позицию в списке
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
		pos = uint8((int(trackHeight-barHeight) * (cursor - scrollTop)) / (visibleRows - 1))
	}

	for i := uint8(0); i < barHeight; i++ {
		setPixel(126, trackTop+pos+i, true)
		setPixel(127, trackTop+pos+i, true)
	}
}

// --- Энкодер ---

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

// --- SSD1306 ---

func initDisplay() {
	buf := make([]byte, 1+len(initCmds))
	buf[0] = oledCmd
	copy(buf[1:], initCmds)
	machine.I2C0.Tx(oledAddr, buf, nil)
}

func setCursor(page, col uint8) {
	cmds := []byte{
		oledCmd,
		0xB0 | page,
		0x00 | (col & 0x0F),
		0x10 | (col >> 4),
	}
	machine.I2C0.Tx(oledAddr, cmds, nil)
}

func sendFramebuffer() {
	for page := uint8(0); page < 8; page++ {
		setCursor(page, 0)

		buf := make([]byte, 129)
		buf[0] = oledData
		copy(buf[1:], framebuffer[page][:])

		machine.I2C0.Tx(oledAddr, buf, nil)
	}
}

// --- Рисование ---

func clearBuffer() {
	for page := 0; page < 8; page++ {
		for col := 0; col < 128; col++ {
			framebuffer[page][col] = 0
		}
	}
}

func setPixel(x, y uint8, on bool) {
	if x >= 128 || y >= 64 {
		return
	}
	page := y / 8
	bit := y % 8

	if on {
		framebuffer[page][x] |= (1 << bit)
	} else {
		framebuffer[page][x] &^= (1 << bit)
	}
}

func drawHLine(x, y, w uint8) {
	for i := uint8(0); i < w; i++ {
		setPixel(x+i, y, true)
	}
}

func drawChar(x, y uint8, ch byte) {
	if ch < 32 || ch > 127 {
		return
	}
	idx := ch - 32
	for col := uint8(0); col < 5; col++ {
		bits := font5x7[idx][col]
		for row := uint8(0); row < 7; row++ {
			if bits&(1<<row) != 0 {
				setPixel(x+col, y+row, true)
			}
		}
	}
}

func drawString(x, y uint8, s string) {
	for i := 0; i < len(s); i++ {
		drawChar(x+uint8(i)*6, y, s[i])
	}
}

// --- Шрифт Adafruit 5x7 (96 символов ASCII, 32..127) ---

var font5x7 = [96][5]byte{
	{0x00, 0x00, 0x00, 0x00, 0x00}, // ' '
	{0x00, 0x00, 0x5F, 0x00, 0x00}, // '!'
	{0x00, 0x07, 0x00, 0x07, 0x00}, // '"'
	{0x14, 0x7F, 0x14, 0x7F, 0x14}, // '#'
	{0x24, 0x2A, 0x7F, 0x2A, 0x12}, // '$'
	{0x23, 0x13, 0x08, 0x64, 0x62}, // '%'
	{0x36, 0x49, 0x56, 0x20, 0x50}, // '&'
	{0x00, 0x08, 0x07, 0x03, 0x00}, // '\''
	{0x00, 0x1C, 0x22, 0x41, 0x00}, // '('
	{0x00, 0x41, 0x22, 0x1C, 0x00}, // ')'
	{0x2A, 0x1C, 0x7F, 0x1C, 0x2A}, // '*'
	{0x08, 0x08, 0x3E, 0x08, 0x08}, // '+'
	{0x00, 0x80, 0x70, 0x30, 0x00}, // ','
	{0x08, 0x08, 0x08, 0x08, 0x08}, // '-'
	{0x00, 0x00, 0x60, 0x60, 0x00}, // '.'
	{0x20, 0x10, 0x08, 0x04, 0x02}, // '/'
	{0x3E, 0x51, 0x49, 0x45, 0x3E}, // '0'
	{0x00, 0x42, 0x7F, 0x40, 0x00}, // '1'
	{0x72, 0x49, 0x49, 0x49, 0x46}, // '2'
	{0x21, 0x41, 0x49, 0x4D, 0x33}, // '3'
	{0x18, 0x14, 0x12, 0x7F, 0x10}, // '4'
	{0x27, 0x45, 0x45, 0x45, 0x39}, // '5'
	{0x3C, 0x4A, 0x49, 0x49, 0x31}, // '6'
	{0x41, 0x21, 0x11, 0x09, 0x07}, // '7'
	{0x36, 0x49, 0x49, 0x49, 0x36}, // '8'
	{0x46, 0x49, 0x49, 0x29, 0x1E}, // '9'
	{0x00, 0x00, 0x14, 0x00, 0x00}, // ':'
	{0x00, 0x40, 0x34, 0x00, 0x00}, // ';'
	{0x00, 0x08, 0x14, 0x22, 0x41}, // '<'
	{0x14, 0x14, 0x14, 0x14, 0x14}, // '='
	{0x00, 0x41, 0x22, 0x14, 0x08}, // '>'
	{0x02, 0x01, 0x59, 0x09, 0x06}, // '?'
	{0x3E, 0x41, 0x5D, 0x59, 0x4E}, // '@'
	{0x7C, 0x12, 0x11, 0x12, 0x7C}, // 'A'
	{0x7F, 0x49, 0x49, 0x49, 0x36}, // 'B'
	{0x3E, 0x41, 0x41, 0x41, 0x22}, // 'C'
	{0x7F, 0x41, 0x41, 0x41, 0x3E}, // 'D'
	{0x7F, 0x49, 0x49, 0x49, 0x41}, // 'E'
	{0x7F, 0x09, 0x09, 0x09, 0x01}, // 'F'
	{0x3E, 0x41, 0x41, 0x51, 0x73}, // 'G'
	{0x7F, 0x08, 0x08, 0x08, 0x7F}, // 'H'
	{0x00, 0x41, 0x7F, 0x41, 0x00}, // 'I'
	{0x20, 0x40, 0x41, 0x3F, 0x01}, // 'J'
	{0x7F, 0x08, 0x14, 0x22, 0x41}, // 'K'
	{0x7F, 0x40, 0x40, 0x40, 0x40}, // 'L'
	{0x7F, 0x02, 0x1C, 0x02, 0x7F}, // 'M'
	{0x7F, 0x04, 0x08, 0x10, 0x7F}, // 'N'
	{0x3E, 0x41, 0x41, 0x41, 0x3E}, // 'O'
	{0x7F, 0x09, 0x09, 0x09, 0x06}, // 'P'
	{0x3E, 0x41, 0x51, 0x21, 0x5E}, // 'Q'
	{0x7F, 0x09, 0x19, 0x29, 0x46}, // 'R'
	{0x26, 0x49, 0x49, 0x49, 0x32}, // 'S'
	{0x03, 0x01, 0x7F, 0x01, 0x03}, // 'T'
	{0x3F, 0x40, 0x40, 0x40, 0x3F}, // 'U'
	{0x1F, 0x20, 0x40, 0x20, 0x1F}, // 'V'
	{0x3F, 0x40, 0x38, 0x40, 0x3F}, // 'W'
	{0x63, 0x14, 0x08, 0x14, 0x63}, // 'X'
	{0x03, 0x04, 0x78, 0x04, 0x03}, // 'Y'
	{0x61, 0x59, 0x49, 0x4D, 0x43}, // 'Z'
	{0x00, 0x7F, 0x41, 0x41, 0x41}, // '['
	{0x02, 0x04, 0x08, 0x10, 0x20}, // '\\'
	{0x00, 0x41, 0x41, 0x41, 0x7F}, // ']'
	{0x04, 0x02, 0x01, 0x02, 0x04}, // '^'
	{0x40, 0x40, 0x40, 0x40, 0x40}, // '_'
	{0x00, 0x03, 0x07, 0x08, 0x00}, // '`'
	{0x20, 0x54, 0x54, 0x78, 0x40}, // 'a'
	{0x7F, 0x28, 0x44, 0x44, 0x38}, // 'b'
	{0x38, 0x44, 0x44, 0x44, 0x28}, // 'c'
	{0x38, 0x44, 0x44, 0x28, 0x7F}, // 'd'
	{0x38, 0x54, 0x54, 0x54, 0x18}, // 'e'
	{0x00, 0x08, 0x7E, 0x09, 0x02}, // 'f'
	{0x18, 0xA4, 0xA4, 0x9C, 0x78}, // 'g'
	{0x7F, 0x08, 0x04, 0x04, 0x78}, // 'h'
	{0x00, 0x44, 0x7D, 0x40, 0x00}, // 'i'
	{0x20, 0x40, 0x40, 0x3D, 0x00}, // 'j'
	{0x7F, 0x10, 0x28, 0x44, 0x00}, // 'k'
	{0x00, 0x41, 0x7F, 0x40, 0x00}, // 'l'
	{0x7C, 0x04, 0x78, 0x04, 0x78}, // 'm'
	{0x7C, 0x08, 0x04, 0x04, 0x78}, // 'n'
	{0x38, 0x44, 0x44, 0x44, 0x38}, // 'o'
	{0xFC, 0x18, 0x24, 0x24, 0x18}, // 'p'
	{0x18, 0x24, 0x24, 0x18, 0xFC}, // 'q'
	{0x7C, 0x08, 0x04, 0x04, 0x08}, // 'r'
	{0x48, 0x54, 0x54, 0x54, 0x24}, // 's'
	{0x04, 0x04, 0x3F, 0x44, 0x24}, // 't'
	{0x3C, 0x40, 0x40, 0x20, 0x7C}, // 'u'
	{0x1C, 0x20, 0x40, 0x20, 0x1C}, // 'v'
	{0x3C, 0x40, 0x30, 0x40, 0x3C}, // 'w'
	{0x44, 0x28, 0x10, 0x28, 0x44}, // 'x'
	{0x4C, 0x90, 0x90, 0x90, 0x7C}, // 'y'
	{0x44, 0x64, 0x54, 0x4C, 0x44}, // 'z'
	{0x00, 0x08, 0x36, 0x41, 0x00}, // '{'
	{0x00, 0x00, 0x77, 0x00, 0x00}, // '|'
	{0x00, 0x41, 0x36, 0x08, 0x00}, // '}'
	{0x02, 0x01, 0x02, 0x04, 0x02}, // '~'
	{0x3C, 0x26, 0x23, 0x26, 0x3C}, // DEL
}
