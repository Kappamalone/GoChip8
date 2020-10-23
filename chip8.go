package main

import (
	"fmt"
	"io/ioutil"
	"math"
)

//CPU describes the general shape of the CHIP-8
type CPU struct {
	//Fonts are loaded in from 0x00

	display [32][64]uint8 //64 x 32 display
	memory  [4096]uint8   //4k of memory
	V       [16]uint8     //Register V0-VF
	stack   [16]uint16    //16 levels of stack

	pc         uint16 //Program counter
	opcode     uint16 //Current opcode
	index      uint16 //Index register
	stkptr     uint8  //Stack pointer
	delayTimer uint16 //Delay timer
	soundTimer uint16 //Sound timer
}

func initCPU(rom string) *CPU {
	cpu := new(CPU)
	cpu.pc = 0x200
	cpu.loadFonts()
	cpu.loadRom(rom)

	return cpu

}

func (c *CPU) loadFonts() {
	//Loads in font data from 0x00

	var fontset = []uint8{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}

	for i := 0; i < len(fontset); i++ {
		c.memory[i] = fontset[i]
	}
}

func (c *CPU) loadRom(filePath string) {
	//Loads rom into memory from 0x200

	fileData, readErr := ioutil.ReadFile(filePath)
	if readErr != nil {
		fmt.Println(readErr)
	}

	for i := 0; i < len(fileData); i++ {
		c.memory[0x200+i] = fileData[i]
	}

}

func (c *CPU) cycle() (string, string, bool) {
	//The fetch-decode-cycle for the system
	c.opcode = uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])
	c.pc += 2

	return c.decodeAndExecute()
}

func (c *CPU) decodeAndExecute() (string, string, bool) {
	//Handles getting operands from the opcode and executing them

	identifier := (c.opcode & 0xF000) >> 12
	addr := (c.opcode & 0x0FFF)
	kk := uint8(c.opcode & 0x00FF)
	x := uint8(c.opcode & 0x0F00 >> 8)
	y := uint8(c.opcode&0x00F0) >> 4
	n := uint8(c.opcode & 0x000F)

	memoryLocation := fmt.Sprintf("%x", c.pc-2)
	instruction := "Err"
	drawBool := false

	//Instruction decoding
	switch identifier {
	case 0x0:
		if kk == 0xE0 {
			c.CLS()
			instruction = "CLS"
			drawBool = true
		}
	case 0x1:
		c.JP(addr)
		instruction = fmt.Sprintf("JP %x", addr)
	case 0x6:
		c.LDVx(x, kk)
		instruction = fmt.Sprintf("LD V %x %x", x, kk)
	case 0x7:
		c.ADDVx(x, kk)
		instruction = fmt.Sprintf("ADD V %x %x", x, kk)
	case 0xA:
		c.LDI(addr)
		instruction = fmt.Sprintf("LD I %x", addr)
	case 0xD:
		instruction = fmt.Sprintf("DRW %x %x %x", x, y, n)
		c.DRW(x, y, n)
		drawBool = true
	}

	return memoryLocation, instruction, drawBool

}

//The following functions are all the opcodes for the chip8 system

func (c *CPU) CLS() {
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			c.display[y][x] = 0
		}
	}
}

func (c *CPU) JP(addr uint16) {
	c.pc = addr
}

func (c *CPU) LDVx(x uint8, kk uint8) {
	c.V[x] = kk
}

func (c *CPU) ADDVx(x uint8, kk uint8) {
	c.V[x] += kk
}

func (c *CPU) LDI(nnn uint16) {
	c.index = nnn
}

func (c *CPU) DRW(x uint8, y uint8, n uint8) {
	xcoord := c.V[x] % 64 //modulo to wrap coords
	ycoord := c.V[y] % 32 //modulo to wrap coords
	c.V[0xF] = 0

	for y := uint16(0); y < uint16(n); y++ {
		byteData := c.memory[c.index+y]
		for x := 0; x < 8; x++ {
			if xcoord < 64 && ycoord < 32 {
				var bitData uint8 = 0
				if (byteData & uint8(math.Pow(2, float64(7-x)))) != 0 {
					bitData = 1
				}

				c.display[ycoord][xcoord] ^= bitData

				if bitData == 1 && c.display[ycoord][xcoord] == 0 {
					c.V[0xF] = 1
				}

			}
			xcoord++
		}
		xcoord -= 8 //Sprites are eight by 8, and so the xcoord should be shifted accordingly for each line, kind of like a typewriter
		ycoord++
	}
}
