package renderer

import (
	"encoding/hex"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/nevisdale/go-chip8/internal/chip8"
)

var keyBindings = map[uint8]ebiten.Key{
	0x0: ebiten.Key0, 0x1: ebiten.Key1, 0x2: ebiten.Key2, 0x3: ebiten.Key3,
	0x4: ebiten.Key4, 0x5: ebiten.Key5, 0x6: ebiten.Key6, 0x7: ebiten.Key7,
	0x8: ebiten.Key8, 0x9: ebiten.Key9, 0xa: ebiten.KeyA, 0xb: ebiten.KeyB,
	0xc: ebiten.KeyC, 0xd: ebiten.KeyD, 0xe: ebiten.KeyE, 0xf: ebiten.KeyF,
}

type Config struct {
	FgColor color.Color
	BgColor color.Color
}

type Renderer struct {
	chip8 *chip8.Chip8

	fgColor color.Color
	bgColor color.Color
}

func NewFromConfig(chip8 *chip8.Chip8, conf Config) *Renderer {
	return &Renderer{
		chip8: chip8,

		fgColor: conf.FgColor,
		bgColor: conf.BgColor,
	}
}

func (r *Renderer) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	for chip8Key, ebitenKey := range keyBindings {
		r.chip8.SetKey(chip8Key, ebiten.IsKeyPressed(ebitenKey))
	}
	r.chip8.Emulate()

	return nil
}

func (r *Renderer) Draw(screen *ebiten.Image) {
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

func (r *Renderer) Layout(int, int) (int, int) {
	w, h := r.chip8.ScreenSize()
	return w, h
}

func (r *Renderer) Run() error {
	ebiten.SetTPS(r.chip8.GetTPS())
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("CHIP8 Emulator: " + r.chip8.GetRomName())

	if err := ebiten.RunGame(r); err != nil {
		return fmt.Errorf("run renderer: %w", err)
	}
	return nil
}

func DecodeColorFromHex(s string) (color.Color, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode a hex string: %w", err)
	}
	if len(data) != 3 && len(data) != 4 {
		return nil, fmt.Errorf("color must be in rgb or rgba format")
	}

	c := color.RGBA{
		R: data[0],
		G: data[1],
		B: data[2],
		A: 0xff,
	}
	if len(data) == 4 {
		c.A = data[3]
	}

	return c, nil
}
