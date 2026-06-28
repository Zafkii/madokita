package scene

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

const DefaultFadeDuration = 0.5

type FadeState int

const (
	FadeNone FadeState = iota
	FadeOut
	FadeIn
)

type Fade struct {
	mu       sync.Mutex
	State    FadeState
	Progress float64
	Duration float64
	Color    color.RGBA
	overlay  *ebiten.Image
}

func NewFade(duration float64, c color.RGBA) *Fade {
	return &Fade{
		Duration: duration,
		Color:    c,
	}
}

func (f *Fade) StartOut() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.State = FadeOut
	f.Progress = 0
}

func (f *Fade) StartIn() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.State = FadeIn
	f.Progress = 0
}

func (f *Fade) Update(dt float64) (done bool, state FadeState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.State == FadeNone {
		return true, FadeNone
	}
	f.Progress += dt / f.Duration
	if f.Progress >= 1.0 {
		f.Progress = 1.0
		f.State = FadeNone
		return true, FadeNone
	}
	return false, f.State
}

func (f *Fade) alpha() float64 {
	switch f.State {
	case FadeOut:
		return f.Progress
	case FadeIn:
		return 1.0 - f.Progress
	default:
		return 0
	}
}

func (f *Fade) Draw(screen *ebiten.Image) {
	f.mu.Lock()
	alpha := f.alpha()
	if alpha <= 0 {
		f.mu.Unlock()
		return
	}
	dw, dh := screen.Bounds().Dx(), screen.Bounds().Dy()
	if dw <= 0 || dh <= 0 {
		f.mu.Unlock()
		return
	}
	if f.overlay == nil || f.overlay.Bounds().Dx() != dw || f.overlay.Bounds().Dy() != dh {
		f.overlay = ebiten.NewImage(dw, dh)
	}
	f.overlay.Clear()
	f.overlay.Fill(color.RGBA{
		R: f.Color.R,
		G: f.Color.G,
		B: f.Color.B,
		A: uint8(float64(f.Color.A) * alpha),
	})
	f.mu.Unlock()
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(f.overlay, op)
}
