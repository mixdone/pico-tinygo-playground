package main

import "machine"

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
