package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
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
	//00E0
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			c.display[y][x] = 0
		}
	}
}

func (c *CPU) RET() {
	//00EE
	c.pc = c.stack[c.stkptr-1]
	c.stkptr--
}

func (c *CPU) JP(addr uint16) {
	//1nnn
	c.pc = addr
}

func (c *CPU) CALL(addr uint16) {
	//2nnn
	c.stkptr++
	c.stack[c.stkptr] = addr
}

func (c *CPU) SEVx(x uint8, kk uint8) {
	//3xkk
	if c.V[x] == kk {
		c.pc += 2
	}
}

func (c *CPU) SNEVx(x uint8, kk uint8) {
	//4xkk
	if c.V[x] != kk {
		c.pc += 2
	}
}

func (c *CPU) SEVxVy(x uint8, y uint8) {
	//5xy0
	if c.V[x] == c.V[y] {
		c.pc += 2
	}
}

func (c *CPU) LDVx(x uint8, kk uint8) {
	//6xkk
	c.V[x] = kk
}

func (c *CPU) ADDVx(x uint8, kk uint8) {
	//7xkk
	c.V[x] += kk
}

func (c *CPU) LDVxVy(x uint8, y uint8) {
	//8xy0
	c.V[x] = c.V[y]
}

func (c *CPU) ORVxVy(x uint8, y uint8) {
	//8xy1
	c.V[x] |= c.V[y]
}

func (c *CPU) ANDVxVy(x uint8, y uint8) {
	//8xy2
	c.V[x] &= c.V[y]
}

func (c *CPU) XORVxVy(x uint8, y uint8) {
	//8xy3
	c.V[x] ^= c.V[y]
}

func (c *CPU) ADDVxVy(x uint8, y uint8) {
	//8xy4 UNIMPLEMENTED
}

func (c *CPU) SUBVxVy(x uint8, y uint8) {
	//8xy5 UNIMPLEMENTED
}

func (c *CPU) SHRVx(x uint8) {
	//8xy6 POSSIBLE PROBLEM
	c.V[0xF] = c.V[x] & 1
	c.V[x] /= 2
}

func (c *CPU) SUBNVxVy(x uint8, y uint8) {
	//8xy7
	c.V[0xF] = 0
	if c.V[y] > c.V[x] {
		c.V[0xF] = 1
	}
	c.V[x] = c.V[y] - c.V[x]
}

func (c *CPU) SHLVx(x uint8) {
	//8xyE
	c.V[0xF] = c.V[x] & 128
	c.V[x] *= 2
}

func (c *CPU) SNEVxVy(x uint8, y uint8) {
	if c.V[x] != c.V[y] {
		c.pc += 2
	}
}

func (c *CPU) LDI(addr uint16) {
	//Annn
	c.index = addr
}

func (c *CPU) JPV(addr uint16) {
	//Bnnn
	c.pc = addr + uint16(c.V[0])
}

func (c *CPU) RNDVx(x uint8, kk uint8) {
	//Cxnn
	c.V[x] = uint8(rand.Intn(256)) & kk
}

func (c *CPU) DRW(x uint8, y uint8, n uint8) {
	//Dxyn
	xcoord := c.V[x] % 64 //modulo to wrap coords
	ycoord := c.V[y] % 32 //modulo to wrap coords
	c.V[0xF] = 0

	for y := uint16(0); y < uint16(n); y++ {
		byteData := c.memory[c.index+y]
		for x := 0; x < 8; x++ {
			if xcoord < 64 && ycoord < 32 {
				bitData := byteData & uint8(math.Pow(2, float64(7-x))) >> (7 - x)
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

func (c *CPU) ADDIVx(x uint8) {
	//Fx1E
	c.index += uint16(x)
}

func (c *CPU) LDFVx(x uint8) {
	//Fx29
	c.index = uint16(5 * c.V[x])
}

func (c *CPU) LDBVx(x uint8) {
	//Fx33
	value := c.V[x]
	c.memory[c.index] = value / 100
	c.memory[c.index+1] = (value / 10) % 10
	c.memory[c.index+2] = value % 10
}

func (c *CPU) LDIVx(x uint8) {
	//Fx55 POSSIBLE PROBLEM
	for i := uint16(0); i < uint16(x)+1; i++ {
		c.memory[c.index+i] = c.V[i]
	}
}

func (c *CPU) LDVXI(x uint8) {
	//Fx65 POSSIBLE PROBLEM
	for i := uint16(0); i < uint16(x)+1; i++ {
		c.V[i] = c.memory[c.index+i]
	}
}
