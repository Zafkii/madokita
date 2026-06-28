package menu

import (
	"image/color"
	"madokita/internal/scene"
	"madokita/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

type LoadingScene struct {
	mgr     *scene.Manager
	cache   *ui.ImageCache
	loadFn  func() error
	target  string
	loadDone bool
	loadErr  error
}

func NewLoadingScene(mgr *scene.Manager, cache *ui.ImageCache) *LoadingScene {
	return &LoadingScene{
		mgr:   mgr,
		cache: cache,
	}
}

func (s *LoadingScene) SetTarget(loadFn func() error, target string) {
	s.loadFn = loadFn
	s.target = target
	s.loadDone = false
	s.loadErr = nil
}

func (s *LoadingScene) Update(dt float64) error {
	if s.loadErr != nil {
		return s.loadErr
	}
	if s.loadDone {
		s.loadDone = false
		s.mgr.SwitchWithFade(s.target, scene.DefaultFadeDuration)
	}
	return nil
}

func (s *LoadingScene) Draw(screen *ebiten.Image) {
	if img := s.cache.Get("images/loading.png"); img != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(img, op)
	} else {
		screen.Fill(color.RGBA{0, 0, 0, 255})
	}
}

func (s *LoadingScene) Enter() error {
	s.loadDone = false
	s.loadErr = nil

	if s.loadFn != nil {
		if err := s.loadFn(); err != nil {
			s.loadErr = err
			return nil
		}
	}
	s.loadDone = true
	return nil
}

func (s *LoadingScene) Exit() error {
	return nil
}

func (s *LoadingScene) Pause()         {}
func (s *LoadingScene) Resume()        {}
func (s *LoadingScene) IsActive() bool { return true }
