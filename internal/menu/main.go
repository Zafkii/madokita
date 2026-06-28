package menu

import (
	"image/color"
	"madokita/internal/audio"
	"madokita/internal/input"
	"madokita/internal/localization"
	"madokita/internal/scene"
	"madokita/internal/ui"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type mainMenuPhase int

const (
	phaseTouchToStart mainMenuPhase = iota
	phaseOptions
)

type MainMenuScene struct {
	mgr         *scene.Manager
	cache       *ui.ImageCache
	inputMgr    *input.Manager
	audioMgr    *audio.AudioManager
	phase       mainMenuPhase
	phaseTimer  float64
	animTitle   *AnimatedTitle
	options     []*ui.MenuOption
	selectedIdx int

	ttsText     string
	ttsAlive    bool
	ttsAlpha    float64
	ttsTimer    float64

	appearStarted bool
	appearTimer   float64

	active    bool
	onNewGame func()
}

func NewMainMenuScene(mgr *scene.Manager, cache *ui.ImageCache, inputMgr *input.Manager, audioMgr *audio.AudioManager) *MainMenuScene {
	return &MainMenuScene{
		mgr:      mgr,
		cache:    cache,
		inputMgr: inputMgr,
		audioMgr: audioMgr,
	}
}

func (s *MainMenuScene) Enter() error {
	s.phase = phaseTouchToStart
	s.phaseTimer = 0
	s.selectedIdx = 0
	s.active = true

	base := s.cache.Get("menu/madokita-title.png")
	overlay := s.cache.Get("menu/madokita-title-top.png")
	cosmic := s.cache.Get("menu/cosmic-effect.png")
	if base != nil {
		s.animTitle = NewAnimatedTitle(base, overlay, cosmic, s.audioMgr)
		s.animTitle.Reset()
	}

	s.ttsText = localization.Get("MENU.TOUCH_TO_START")
	s.ttsAlive = true
	s.ttsAlpha = 1.0
	s.ttsTimer = 0

	labels := []string{
		localization.Get("MENU.NEW_GAME"),
		localization.Get("MENU.CONTINUE"),
		localization.Get("MENU.SETTINGS"),
		localization.Get("MENU.EXIT"),
	}

	scale := float64(DesignHeight) / DesignHeight
	optScale := 2.5
	spacing := int(float64(ListSpacing) * scale)
	totalH := (len(labels) - 1) * spacing
	startY := DesignHeight/2 + int(float64(ListOffset)*scale) - totalH/2
	cx := DesignWidth / 2

	s.options = make([]*ui.MenuOption, len(labels))
	for i, lbl := range labels {
		opt := ui.NewMenuOption(cx, startY+i*spacing, lbl, optScale)
		idx := i
		opt.SetOnClick(func() { s.onOption(idx) })
		opt.Alpha = 0
		opt.YOffset = -8
		s.options[i] = opt
	}

	s.appearStarted = false
	s.appearTimer = 0
	s.phaseTimer = 0

	return nil
}

func (s *MainMenuScene) SetOnNewGame(fn func()) {
	s.onNewGame = fn
}

func (s *MainMenuScene) onOption(idx int) {
	switch idx {
	case 0:
		if s.onNewGame != nil {
			s.onNewGame()
		}
	case 1:
	case 2:
		s.mgr.SwitchTo("settings")
	case 3:
	}
}

func (s *MainMenuScene) Update(dt float64) error {
	if !s.active {
		return nil
	}
	s.phaseTimer += dt

	if s.animTitle != nil {
		s.animTitle.Update(dt)
	}

	mx, my := s.inputMgr.CursorPosition()
	gw, gh := s.mgr.GameSize()
	mx = mx * DesignWidth / gw
	my = my * DesignHeight / gh
	clicked := s.inputMgr.IsLeftClickJustPressed()
	confirm := s.inputMgr.IsJustPressed(input.ActionMenuConfirm)

	switch s.phase {
	case phaseTouchToStart:
		s.ttsTimer += dt
		tt := math.Mod(s.ttsTimer, 1.2)
		if tt < 0.6 {
			s.ttsAlpha = 1.0 - tt/0.6
		} else {
			s.ttsAlpha = (tt - 0.6) / 0.6
		}

		if clicked || s.inputMgr.IsJustPressed(input.ActionAttack) || confirm {
			s.phase = phaseOptions
			s.phaseTimer = 0
			s.ttsAlive = false
			s.appearStarted = true
			s.appearTimer = 0
		}

	case phaseOptions:
		if s.inputMgr.IsJustPressed(input.ActionMenuUp) || s.inputMgr.IsJustPressed(input.ActionMoveLeft) {
			s.setSelected((s.selectedIdx - 1 + len(s.options)) % len(s.options))
		}
		if s.inputMgr.IsJustPressed(input.ActionMenuDown) || s.inputMgr.IsJustPressed(input.ActionMoveRight) {
			s.setSelected((s.selectedIdx + 1) % len(s.options))
		}
		for _, opt := range s.options {
			opt.HandleMouse(mx, my, clicked)
		}
		if confirm && len(s.options) > 0 {
			s.options[s.selectedIdx].Click()
		}
	}

	if s.appearStarted {
		s.appearTimer += dt
		stagger := 0.12
		dur := 0.4
		for i, opt := range s.options {
			t := s.appearTimer - float64(i)*stagger
			if t <= 0 {
				continue
			}
			if t >= dur {
				opt.Alpha = 1
				opt.YOffset = 0
				continue
			}
			progress := math.Sin(t/dur * math.Pi / 2)
			opt.Alpha = progress
			opt.YOffset = -8 * (1 - progress)
		}
	}

	return nil
}

func (s *MainMenuScene) setSelected(idx int) {
	for i := range s.options {
		s.options[i].SetFocused(i == idx)
	}
	s.selectedIdx = idx
}

func (s *MainMenuScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	if s.animTitle != nil {
		s.animTitle.Draw(screen)
	}

	switch s.phase {
	case phaseTouchToStart:
		if s.ttsAlive {
			ttsY := int(float64(DesignHeight) * TouchYRatio)
			ui.DrawTextCentered(screen, s.ttsText, DesignWidth/2, ttsY, 2.5,
				color.RGBA{255, 255, 255, uint8(math.Max(0, math.Min(255, s.ttsAlpha*255)))})
		}

	case phaseOptions:
		for _, opt := range s.options {
			opt.Draw(screen)
		}
	}
}

func (s *MainMenuScene) Exit() error {
	s.active = false
	return nil
}
func (s *MainMenuScene) Pause()         { s.active = false }
func (s *MainMenuScene) Resume()        { s.active = true }
func (s *MainMenuScene) IsActive() bool { return s.active }
