package main

import (
	"flag"
	"fmt"
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
		fmt.Fprintf(os.Stderr, "rom file is empty\n")
		os.Exit(1)
	}

	fgColor, err := renderer.DecodeColorFromHex(fgColorHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't decode fg color from hex %s: %s\n", fgColorHex, err.Error())
		os.Exit(1)
	}
	bgColor, err := renderer.DecodeColorFromHex(bgColorHex)
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
