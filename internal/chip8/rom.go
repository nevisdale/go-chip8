package chip8

import (
	"fmt"
	"os"
	"path"
)

type Rom struct {
	Name string
	Data []byte
}

func NewRomFromFile(romPath string) (Rom, error) {
	data, err := os.ReadFile(romPath)
	if err != nil {
		return Rom{}, fmt.Errorf("read data from rom file %s: %w", romPath, err)
	}

	if len(data) > RomMaxSizeBytes {
		return Rom{}, fmt.Errorf("rom file %s is too large. actual size is %d bytes, max size is %d bytes",
			romPath, len(data), RomMaxSizeBytes,
		)
	}

	return Rom{
		Name: path.Base(romPath),
		Data: data,
	}, nil
}
