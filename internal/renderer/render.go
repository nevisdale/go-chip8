package renderer

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nevisdale/go-chip8/internal/chip8"
)

type Config struct {
	FgColor color.Color
	BgColor color.Color
}

type Renderer struct {
	chip8 *chip8.Chip8

	fgColor color.Color
	bgColor color.Color
}

func NewFromConfig(chip8 *chip8.Chip8, conf Config) Renderer {
	return Renderer{
		chip8: chip8,

		fgColor: conf.FgColor,
		bgColor: conf.BgColor,
	}
}

func (r Renderer) Update() error {
	r.chip8.Emulate()
	return nil
}

func (r Renderer) Draw(screen *ebiten.Image) {
	screen.Fill(r.bgColor)

	w, _ := r.chip8.ScreenSize()
	for i := 0; i < len(r.chip8.Screen); i++ {
		if !r.chip8.Screen[i] {
			continue
		}

		posX := i % w
		posY := i / w
		screen.Set(posX, posY, r.fgColor)
	}
}

func (r Renderer) Layout(int, int) (int, int) {
	w, h := r.chip8.ScreenSize()
	return w, h
}

func (r Renderer) Run() error {
	ebiten.SetTPS(r.chip8.GetTPS())
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("CHIP8 Emulator: " + r.chip8.GetRomBaseName())

	if err := ebiten.RunGame(r); err != nil {
		return fmt.Errorf("run renderer: %w", err)
	}
	return nil
}
