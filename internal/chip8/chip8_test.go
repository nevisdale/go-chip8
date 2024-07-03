package chip8

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChip8_Emulate(t *testing.T) {
	t.Parallel()

	t.Run("00E0", func(t *testing.T) {
		rom := Rom{
			Data: []byte{
				0x00, 0xe0, // clear screen
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		// dirty screen
		for i := 0; i < ScreenSize; i++ {
			chip8.Screen[i] = true
		}

		chip8.Emulate()

		for i := 0; i < ScreenSize; i++ {
			require.False(t, chip8.Screen[i])
		}
	})

	t.Run("1NNN", func(t *testing.T) {
		rom := Rom{
			Data: []byte{
				0x1c, 0xfe, // // jump to 0xcfe
			},
		}
		var expectedPc uint16 = 0x0cfe

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate()

		require.Equal(t, expectedPc, chip8.pc)
	})

	t.Run("2NNN_00EE", func(t *testing.T) {
		var expectedV0 uint8 = 0x78

		rom := Rom{
			Data: []byte{
				0x22, 0x04, // 0x200: go to 0x204
				0x00, 0xe0, // 0x202: clear screen
				0x60, 0x78, // 0x204: v[0] = 0x78
				0x00, 0xee, // 0x206: return to 0x202
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)
		chip8.Screen[0] = true

		chip8.Emulate() // go to 0x204
		chip8.Emulate() // v[0] = 0x78
		require.Equal(t, expectedV0, chip8.regsV[0], "reg v0")
		require.True(t, chip8.Screen[0], "screen")

		chip8.Emulate() // return to 0x202
		chip8.Emulate() // clear screen
		require.False(t, chip8.Screen[0], "screen")
	})

	t.Run("3XNN", func(t *testing.T) {
		var expectedV0 uint8 = 0x11

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x30, 0x11, // if v[0] == 0x11 then skip the next instruction
				0x60, 0x12, // v[0] = 0x12
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate() // v[0] = 0x11
		require.Equal(t, expectedV0, chip8.regsV[0])

		chip8.Emulate() // checking v[0] == 0x11 and skip the next instruction
		chip8.Emulate() // do nothing, just to check

		require.Equal(t, expectedV0, chip8.regsV[0])
	})

	t.Run("4XNN", func(t *testing.T) {
		var expectedV0 uint8 = 0x11

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x40, 0x12, // if v[0] != 0x12 then skip the next instruction
				0x60, 0x12, // v[0] = 0x12
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate() // v[0] = 0x11
		require.Equal(t, expectedV0, chip8.regsV[0])

		chip8.Emulate() // checking v[0] != 0x12 and skip the next instruction
		chip8.Emulate() // do nothing, just to check

		require.Equal(t, expectedV0, chip8.regsV[0])
	})

	t.Run("5XY0", func(t *testing.T) {
		var expectedV0 uint8 = 0x11
		var expectedV1 uint8 = 0x11

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x61, 0x11, // v[1] = 0x11
				0x50, 0x10, // if v[0] == v[1] then skip the next instruction
				0x60, 0x12, // v[0] = 0x12
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate() // v[0] = 0x11
		require.Equal(t, expectedV0, chip8.regsV[0])

		chip8.Emulate() // v[1] = 0x11
		require.Equal(t, expectedV1, chip8.regsV[1])

		chip8.Emulate() // checking v[0] == v[1] and skip the next instruction
		chip8.Emulate() // do nothing, just to check

		require.Equal(t, expectedV0, chip8.regsV[0])
		require.Equal(t, expectedV1, chip8.regsV[1])
	})

	t.Run("6XNN", func(t *testing.T) {
		var expectedV0_1 uint8 = 0x11
		var expectedV0_2 uint8 = 0x14

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x60, 0x14, // v[0] = 0x14
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate() // v[0] = 0x11
		require.Equal(t, expectedV0_1, chip8.regsV[0])

		chip8.Emulate() // v[0] = 0x14
		require.Equal(t, expectedV0_2, chip8.regsV[0])
	})

	t.Run("7XNN", func(t *testing.T) {
		var expectedV0 uint8 = 0x11
		var expectedVf uint8 = 0x0

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x70, 0x03, // v[0] += 0x03
				0x70, 0xff, // v[0] += 0xff (do not set v[f])
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate() // v[0] = 0x11
		require.Equal(t, expectedV0, chip8.regsV[0])

		chip8.Emulate() // v[0] += 0x03
		expectedV0 += 0x03
		require.Equal(t, expectedV0, chip8.regsV[0])
		require.Equal(t, expectedVf, chip8.regsV[0xf], "v[f]")

		chip8.Emulate() // v[0] += 0xff (do not set v[f])
		expectedV0 += 0xff
		require.Equal(t, expectedV0, chip8.regsV[0])
		require.Equal(t, expectedVf, chip8.regsV[0xf], "v[f]")
	})

	t.Run("8XY0", func(t *testing.T) {
		var expectedV0 uint8 = 0x11
		var expectedV1 uint8 = 0x14

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x61, 0x14, // v[1] = 0x14
				0x80, 0x10, // v[0] = v[1]
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate() // v[0] = 0x11
		require.Equal(t, expectedV0, chip8.regsV[0])

		chip8.Emulate() // v[1] = 0x14
		require.Equal(t, expectedV1, chip8.regsV[1])

		chip8.Emulate() // v[0] = v[1]
		require.Equal(t, expectedV1, chip8.regsV[0])
		require.Equal(t, expectedV1, chip8.regsV[1])
	})

	t.Run("8XY1", func(t *testing.T) {
		var expectedV0 uint8 = 0x11
		var expectedV1 uint8 = 0x14

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x61, 0x14, // v[1] = 0x14
				0x80, 0x11, // v[0] |= v[1]
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate() // v[0] = 0x11
		require.Equal(t, expectedV0, chip8.regsV[0])

		chip8.Emulate() // v[1] = 0x14
		require.Equal(t, expectedV1, chip8.regsV[1])

		chip8.Emulate() // v[0] |= v[1]
		expectedV0 |= expectedV1
		require.Equal(t, expectedV0, chip8.regsV[0])
		require.Equal(t, expectedV1, chip8.regsV[1])
	})

	t.Run("8XY2", func(t *testing.T) {
		var expectedV0 uint8 = 0x11
		var expectedV1 uint8 = 0x14

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x61, 0x14, // v[1] = 0x14
				0x80, 0x12, // v[0] &= v[1]
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate() // v[0] = 0x11
		require.Equal(t, expectedV0, chip8.regsV[0])

		chip8.Emulate() // v[1] += 0x14
		require.Equal(t, expectedV1, chip8.regsV[1])

		chip8.Emulate() // v[0] &= v[1]
		expectedV0 &= expectedV1
		require.Equal(t, expectedV0, chip8.regsV[0])
		require.Equal(t, expectedV1, chip8.regsV[1])
	})

	t.Run("8XY3", func(t *testing.T) {
		var expectedV0 uint8 = 0x11
		var expectedV1 uint8 = 0x14

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x61, 0x14, // v[1] = 0x14
				0x80, 0x13, // v[0] ^= v[1]
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate() // v[0] = 0x11
		require.Equal(t, expectedV0, chip8.regsV[0])

		chip8.Emulate() // v[1] += 0x14
		require.Equal(t, expectedV1, chip8.regsV[1])

		chip8.Emulate() // v[0] ^= v[1]
		expectedV0 ^= expectedV1
		require.Equal(t, expectedV0, chip8.regsV[0])
		require.Equal(t, expectedV1, chip8.regsV[1])
	})

	t.Run("8XY4", func(t *testing.T) {
		var expectedV0 uint8 = 0x11
		var expectedV1 uint8 = 0x14

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x61, 0x14, // v[1] = 0x14
				0x80, 0x14, // v[0] += v[1] (v[f] = 0)
				0x61, 0xff, // v[1] = 0xff
				0x80, 0x14, // v[0] += v[1] (v[f] = 1)
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate()
		chip8.Emulate()
		chip8.Emulate()

		expectedV0 += expectedV1
		require.Equal(t, expectedV0, chip8.regsV[0])
		require.Equal(t, expectedV1, chip8.regsV[1])
		require.Equal(t, uint8(0), chip8.regsV[0xf])

		chip8.Emulate()
		chip8.Emulate()
		expectedV0 += 0xff
		require.Equal(t, expectedV0, chip8.regsV[0])
		require.Equal(t, uint8(1), chip8.regsV[0xf])
	})

	t.Run("8XY5", func(t *testing.T) {
		var expectedV0 uint8 = 0x11
		var expectedV1 uint8 = 0x14

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x61, 0x14, // v[1] = 0x14
				0x80, 0x15, // v[0] = v[0] - v[1] (v[f] = 0)
				0x60, 0x11, // v[0] = 0x11
				0x81, 0x05, // v[1] = v[1] - v[0] (v[f] = 1)
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate()
		chip8.Emulate()

		chip8.Emulate()
		subV0 := expectedV0 - expectedV1
		require.Equal(t, subV0, chip8.regsV[0])
		require.Equal(t, expectedV1, chip8.regsV[1])
		require.Equal(t, uint8(0), chip8.regsV[0xf])

		chip8.Emulate()

		chip8.Emulate()
		subV1 := expectedV1 - expectedV0
		require.Equal(t, subV1, chip8.regsV[1])
		require.Equal(t, expectedV0, chip8.regsV[0])
		require.Equal(t, uint8(1), chip8.regsV[0xf])
	})

	t.Run("8XY6", func(t *testing.T) {
		var expectedV0 uint8 = 0x11 // 0b10001

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x80, 0x16, // v[f] = v[0] & 0x1; v[0] >>= 1
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate()
		chip8.Emulate()

		require.Equal(t, expectedV0&0x01, chip8.regsV[0xf])
		require.Equal(t, expectedV0>>1, chip8.regsV[0])
	})

	t.Run("8XY7", func(t *testing.T) {
		var expectedV0 uint8 = 0x11
		var expectedV1 uint8 = 0x14

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x61, 0x14, // v[1] = 0x14
				0x80, 0x17, // v[0] = v[1] - v[0] (v[f] = 1)
				0x60, 0x11, // v[0] = 0x11
				0x81, 0x07, // v[1] = v[0] - v[1] (v[f] = 0)
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate()
		chip8.Emulate()

		chip8.Emulate()
		subV0 := expectedV1 - expectedV0
		require.Equal(t, subV0, chip8.regsV[0])
		require.Equal(t, expectedV1, chip8.regsV[1])
		require.Equal(t, uint8(1), chip8.regsV[0xf])

		chip8.Emulate()

		chip8.Emulate()
		subV1 := expectedV0 - expectedV1
		require.Equal(t, subV1, chip8.regsV[1])
		require.Equal(t, expectedV0, chip8.regsV[0])
		require.Equal(t, uint8(0), chip8.regsV[0xf])
	})

	t.Run("8XYE", func(t *testing.T) {
		var expectedV0 uint8 = 0x11 // 0001 0001

		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x80, 0x1e, // v[f] = 0x0; v[0] <<= 1

				0x60, 0x82, // v[0] = 0x82
				0x80, 0x1e, // v[f] = 0x1; v[0] <<= 1
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate()
		chip8.Emulate()

		require.Equal(t, uint8(0), chip8.regsV[0xf])
		require.Equal(t, expectedV0<<1, chip8.regsV[0])

		chip8.Emulate()
		chip8.Emulate()
		expectedV0 = 0x82

		require.Equal(t, uint8(1), chip8.regsV[0xf])
		require.Equal(t, expectedV0<<1, chip8.regsV[0])
	})

	t.Run("9XY0", func(t *testing.T) {
		rom := Rom{
			Data: []byte{
				0x60, 0x11, // v[0] = 0x11
				0x61, 0x14, // v[1] = 0x14
				0x90, 0x10, // if v[0] != v[1] then skip the next instruction
				0x00, 0xe0, // clear screen
				0x00, 0x00, // do nothing

				0x80, 0x10, // v[0] = v[1]
				0x90, 0x10, // if v[0] != v[1] then skip the next instruction
				0x00, 0xe0, // clear screen
				0x00, 0x00, // do nothing
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Screen[0] = true

		chip8.Emulate()
		chip8.Emulate()
		chip8.Emulate()
		chip8.Emulate()
		require.True(t, chip8.Screen[0])

		chip8.Emulate()
		chip8.Emulate()
		chip8.Emulate()
		require.False(t, chip8.Screen[0])
	})

	t.Run("ANNN", func(t *testing.T) {
		expectedVI := uint16(0x189)

		rom := Rom{
			Data: []byte{
				0xa1, 0x89, // regI = 0x189
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate()
		require.Equal(t, expectedVI, chip8.regI)
	})

	t.Run("BNNN", func(t *testing.T) {
		rom := Rom{
			Data: []byte{
				0x60, 0x06, // 0x200: v[0] = 0x06
				0xb2, 0x00, // 0x202: jump to v[0] + 0x200 (0x206)
				0x00, 0xe0, // 0x204: clear screen
				0x00, 0x00, // do nothing

			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Screen[0] = true

		chip8.Emulate()
		chip8.Emulate()
		chip8.Emulate()

		require.True(t, chip8.Screen[0])
	})

	t.Run("CXNN", func(t *testing.T) {
		expectedNN := uint8(0x67)

		rom := Rom{
			Data: []byte{
				0xc0, 0x67, // v[0] = rand() & 0x67
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.Emulate()
		require.GreaterOrEqual(t, expectedNN, chip8.regsV[0])
	})

	t.Run("EX9E", func(t *testing.T) {
		rom := Rom{
			Data: []byte{
				0xe0, 0x9e, // if keypad[v[0]] == pressed then skip the next instruction
				0x00, 0xe0, // clear screen
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.KeyPad[0] = true
		chip8.Screen[0] = true

		chip8.Emulate()
		chip8.Emulate()

		require.True(t, chip8.Screen[0])
	})

	t.Run("EXA1", func(t *testing.T) {
		rom := Rom{
			Data: []byte{
				0xe0, 0xa1, // if keypad[v[0]] != pressed then skip the next instruction
				0x00, 0xe0, // clear screen
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.KeyPad[0] = false
		chip8.Screen[0] = true

		chip8.Emulate()
		chip8.Emulate()

		require.True(t, chip8.Screen[0])
	})

	t.Run("FX07", func(t *testing.T) {
		expectedDelayTimer := uint8(8)

		rom := Rom{
			Data: []byte{
				0xf0, 0x07, // v[0] = delay timer
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)

		chip8.delayTimer = expectedDelayTimer
		chip8.Emulate()

		require.Equal(t, expectedDelayTimer, chip8.regsV[0])
	})

	t.Run("FX15", func(t *testing.T) {
		expectedDelayTimer := uint8(0x8)

		rom := Rom{
			Data: []byte{
				0x60, 0x08, // v[0] = 0x8
				0xf0, 0x15, // delay timer = v[0]
			},
		}

		chip8 := NewChip8()
		chip8.LoadRom(rom)
		chip8.Emulate()
		chip8.Emulate()

		// delay timer decreases every tick even after setting a value
		expectedDelayTimer--
		require.Equal(t, expectedDelayTimer, chip8.delayTimer)
	})
}
