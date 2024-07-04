package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nevisdale/go-chip8/internal/beep"
	"github.com/nevisdale/go-chip8/internal/chip8"
	"github.com/nevisdale/go-chip8/internal/renderer"
)

var (
	soundVolume float64
	romPath     string
	fgColorHex  string
	bgColorHex  string
	tps         int
)

func main() {
	flag.StringVar(&romPath, "f", "", "rom file. is required")
	flag.StringVar(&fgColorHex, "fg", "FFFFFFFF", "rgba foreground color in hex. white is default")
	flag.StringVar(&bgColorHex, "bg", "000000FF", "rgba background color in hex. black is default")
	flag.IntVar(&tps, "tps", 60, "tps")
	flag.Float64Var(&soundVolume, "volume", 0.5, "sound volume. must be between 0 and 1")
	flag.Parse()

	if len(romPath) == 0 {
		fmt.Fprintf(os.Stderr, "rom file is empty\n")
		os.Exit(1)
	}

	if soundVolume < 0 || soundVolume > 1 {
		fmt.Fprintf(os.Stderr, "sound volume is invalid, must be between 0 and 1")
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

	beepPlayer, err := beep.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "beep player: %s\n", err.Error())
		os.Exit(1)
	}
	beepPlayer.SetVolume(soundVolume)

	chip8 := chip8.NewChip8()
	chip8.LoadRom(rom)
	chip8.SetTPS(tps)
	chip8.SetSoundPlayer(beepPlayer)

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
