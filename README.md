# go-chip8
CHIP8 Emulator in go

## Build
```bash
make build
```

## Keypad:
```sh
// ====================
// keyboard key mapping
// ====================
//
//	1 2 3 C  -> 1 2 3 4
//	4 5 6 D  -> Q W E R
//	7 8 9 E  -> A S D F
//	A 0 B F  -> Z X C V
```

## Run roms:
### 1. IBM Logo:
```bash
./bin/chip8 -f ./roms/IBM_Logo.ch8
```

### 2. Test opcodes:
```bash
./bin/chip8 -f ./roms/test_opcode.ch8
```

### 3. More roms:
- [kripod/chip8-roms](https://github.com/kripod/chip8-roms)

## Special keys:
- P - pause/play a game
- K - show/hide a keypad window
