package menu

import (
	"madokita/internal/audio"
	"madokita/internal/input"
	"madokita/internal/scene"
	"madokita/internal/ui"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type bindingEntry struct {
	label  string
	action input.Action
}

var (
	uiBindings = []bindingEntry{
		{"MENU UP", input.ActionMenuUp},
		{"MENU DOWN", input.ActionMenuDown},
		{"MENU LEFT", input.ActionMenuLeft},
		{"MENU RIGHT", input.ActionMenuRight},
		{"MENU CONFIRM", input.ActionMenuConfirm},
		{"MENU BACK", input.ActionMenuBack},
	}
	gameBindings = []bindingEntry{
		{"MOVE LEFT", input.ActionMoveLeft},
		{"MOVE RIGHT", input.ActionMoveRight},
		{"JUMP", input.ActionJump},
		{"DODGE", input.ActionDodge},
		{"ATTACK", input.ActionAttack},
		{"SKILL", input.ActionSkill},
		{"TRANSFORM", input.ActionTransform},
	}
	systemBindings = []bindingEntry{
		{"TOGGLE FULLSCREEN", input.ActionToggleFullscreen},
	}
	navLabels  = []string{"Profile", "Display", "Controllers", "Volume", "System"}
	navTabKeys = []navTab{navProfile, navDisplay, navControllers, navVolume, navSystem}
)

var resPresets = []string{"1920x1080", "1600x900", "1280x720", "854x480"}
var fpsPresets = []string{"30", "60", "120"}

type SettingsScene struct {
	mgr       *scene.Manager
	inputMgr  *input.Manager
	cache     *ui.ImageCache
	audioMgr  *audio.AudioManager
	animTitle *AnimatedTitle

	selNav    navTab
	selMiddle int
	rightType rightContent
	colFocus  colFocus
	selRight  int

	capturing  bool
	captureIdx int

	hoverNav    int
	hoverMiddle int
	hoverRight  int

	inNav    bool
	inMiddle bool
	inRight  bool

	navX, infoX, actionX int
	contentY, titleY     int
	spacing              int
	active               bool

	showResetDialog bool
	resetDialogTab  navTab
	resetFocusYes   bool
}

func NewSettingsScene(mgr *scene.Manager, cache *ui.ImageCache, inputMgr *input.Manager, audioMgr *audio.AudioManager) *SettingsScene {
	return &SettingsScene{
		mgr:      mgr,
		cache:    cache,
		inputMgr: inputMgr,
		audioMgr: audioMgr,
	}
}

func (s *SettingsScene) Enter() error {
	s.active = true
	s.selNav = navProfile
	s.selMiddle = -1
	s.rightType = rcProfile
	s.colFocus = focusNav
	s.selRight = 0
	s.capturing = false
	s.showResetDialog = false

	base := s.cache.Get("menu/madokita-title.png")
	overlay := s.cache.Get("menu/madokita-title-top.png")
	cosmic := s.cache.Get("menu/cosmic-effect.png")
	if base != nil {
		s.animTitle = NewAnimatedTitle(base, overlay, cosmic, s.audioMgr)
		s.animTitle.Reset()
	}

	s.titleY = int(math.Floor(float64(DesignHeight) * TitleYRatio))
	s.contentY = int(math.Floor(float64(DesignHeight) * ContentYRatio))
	s.spacing = int(math.Max(float64(SpacingMin), float64(DesignHeight)*SpacingRatio))

	sc := 2.0

	navMax := 0
	for _, lbl := range navLabels {
		w := int(float64(ui.TextWidth(lbl)) * sc)
		if w > navMax {
			navMax = w
		}
	}

	s.infoX = SettingsInfoX
	s.actionX = SettingsActionX
	s.navX = s.infoX - navMax - 10
	return nil
}

func (s *SettingsScene) Exit() error {
	s.active = false
	return nil
}
func (s *SettingsScene) Pause()         { s.active = false }
func (s *SettingsScene) Resume()        { s.active = true }
func (s *SettingsScene) IsActive() bool { return s.active }

func keyDisplayName(k ebiten.Key) string {
	switch k {
	case ebiten.KeyUp:
		return "↑"
	case ebiten.KeyDown:
		return "↓"
	case ebiten.KeyLeft:
		return "←"
	case ebiten.KeyRight:
		return "→"
	case ebiten.KeyEnter:
		return "ENTER"
	case ebiten.KeyEscape:
		return "ESC"
	case ebiten.KeySpace:
		return "SPACE"
	case ebiten.KeyShift:
		return "SHIFT"
	case ebiten.KeyTab:
		return "TAB"
	case ebiten.KeyBackspace:
		return "BACK"
	default:
		return strings.TrimPrefix(k.String(), "Key")
	}
}

func (s *SettingsScene) pageRowY(i int) int {
	return s.contentY + i*s.spacing
}

func (s *SettingsScene) backY() int {
	return DesignHeight - 40
}

func (s *SettingsScene) textBounds(text string, x, y int, scale float64) (x0, y0, x1, y1 int) {
	w := int(float64(ui.TextWidth(text)) * scale)
	h := int(float64(ui.TextHeight()) * scale)
	return x, y, x + w, y + h
}
