package main

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
