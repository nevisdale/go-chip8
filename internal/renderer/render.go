package renderer

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/nevisdale/go-chip8/internal/chip8"
)

// ====================
// keyboard key mapping
// ====================
//
//	1 2 3 C  -> 1 2 3 4
//	4 5 6 D  -> Q W E R
//	7 8 9 E  -> A S D F
//	A 0 B F  -> Z X C V
var keyboardMapping = map[uint8]ebiten.Key{
	0x1: ebiten.Key1, 0x2: ebiten.Key2, 0x3: ebiten.Key3, 0xC: ebiten.Key4,
	0x4: ebiten.KeyQ, 0x5: ebiten.KeyW, 0x6: ebiten.KeyE, 0xD: ebiten.KeyR,
	0x7: ebiten.KeyA, 0x8: ebiten.KeyS, 0x9: ebiten.KeyD, 0xE: ebiten.KeyF,
	0xA: ebiten.KeyZ, 0x0: ebiten.KeyX, 0xB: ebiten.KeyC, 0xF: ebiten.KeyV,
}

var keyboardPosition = map[uint8]uint8{
	0x0: 0x1, 0x1: 0x2, 0x2: 0x3, 0x3: 0xC,
	0x4: 0x4, 0x5: 0x5, 0x6: 0x6, 0x7: 0xD,
	0x8: 0x7, 0x9: 0x8, 0xA: 0x9, 0xB: 0xE,
	0xC: 0xA, 0xD: 0x0, 0xE: 0xB, 0xF: 0xF,
}

var (
	buttonReleasedColor color.Color = MustDecodeColorFromHex("999999")
	buttonPressedColor  color.Color = MustDecodeColorFromHex("65f057")
)

type Config struct {
	FgColor color.Color
	BgColor color.Color
}

type Renderer struct {
	chip8 *chip8.Chip8

	fgColor color.Color
	bgColor color.Color

	keypadMode bool
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

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		r.chip8.TogglePause()
		r.setWindowTitle()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		r.keypadMode = !r.keypadMode
	}

	switch {
	case inpututil.IsKeyJustPressed(ebiten.Key0):
		r.chip8.SoundVolumeUp()
	case inpututil.IsKeyJustPressed(ebiten.Key9):
		r.chip8.SoundVolumeDown()
	}

	for chip8Key, ebitenKey := range keyboardMapping {
		r.chip8.SetKey(chip8Key, ebiten.IsKeyPressed(ebitenKey))
	}
	r.chip8.Emulate()

	return nil
}

func (r *Renderer) Draw(screen *ebiten.Image) {
	// CHIP8 screen
	chip8ScreenOffsetX := 0
	chip8ScreenOffsetY := 0
	for x := 0; x < r.chip8.ScreenWidth(); x++ {
		for y := 0; y < r.chip8.ScreenHeight(); y++ {
			pixelColor := r.bgColor
			if r.chip8.ScreenPixelSetAt(x, y) {
				pixelColor = r.fgColor
			}

			screen.Set(chip8ScreenOffsetX+x, chip8ScreenOffsetY+y, pixelColor)
		}
	}

	// Keypad screen
	if r.keypadMode {
		buttonsInRow := 4
		buttonSize := 4

		// center by X
		screenOffsetX := chip8ScreenOffsetX + (r.chip8.ScreenWidth()-(buttonsInRow*buttonSize+buttonsInRow-1))>>1
		screenOffsetY := chip8ScreenOffsetY + r.chip8.ScreenHeight() + 1

		for x := 0; x < 4; x++ {
			for y := 0; y < 4; y++ {
				pixelColor := buttonReleasedColor
				key := y<<2 | x&0xf
				if r.chip8.KeyIsPressed(keyboardPosition[uint8(key)]) {
					pixelColor = buttonPressedColor
				}

				posX := screenOffsetX + (x * (buttonSize + 1))
				posY := screenOffsetY + (y * (buttonSize + 1))

				vector.DrawFilledRect(screen,
					float32(posX),
					float32(posY),
					float32(buttonSize),
					float32(buttonSize),
					pixelColor, false,
				)
			}
		}

	}
}

func (r *Renderer) Layout(int, int) (int, int) {
	w, h := r.chip8.ScreenSize()
	if r.keypadMode {
		return w, h + 22
	}
	return w, h
}

func (r *Renderer) Run() error {
	ebiten.SetTPS(r.chip8.GetTPS())
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	r.setWindowTitle()

	if err := ebiten.RunGame(r); err != nil {
		return fmt.Errorf("run renderer: %w", err)
	}
	return nil
}

func (r *Renderer) setWindowTitle() {
	ebiten.SetWindowTitle("CHIP8 Emulator: " + r.chip8.GetRomName() + " " + r.chip8.GetState().String())
}

func MustDecodeColorFromHex(s string) color.Color {
	color, err := DecodeColorFromHex(s)
	if err != nil {
		log.Fatal(err.Error())
	}
	return color
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
