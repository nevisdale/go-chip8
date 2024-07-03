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

	// // http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#2.3
	KeyPadSize = 0x10

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

	KeyPad [KeyPadSize]bool

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

	delayTimer uint8
	soundTimer uint8
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
		howMuchToSleep := (time.Second / DefaultTPS) - elapsed
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

		// 8XYE
		// Shifts VX to the left by 1,
		// then sets VF to 1 if the most significant bit of VX prior to that shift was set,
		// or to 0 if it was unset
		case 0x0e:
			c.regsV[0xf] = 0
			if c.regsV[x]&0x80 > 0 {
				c.regsV[0xf] = 1
			}
			c.regsV[x] <<= 1

			opcodeString = fmt.Sprintf("V%X <<= 1", x)
		}

	case 0x09:
		switch n {

		// 9XY0
		// Skips the next instruction if VX does not equal VY
		case 0x0:
			if c.regsV[x] != c.regsV[y] {
				c.pc += 2
			}

			opcodeString = fmt.Sprintf("if V%X != V%X", x, y)

		default:
			opcodeString = "undocumented. n must be 0"
			log.Println("n must be 0 for 9XY0 opcode")
		}

	// ANNN
	// Sets I to the address NNN
	case 0x0a:
		c.regI = nnn

		opcodeString = fmt.Sprintf("vI = %02X", nnn)

	// BNNN
	// Jumps to the address NNN plus V0
	case 0x0b:
		c.pc = nnn + uint16(c.regsV[0])

		opcodeString = fmt.Sprintf("jump to %02X", c.pc)

	// CXNN
	// Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN
	case 0xc:
		c.regsV[x] = uint8(v2.IntN(0x100)) & nn

		opcodeString = fmt.Sprintf("V%X = rnd() & %02X", x, nn)

	// DXYN
	// Draws a sprite at coordinate (VX, VY) that has a width of 8 pixels and a height of N pixels.
	// Each row of 8 pixels is read as bit-coded starting from memory location I;
	// I value does not change after the execution of this instruction.
	// As described above, VF is set to 1 if any screen pixels are flipped from set to unset when the sprite is drawn,
	// and to 0 if that does not happen.
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

		opcodeString = fmt.Sprintf("draw(%X, %X, %X)", x, y, n)

	case 0x0e:
		switch nn {

		// EX9E
		// Skips the next instruction if the key stored in VX is pressed
		case 0x9e:
			if c.regsV[x] < 0x10 && c.KeyPad[c.regsV[x]] {
				c.pc += 2
			}

			opcodeString = fmt.Sprintf("if keypad[%X] pressed than skip the next", c.regsV[x])

		// EXA1
		// Skips the next instruction if the key stored in VX is not pressed
		case 0xa1:
			if c.regsV[x] < 0x10 && !c.KeyPad[c.regsV[x]] {
				c.pc += 2
			}

			opcodeString = fmt.Sprintf("if keypad[%X] not pressed than skip the next", c.regsV[x])

		default:
			opcodeString = fmt.Sprintf("unknown opcode %04X", opcode)
			log.Println(opcodeString)
		}

	case 0x0f:
		switch nn {

		// FX07
		// Sets VX to the value of the delay timer
		case 0x07:
			c.regsV[x] = c.delayTimer

			opcodeString = fmt.Sprintf("V%X = delay timer(%X)", x, c.delayTimer)

		// FX0A
		// A key press is awaited, and then stored in VX
		// (blocking operation, all instruction halted until next key event)
		case 0x0a:
			var i uint8
		outer:

			for i = uint8(0); i < KeyPadSize; i++ {
				if c.KeyPad[i] {
					c.regsV[x] = i
					break outer
				}
			}

			if i == KeyPadSize {
				c.pc -= 2
				log.Println("waiting to press")
				return
			}

			opcodeString = fmt.Sprintf("%X is pressed", i)

		// FX15
		// Sets the delay timer to VX
		case 0x15:
			c.delayTimer = c.regsV[x]

			opcodeString = fmt.Sprintf("delay timer = V%X", c.regsV[x])

		// FX18
		// Sets the sound timer to VX
		case 0x18:
			c.soundTimer = c.regsV[x]

			opcodeString = fmt.Sprintf("sound timer = V%X", c.regsV[x])

		// FX1E
		// Adds VX to I. VF is not affected
		case 0x1e:
			c.regI += uint16(c.regsV[x])

			opcodeString = fmt.Sprintf("VI += V%X", x)

		// FX29
		// Sets I to the location of the sprite for the character in VX
		case 0x29:
			c.regI = uint16(c.regsV[x]) * 5

			opcodeString = fmt.Sprintf("VI set to %X font sprite", c.regsV[x])

		// FX33
		// Stores the binary-coded decimal representation of VX,
		// with the hundreds digit in memory at location in I,
		// the tens digit at location I+1,
		// and the ones digit at location I+2
		case 0x33:
			c100 := c.regsV[x] / 100
			c10 := (c.regsV[x] - c100*100) / 10
			c1 := c.regsV[x] - c100*100 - c10*10

			c.ram[c.regI] = c100
			c.ram[c.regI+1] = c10
			c.ram[c.regI+2] = c1

			opcodeString = fmt.Sprintf("BCD(V%X)", x)

		// FX55
		// Stores from V0 to VX (including VX) in memory, starting at address I.
		// The offset from I is increased by 1 for each value written,
		// but I itself is left unmodified
		case 0x55:
			for i := uint16(0); i <= uint16(x); i++ {
				c.ram[c.regI+i] = c.regsV[i]
			}

			opcodeString = fmt.Sprintf("store from V0 to V%X", x)

		// FX65
		// Fills from V0 to VX (including VX) with values from memory, starting at address I.
		// The offset from I is increased by 1 for each value read,
		// but I itself is left unmodified
		case 0x65:
			for i := uint16(0); i <= uint16(x); i++ {
				c.regsV[i] = c.ram[c.regI+i]
			}

			opcodeString = fmt.Sprintf("decode from RAM to V0 to V%X", x)

		default:
			opcodeString = fmt.Sprintf("unknown opcode %04X", opcode)
		}

	default:
		opcodeString = fmt.Sprintf("unimplemented opcode: %04x", opcode)

	}

	if c.delayTimer > 0 {
		c.delayTimer--
	}
	if c.soundTimer > 0 {
		if c.soundTimer == 1 {
			log.Println("PLAY SOUND")
		}
		c.soundTimer--
	}

	fmt.Printf("%04X: %04X %s\n", c.pc, opcode, opcodeString)
}

var emptyScreen = make([]bool, ScreenSize)

func (c *Chip8) clearScreen() {
	copy(c.Screen[:], emptyScreen)
}

func (c *Chip8) SetKey(key uint8, isPressed bool) {
	if key >= KeyPadSize {
		log.Println("key is invalid. do nothing")
		return
	}
	c.KeyPad[key] = isPressed
}
