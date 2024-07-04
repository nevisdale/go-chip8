package beep

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	sampleRate = 44100
	beepHz     = 440
	duration   = time.Second

	volumeStep = 0.2
	volumeMax  = 1.0
	volumeMin  = 0.0
)

type Beep struct {
	p *audio.Player
}

func New() (*Beep, error) {
	numSamples := sampleRate * int(duration.Seconds())
	buf := make([]byte, numSamples*2)
	for i := 0; i < numSamples; i++ {
		a := math.Sin(2.0 * math.Pi * float64(beepHz) * float64(i) / float64(sampleRate))
		s := int16(a * math.MaxInt16)
		buf[2*i] = byte(s)
		buf[2*i+1] = byte(s >> 8)
	}

	audioCtx := audio.NewContext(sampleRate)
	player, err := audioCtx.NewPlayer(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("couldn't create an audio player: %w", err)
	}

	return &Beep{
		p: player,
	}, nil
}

func (b *Beep) Play() {
	if err := b.p.Rewind(); err != nil {
		log.Printf("couldn't rewind the audio player: %s\n", err.Error())
		return
	}
	b.p.Play()
}

func (b *Beep) VolumeUp() {
	volume := b.p.Volume()
	volume = min(volume+volumeStep, volumeMax)
	b.p.SetVolume(volume)
}

func (b *Beep) VolumeDown() {
	volume := b.p.Volume()
	volume = max(volume-volumeStep, volumeMin)
	b.p.SetVolume(volume)
}

func (b *Beep) SetVolume(volume float64) {
	volume = min(volume, volumeMax)
	volume = max(volume, volumeMin)
	b.p.SetVolume(volume)
}
