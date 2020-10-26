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

	pc     uint16 //Program counter
	opcode uint16 //Current opcode
	index  uint16 //Index register
	stkptr uint8  //Stack pointer

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

	memoryLocation := fmt.Sprintf("%X", c.pc-2)
	instruction := fmt.Sprintf("ERR: #%X", c.opcode)
	drawBool := false

	//Instruction decoding
	switch identifier {
	case 0x0:
		if kk == 0xE0 {
			c.CLS()
			instruction = "CLS"
			drawBool = true
		} else if kk == 0xEE {
			c.RET()
			instruction = "RET"
		}
	case 0x1:
		c.JP(addr)
		instruction = fmt.Sprintf("JP #%X", addr)
	case 0x2:
		c.CALL(addr)
		instruction = fmt.Sprintf("CALL #%X", addr)
	case 0x3:
		c.SEVx(x, kk)
		instruction = fmt.Sprintf("SE V%X #%X", x, kk)
	case 0x4:
		c.SNEVx(x, kk)
		instruction = fmt.Sprintf("SNE V%X #%X", x, kk)
	case 0x5:
		c.SEVxVy(x, y)
		instruction = fmt.Sprintf("SE V%X V%X", x, y)
	case 0x6:
		c.LDVx(x, kk)
		instruction = fmt.Sprintf("LD V%X #%X", x, kk)
	case 0x7:
		c.ADDVx(x, kk)
		instruction = fmt.Sprintf("ADD V%X #%X", x, kk)
	case 0x8:
		switch n {
		case 0x0:
			c.LDVxVy(x, y)
			instruction = fmt.Sprintf("LD V%X V%X", x, y)
		case 0x1:
			c.ORVxVy(x, y)
			instruction = fmt.Sprintf("OR V%X V%X", x, y)
		case 0x2:
			c.ANDVxVy(x, y)
			instruction = fmt.Sprintf("AND V%X V%X", x, y)
		case 0x3:
			c.XORVxVy(x, y)
			instruction = fmt.Sprintf("XOR V%X V%X", x, y)
		case 0x4:
			c.ADDVxVy(x, y)
			instruction = fmt.Sprintf("ADD V%X V%X", x, y)
		case 0x5:
			c.SUBVxVy(x, y)
			instruction = fmt.Sprintf("SUB V%X V%X", x, y)
		case 0x6:
			c.SHRVx(x)
			instruction = fmt.Sprintf("SHR V%X", x)
		case 0x7:
			c.SUBNVxVy(x, y)
			instruction = fmt.Sprintf("SUBN V%X V%X", x, y)
		case 0xE:
			c.SHLVx(x)
			instruction = fmt.Sprintf("SHL V%X", x)
		}
	case 0x9:
		c.SNEVxVy(x, y)
		instruction = fmt.Sprintf("SNE V%X V%X", x, y)
	case 0xA:
		c.LDI(addr)
		instruction = fmt.Sprintf("LD I #%X", addr)
	case 0xB:
		c.JPV(addr)
		instruction = fmt.Sprintf("JP V0 #%X", addr)
	case 0xC:
		c.RNDVx(x, kk)
		instruction = fmt.Sprintf("RND V%X #%X", x, kk)
	case 0xD:
		instruction = fmt.Sprintf("DRW V%X V%X #%X", x, y, n)
		c.DRW(x, y, n)
		drawBool = true
	case 0xF:
		switch kk {
		case 0x07:
			c.LDVxDT(x)
			instruction = fmt.Sprintf("LD V%X DT", x)
		case 0x15:
			c.LDDTVx(x)
			instruction = fmt.Sprintf("LD DT V%X", x)
		case 0x18:
			c.LDSTVx(x)
			instruction = fmt.Sprintf("LD ST V%X", x)
		case 0x1E:
			c.ADDIVx(x)
			instruction = fmt.Sprintf("ADD I V%X", x)
		case 0x29:
			c.LDFVx(x)
			instruction = fmt.Sprintf("LD F V%X", x)
		case 0x33:
			c.LDBVx(x)
			instruction = fmt.Sprintf("LD B V%X", x)
		case 0x55:
			c.LDIVx(x)
			instruction = fmt.Sprintf("LD I V%X", x)
		case 0x65:
			c.LDVxI(x)
			instruction = fmt.Sprintf("LD V%X I", x)
		}
	}

	return memoryLocation, instruction, drawBool

}

//The following functions are all the opcodes for the chip8 system

//CLS 00E0
func (c *CPU) CLS() {
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			c.display[y][x] = 0
		}
	}
}

//RET 00EE
func (c *CPU) RET() {
	c.pc = c.stack[c.stkptr-1]
	c.stack[c.stkptr-1] = 0 //clear value from stack
	c.stkptr--
}

//JP 1nnn
func (c *CPU) JP(addr uint16) {
	c.pc = addr
}

//CALL 2nnn
func (c *CPU) CALL(addr uint16) {
	c.stack[c.stkptr] = c.pc
	c.stkptr++
	c.pc = addr
}

//SEVx 3xkk
func (c *CPU) SEVx(x uint8, kk uint8) {
	if c.V[x] == kk {
		c.pc += 2
	}
}

//SNEVx 4xkk
func (c *CPU) SNEVx(x uint8, kk uint8) {
	if c.V[x] != kk {
		c.pc += 2
	}
}

//SEVxVy 5xy0
func (c *CPU) SEVxVy(x uint8, y uint8) {
	if c.V[x] == c.V[y] {
		c.pc += 2
	}
}

//LDVx 6xkk
func (c *CPU) LDVx(x uint8, kk uint8) {
	c.V[x] = kk
}

//ADDVx 7xkk
func (c *CPU) ADDVx(x uint8, kk uint8) {
	c.V[x] += kk
}

//LDVxVy 8xy0
func (c *CPU) LDVxVy(x uint8, y uint8) {
	c.V[x] = c.V[y]
}

//ORVxVy 8xy1
func (c *CPU) ORVxVy(x uint8, y uint8) {
	c.V[x] |= c.V[y]
}

//ANDVxVy 8xy2
func (c *CPU) ANDVxVy(x uint8, y uint8) {
	c.V[x] &= c.V[y]
}

//XORVxVy 8xy3
func (c *CPU) XORVxVy(x uint8, y uint8) {
	c.V[x] ^= c.V[y]
}

//ADDVxVy 8xy4
func (c *CPU) ADDVxVy(x uint8, y uint8) {
	overflow := c.V[x] + c.V[y]
	c.V[0xF] = 0
	if overflow > 255 {
		c.V[0xF] = 1
	}
	c.V[x] += c.V[y]
}

//SUBVxVy 8xy5
func (c *CPU) SUBVxVy(x uint8, y uint8) {
	c.V[0xF] = 0
	if c.V[x] > c.V[y] {
		c.V[0xF] = 1
	}
	c.V[x] -= c.V[y]
}

//SHRVx 8xy6
func (c *CPU) SHRVx(x uint8) {
	c.V[0xF] = c.V[x] & 1
	c.V[x] /= 2
}

//SUBNVxVy 8xy7
func (c *CPU) SUBNVxVy(x uint8, y uint8) {

	c.V[0xF] = 0
	if c.V[y] > c.V[x] {
		c.V[0xF] = 1
	}
	c.V[x] = c.V[y] - c.V[x]
}

//SHLVx 8xyE
func (c *CPU) SHLVx(x uint8) {

	c.V[0xF] = (c.V[x] & 128) >> 7
	c.V[x] *= 2
}

//SNEVxVy 9xy0
func (c *CPU) SNEVxVy(x uint8, y uint8) {
	if c.V[x] != c.V[y] {
		c.pc += 2
	}
}

//LDI Annn
func (c *CPU) LDI(addr uint16) {
	c.index = addr
}

//JPV Bnnn
func (c *CPU) JPV(addr uint16) {
	c.pc = addr + uint16(c.V[0])
}

//RNDVx Cxnn
func (c *CPU) RNDVx(x uint8, kk uint8) {
	c.V[x] = uint8(rand.Intn(256)) & kk
}

//DRW Dxyn
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

//LDVxDT Fx07
func (c *CPU) LDVxDT(x uint8) {
	c.V[x] = uint8(c.delayTimer)
}

//LDDTVx Fx15
func (c *CPU) LDDTVx(x uint8) {
	c.delayTimer = uint16(c.V[x])
}

//LDSTVx
func (c *CPU) LDSTVx(x uint8) {
	c.soundTimer = uint16(c.V[x])
}

//ADDIVx Fx1E
func (c *CPU) ADDIVx(x uint8) {
	//Fx1E
	c.index += uint16(c.V[x])
}

//LDFVx Fx29
func (c *CPU) LDFVx(x uint8) {
	c.index = uint16(5 * c.V[x])
}

//LDBVx Fx33
func (c *CPU) LDBVx(x uint8) {
	value := c.V[x]
	c.memory[c.index] = value / 100
	c.memory[c.index+1] = (value / 10) % 10
	c.memory[c.index+2] = value % 10
	value2 := c.V[x]
	fmt.Println(value, value2)
}

//LDIVx Fx55
func (c *CPU) LDIVx(x uint8) {
	for i := uint16(0); i < uint16(x)+1; i++ {
		c.memory[c.index+i] = c.V[i]
	}
}

//LDVxI Fx65
func (c *CPU) LDVxI(x uint8) {
	for i := uint16(0); i < uint16(x)+1; i++ {
		c.V[i] = c.memory[c.index+i]
	}
}
