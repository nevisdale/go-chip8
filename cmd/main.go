package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"image/color"
	"os"

	"github.com/nevisdale/go-chip8/internal/chip8"
	"github.com/nevisdale/go-chip8/internal/renderer"
)

var (
	romPath    string
	fgColorHex string
	bgColorHex string
)

func main() {
	flag.StringVar(&romPath, "f", "", "rom file. is required")
	flag.StringVar(&fgColorHex, "fg", "FFFFFFFF", "rgba foreground color in hex. white is default")
	flag.StringVar(&bgColorHex, "bg", "000000FF", "rgba background color in hex. black is default")
	flag.Parse()

	if len(romPath) == 0 {
		// FOR DEBUG
		if testRomName := os.Getenv("TEST_ROM"); len(testRomName) > 0 {
			romPath = testRomName
		} else {
			fmt.Fprintf(os.Stderr, "rom file is empty\n")
			os.Exit(1)
		}
	}

	fgColor, err := decodeColorFromHex(fgColorHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't decode fg color from hex %s: %s\n", fgColorHex, err.Error())
		os.Exit(1)
	}
	bgColor, err := decodeColorFromHex(bgColorHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't decode bg color from hex %s: %s\n", bgColorHex, err.Error())
		os.Exit(1)
	}

	rom, err := chip8.NewRomFromFile(romPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't creare a rom from the file: %s\n", err.Error())
		os.Exit(1)
	}

	chip8 := chip8.NewChip8()
	chip8.LoadRom(rom)
	chip8.LogRam()
	fmt.Println()

	renderer := renderer.NewFromConfig(&chip8, renderer.Config{
		FgColor: fgColor,
		BgColor: bgColor,
	})
	if err := renderer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "couldn't run a renderer: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

// TODO: decode from 3-components rgba
func decodeColorFromHex(s string) (color.Color, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode a hex string: %w", err)
	}

	return color.RGBA{
		R: data[0],
		G: data[1],
		B: data[2],
		A: data[3],
	}, nil
}
