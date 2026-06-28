package menu

import (
	"image/color"
	"madokita/internal/audio"
	"madokita/internal/localization"
	"madokita/internal/scene"
	"madokita/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

type PreloadScene struct {
	mgr       *scene.Manager
	cache     *ui.ImageCache
	audioMgr  *audio.AudioManager
	images    []string
	loadDone  bool
	loadErr   error
	musicPath string
}

func NewPreloadScene(mgr *scene.Manager, cache *ui.ImageCache, audioMgr *audio.AudioManager) *PreloadScene {
	return &PreloadScene{
		mgr:      mgr,
		cache:    cache,
		audioMgr: audioMgr,
		images: []string{
			"menu/madokita-title.png",
			"menu/madokita-title-top.png",
			"menu/star-title-.png",
			"menu/star-title-top.png",
			"menu/prevmenu1.png",
			"menu/prevmenu2.png",
			"menu/cosmic-effect.png",
			"images/loading.png",
		},
		musicPath: "sounds/music/menutheme.ogg",
	}
}

func (s *PreloadScene) Update(dt float64) error {
	if s.loadErr != nil {
		return s.loadErr
	}
	if s.loadDone {
		s.loadDone = false
		s.mgr.SwitchWithFade("menu-intro", scene.DefaultFadeDuration)
	}
	return nil
}

func (s *PreloadScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})
}

func (s *PreloadScene) Enter() error {
	s.loadDone = false
	s.loadErr = nil

	if err := s.cache.Preload(s.images...); err != nil {
		s.loadErr = err
		return nil
	}

	localization.Initialize("en")

	if s.audioMgr != nil {
		if err := s.audioMgr.LoadOGGLoop("menu-theme", s.musicPath, "music"); err == nil {
			s.audioMgr.PlayLoop("menu-theme")
		}
	}

	s.loadDone = true

	return nil
}

func (s *PreloadScene) Exit() error {
	return nil
}

func (s *PreloadScene) Pause()         {}
func (s *PreloadScene) Resume()        {}
func (s *PreloadScene) IsActive() bool { return true }
