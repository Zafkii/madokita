package menu

import (
	"image/color"
	"madokita/internal/audio"
	"madokita/internal/scene"
	"madokita/internal/ui"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// time values for the intro slides, in seconds
const (
	introFadeIn   = 0.5
	introHold     = 1.2
	introFadeOut  = 0.5
	slide2FadeIn  = 1.0
	slide2Hold    = 3.0
	slide2FadeOut = 0.8
)

type slideState int

const (
	slideFadeIn slideState = iota
	slideHold
	slideFadeOut
	slideDone
)

type slide struct {
	path        string
	fadeInTime  float64
	holdTime    float64
	fadeOutTime float64
	img         *ebiten.Image
	baseScale   float64
	baseOX      float64
	baseOY      float64
}

type MenuIntroScene struct {
	mgr           *scene.Manager
	cache         *ui.ImageCache
	audioMgr      *audio.AudioManager
	slides        []slide
	curSlide      int
	state         slideState
	timer         float64
	entered       bool
	startPos      float64
	slideStartPos float64
}

func NewMenuIntroScene(mgr *scene.Manager, cache *ui.ImageCache, audioMgr *audio.AudioManager) *MenuIntroScene {
	return &MenuIntroScene{
		mgr:      mgr,
		cache:    cache,
		audioMgr: audioMgr,
		slides: []slide{
			{path: "menu/prevmenu1.png", fadeInTime: introFadeIn, holdTime: introHold, fadeOutTime: introFadeOut},
			{path: "menu/prevmenu2.png", fadeInTime: slide2FadeIn, holdTime: slide2Hold, fadeOutTime: slide2FadeOut},
		},
	}
}

func easeOutBack(t float64) float64 {
	const c1 = 1.70158
	const c3 = c1 + 1
	return 1 + c3*math.Pow(t-1, 3) + c1*math.Pow(t-1, 2)
}

func easeInQuad(t float64) float64 {
	return t * t
}

func fitToScreen(img *ebiten.Image, screenW, screenH int) (scale, ox, oy float64) {
	b := img.Bounds()
	scaleX := float64(screenW) / float64(b.Dx())
	scaleY := float64(screenH) / float64(b.Dy())
	scale = math.Min(scaleX, scaleY)
	fw := float64(b.Dx()) * scale
	fh := float64(b.Dy()) * scale
	ox = (float64(screenW) - fw) / 2
	oy = (float64(screenH) - fh) / 2
	return
}

func (s *MenuIntroScene) currentPos() float64 {
	if s.audioMgr != nil {
		pos, ok := s.audioMgr.GetPosition("menu-theme")
		if ok {
			return pos.Seconds() - s.startPos
		}
	}
	return s.timer
}

func (s *MenuIntroScene) Update(dt float64) error {
	if !s.entered || s.curSlide >= len(s.slides) {
		return nil
	}
	s.timer += dt
	t := s.currentPos() - s.slideStartPos
	sl := &s.slides[s.curSlide]
	switch s.state {
	case slideFadeIn:
		if t >= sl.fadeInTime {
			s.state = slideHold
		}
	case slideHold:
		if t >= sl.holdTime {
			s.state = slideFadeOut
		}
	case slideFadeOut:
		if t >= sl.fadeOutTime {
			s.curSlide++
			s.loadSlide(s.curSlide)
			if s.curSlide >= len(s.slides) {
				s.mgr.SwitchWithFade("main-menu", scene.DefaultFadeDuration)
				return nil
			}
			s.slideStartPos = s.currentPos()
			s.state = slideFadeIn
		}
	}
	return nil
}

func (s *MenuIntroScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})
	if s.curSlide >= len(s.slides) {
		return
	}
	sl := &s.slides[s.curSlide]
	if sl.img == nil {
		return
	}

	t := s.currentPos() - s.slideStartPos
	clamp := func(v, lo, hi float64) float64 {
		if v < lo {
			return lo
		}
		if v > hi {
			return hi
		}
		return v
	}

	scale := sl.baseScale
	alpha := 1.0

	switch s.state {
	case slideFadeIn:
		progress := 0.0
		if sl.fadeInTime > 0 {
			progress = clamp(t/sl.fadeInTime, 0, 1)
		}
		tweenBack := easeOutBack(progress)
		animScale := 0.85 + 0.15*tweenBack
		scale = sl.baseScale * animScale
		alpha = progress

	case slideFadeOut:
		progress := 0.0
		if sl.fadeOutTime > 0 {
			progress = clamp(t/sl.fadeOutTime, 0, 1)
		}
		zoomScale := 1.0 + 0.12*easeInQuad(progress)
		scale = sl.baseScale * zoomScale
		alpha = 1.0 - progress
	}

	if alpha <= 0 {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(sl.baseOX, sl.baseOY)
	op.ColorScale.Scale(1, 1, 1, float32(alpha))
	screen.DrawImage(sl.img, op)
}

func (s *MenuIntroScene) loadSlide(idx int) {
	if idx >= len(s.slides) {
		return
	}
	sl := &s.slides[idx]
	sl.img = nil
	img, err := s.cache.Load(sl.path)
	if err != nil {
		return
	}
	sl.img = img
	sl.baseScale, sl.baseOX, sl.baseOY = fitToScreen(img, 1280, 720)
}

func (s *MenuIntroScene) Enter() error {
	s.curSlide = 0
	s.state = slideFadeIn
	s.timer = 0
	s.slideStartPos = 0
	s.entered = true

	s.loadSlide(0)

	if s.audioMgr != nil {
		s.audioMgr.PlayLoop("menu-theme")
		pos, ok := s.audioMgr.GetPosition("menu-theme")
		if ok {
			s.startPos = pos.Seconds()
		}
	}
	return nil
}

func (s *MenuIntroScene) Exit() error {
	s.entered = false
	return nil
}
func (s *MenuIntroScene) Pause()         {}
func (s *MenuIntroScene) Resume()        {}
func (s *MenuIntroScene) IsActive() bool { return s.entered }
