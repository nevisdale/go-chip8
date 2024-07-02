package chip8

import (
	"fmt"
	"log"
	"math"
	v2 "math/rand/v2"
	"os"
	"time"
)

const (
	RamSizeBytes = 0x1000 // 4096
	EntryPoint   = 0x200  // 512

	// from 0x000 to 0x1FF is Reserved for interpreter
	//
	// see more http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#2.1
	RomMaxSizeBytes = RamSizeBytes - EntryPoint

	// The original implementation of the Chip-8 language used
	// a 64x32-pixel monochrome display
	ScreenWidth  = 64
	ScreenHeight = 32
	ScreenSize   = ScreenWidth * ScreenHeight

	// Ticks per second
	// see more http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#2.5
	DefaultTPS = 60

	StackMaxSize = 16
)

// http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#font
var font []byte = []byte{
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

type State int

const (
	StateRunning State = iota
	StatePaused
	StateQuit
)

type Chip8 struct {
	ram [RamSizeBytes]byte
	rom Rom

	State State

	Screen [ScreenSize]bool

	// http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#2.3
	KeyPad [0x10]bool

	// 16 general purpose 8-bit registers
	regsV [0x10]uint8

	// There is also a 16-bit register called I.
	// This register is generally used to store memory addresses,
	// so only the lowest (rightmost) 12 bits are usually used.
	regI uint16

	// Used to store the currently executing address.
	pc uint16

	// The stack is an array of 16 16-bit values,
	// used to store the address that the interpreter shoud return to when finished with a subroutine.
	// Chip-8 allows for up to 16 levels of nested subroutines
	stack [StackMaxSize]uint16

	// used to point to the next level of the stack.
	// starts from 0
	sp uint8

	// TODO: need to implement
	// delayTimer uint8
	// soundTimer uint8
}

func NewChip8() Chip8 {
	chip8 := Chip8{
		State: StateRunning,
		pc:    EntryPoint,
	}

	copy(chip8.ram[:], font)

	return chip8
}

func (c *Chip8) LoadRom(rom Rom) {
	c.rom = rom
	copy(c.ram[c.pc:], rom.Data)
}

func (c Chip8) ScreenSize() (int, int) {
	return ScreenWidth, ScreenHeight
}

func (c Chip8) GetTPS() int {
	return DefaultTPS
}

func (c Chip8) GetRomName() string {
	return c.rom.Name
}

func (c *Chip8) Emulate() {
	// emulate 60 hz
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		howMuchToSleep := (time.Second / 60) - elapsed
		time.Sleep(howMuchToSleep)
	}()

	if c.pc >= RamSizeBytes {
		return
	}

	opcode := uint16(c.ram[c.pc])<<8 | uint16(c.ram[c.pc+1])
	typ := uint8((opcode >> 12) & 0x0f)
	nnn := uint16(opcode & 0x0fff)
	nn := uint8(opcode & 0x00ff)
	n := uint8(opcode & 0x000f)
	x := uint8((opcode >> 8) & 0x0f)
	y := uint8((opcode >> 4) & 0x0f)
	opcodeString := "unimplemented"

	c.pc += 2

	// Standard Chip-8 Instructions
	//
	// http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#3.0
	switch typ {
	case 0x0:
		switch nn {

		// 00E0
		// Clears the screen
		case 0xe0:
			c.clearScreen()
			opcodeString = "clear screen"

		// 00EE
		// Returns from a subroutine
		case 0xee:
			if c.sp == 0 {
				log.Println("incorrect stack")
				os.Exit(1)
			}

			c.sp--
			c.pc = c.stack[c.sp]
			opcodeString = fmt.Sprintf("return (jump to %04X)", c.pc)

		// This instruction is only used on the old computers on which Chip-8 was originally implemented.
		// It is ignored by modern interpreters.
		default:
			log.Println("unsupport 0NNN")
		}

	// 1NNN
	// Jumps to address NNN
	case 0x01:
		c.pc = nnn
		opcodeString = fmt.Sprintf("jump to %04X", c.pc)

	// 2NNN
	// Calls subroutine at NNN
	case 0x02:
		if c.sp == StackMaxSize {
			log.Println("stack overflow")
			os.Exit(1)
		}
		c.stack[c.sp] = c.pc
		c.sp++
		c.pc = nnn
		opcodeString = fmt.Sprintf("call %04X (jump to %04X)", c.pc, c.pc)

	// 3XNN
	// Skips the next instruction if VX equals NN
	case 0x03:
		if c.regsV[x] == nn {
			c.pc += 2
			opcodeString = fmt.Sprintf("jump to %04X because v%X == %02X", c.pc, x, nn)
		} else {
			opcodeString = fmt.Sprintf("continue to %04X because v%X != %02X", c.pc, x, nn)
		}

	// 4XNN
	// Skips the next instruction if VX does not equal NN
	case 0x04:
		if c.regsV[x] != nn {
			c.pc += 2
			opcodeString = fmt.Sprintf("jump to %04X because v%X != %02X", c.pc, x, nn)
		} else {
			opcodeString = fmt.Sprintf("continue to %04X because v%X == %02X", c.pc, x, nn)
		}

	// 5XY0
	// Skips the next instruction if VX equals VY
	case 0x05:
		switch n {
		case 0x0:
			if c.regsV[x] == c.regsV[y] {
				c.pc += 2
				opcodeString = fmt.Sprintf("jump to %04X because v%X == v%X", c.pc, x, y)
			} else {
				opcodeString = fmt.Sprintf("continue to %04X because v%X != v%X", c.pc, x, y)
			}
		default:
			opcodeString = "undocumented. n must be 0"
			log.Println("n must be 0 for 5XY0 opcode")
		}

	// 6XNN
	// Sets VX to NN
	case 0x06:
		c.regsV[x] = nn
		opcodeString = fmt.Sprintf("v%X = %02X", x, nn)

	// 7XNN
	// Adds NN to VX (carry flag is not changed)
	case 0x07:
		c.regsV[x] += nn
		opcodeString = fmt.Sprintf("v%X += %02X without flags", x, nn)

	case 0x08:
		switch n {

		// 8XY0
		// Sets VX to the value of VY
		case 0x00:
			c.regsV[x] = c.regsV[y]
			opcodeString = fmt.Sprintf("v%X = v%X", x, y)

		// 8XY1
		// Sets VX to VX or VY
		case 0x01:
			c.regsV[x] |= c.regsV[y]
			opcodeString = fmt.Sprintf("v%X |= v%X", x, y)

		// 8XY2
		// Sets VX to VX and VY
		case 0x02:
			c.regsV[x] &= c.regsV[y]
			opcodeString = fmt.Sprintf("v%X &= v%X", x, y)

		// 8XY3
		// Sets VX to VX xor VY
		case 0x03:
			c.regsV[x] ^= c.regsV[y]
			opcodeString = fmt.Sprintf("v%X ^= v%X", x, y)

		// 8XY4
		// Adds VY to VX. VF is set to 1 when there's an overflow, and to 0 when there is not
		case 0x04:
			c.regsV[0xf] = 0
			if math.MaxUint8-c.regsV[x] < c.regsV[y] {
				c.regsV[0xf] = 1
			}
			c.regsV[x] += c.regsV[y]
			opcodeString = fmt.Sprintf("v%X += v%X with flags", x, y)

		// 8XY5
		// VY is subtracted from VX. VF is set to 0 when there's an underflow, and 1 when there is not
		case 0x05:
			c.regsV[0xf] = 0
			if c.regsV[x] >= c.regsV[y] {
				c.regsV[0xf] = 0x1
			}
			c.regsV[x] -= c.regsV[y]

			opcodeString = fmt.Sprintf("v%X -= v%X with flags", x, y)

		// 8XY6
		// If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0.
		// Then Vx is divided by 2.
		case 0x06:
			c.regsV[0xf] = c.regsV[x] & 0x01
			c.regsV[x] >>= 1

			opcodeString = fmt.Sprintf("V%X >>= 1", x)

		// 8XY7
		// Sets VX to VY minus VX. VF is set to 0 when there's an underflow,
		// and 1 when there is not.
		case 0x07:
			c.regsV[0xf] = 0
			if c.regsV[y] >= c.regsV[x] {
				c.regsV[0xf] = 1
			}
			c.regsV[x] = c.regsV[y] - c.regsV[x]

			opcodeString = fmt.Sprintf("V%X = V%X - V%X with flags", x, y, x)
		case 0x0e:
		}

	// ANNN: Store memory address NNN in register I
	//
	// TODO: need to test
	case 0x0a:
		c.regI = nnn
		opcodeString = fmt.Sprintf("vI = %02X", nnn)

	// BNNN: Jump to address NNN + V0
	//
	// TODO: need to test
	case 0x0b:
		c.pc = nnn + uint16(c.regsV[0])

	// CXNN: Set VX to a random number with a mask of NN
	//
	// TODO: need to test
	case 0xc:
		c.regsV[x] = uint8(v2.IntN(0x100)) & nn

	// DXYN: Draw a sprite at position VX, VY with N bytes of sprite data
	// starting at the address stored in I.
	// Set VF to 01 if any set pixels are changed to unset, and 00 otherwise
	//
	// TODO: need to test
	case 0x0d:
		posX := int(c.regsV[x] & (ScreenWidth - 1))
		posY := int(c.regsV[y] & (ScreenHeight - 1))
		c.regsV[0xf] = 0x0

		for i := uint8(0); i < n; i++ {
			spriteData := c.ram[c.regI+uint16(i)]

			posXi := posX
			for j := int8(7); j >= 0; j-- {
				sprPixelOn := spriteData&(1<<j) > 0
				posScreen := posY*ScreenWidth + posXi

				// screen pixel is on and sprite pixel is on, set carry flag
				if sprPixelOn && c.Screen[posScreen] {
					c.regsV[0xf] = 0x1
				}
				c.Screen[posScreen] = c.Screen[posScreen] != sprPixelOn

				posXi++
				if posXi >= ScreenWidth {
					break
				}
			}

			posY++
			if posY >= ScreenHeight {
				break
			}
		}

	default:
		// log.Printf("unimplemented opcode: %04x\n", uint16(opcode))
	}

	fmt.Printf("%04X: %04X %s\n", c.pc, opcode, opcodeString)

	// fmt.Printf("regs:\n")
	// fmt.Printf("     %X-%02X %X-%02X %X-%02X %X-%02X\n",
	// 	0, c.regsV[0], 1, c.regsV[1],
	// 	2, c.regsV[2], 3, c.regsV[3],
	// )
	// fmt.Printf("     %X-%02X %X-%02X %X-%02X %X-%02X\n",
	// 	4, c.regsV[4], 5, c.regsV[5],
	// 	6, c.regsV[6], 7, c.regsV[7],
	// )
	// fmt.Printf("     %X-%02X %X-%02X %X-%02X %X-%02X\n",
	// 	8, c.regsV[8], 9, c.regsV[9],
	// 	0xa, c.regsV[0xa], 0xb, c.regsV[0xb],
	// )
	// fmt.Printf("     %X-%02X %X-%02X %X-%02X %X-%02X\n",
	// 	0xc, c.regsV[0xc], 0xd, c.regsV[0xd],
	// 	0xe, c.regsV[0xe], 0xf, c.regsV[0xf],
	// )
	// stackInfo := "stack: [ "
	// for i := 0; i < int(c.sp); i++ {
	// 	stackInfo += fmt.Sprintf("%04X ", c.stack[i])
	// }
	// stackInfo += "]"
	// fmt.Println(stackInfo)

	// fmt.Println()
}

func (c Chip8) LogRam() {
	for i := 0; i < RamSizeBytes; i += 8 {
		// prevent BCE
		_ = c.ram[i+7]

		if c.ram[i]|
			c.ram[i+1]|
			c.ram[i+2]|
			c.ram[i+3]|
			c.ram[i+4]|
			c.ram[i+5]|
			c.ram[i+6]|
			c.ram[i+7] == 0 {
			continue
		}
		fmt.Printf("%04X: %02x %02x %02x %02x %02x %02x %02x %02x\n",
			i,
			c.ram[i],
			c.ram[i+1],
			c.ram[i+2],
			c.ram[i+3],
			c.ram[i+4],
			c.ram[i+5],
			c.ram[i+6],
			c.ram[i+7],
		)
	}
}

var emptyScreen = make([]bool, ScreenSize)

func (c *Chip8) clearScreen() {
	copy(c.Screen[:], emptyScreen)
}
