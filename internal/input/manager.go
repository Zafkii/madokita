package input

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

type Action int

const (
	ActionMoveLeft Action = iota
	ActionMoveRight
	ActionJump
	ActionDodge
	ActionAttack
	ActionSkill
	ActionTransform
	ActionMenuConfirm
	ActionMenuBack
	ActionMenuUp
	ActionMenuDown
	ActionMenuLeft
	ActionMenuRight
	ActionToggleFullscreen
	ActionToggleDebug
)

type Chord struct {
	Mod ebiten.Key
	Key ebiten.Key
}

var DefaultBindings = map[Action]ebiten.Key{
	ActionMoveLeft:    ebiten.KeyA,
	ActionMoveRight:   ebiten.KeyD,
	ActionJump:        ebiten.KeyW,
	ActionDodge:       ebiten.KeyShift,
	ActionAttack:      ebiten.KeySpace,
	ActionSkill:       ebiten.KeyE,
	ActionTransform:   ebiten.KeyQ,
	ActionMenuConfirm: ebiten.KeyEnter,
	ActionMenuBack:    ebiten.KeyEscape,
	ActionMenuUp:      ebiten.KeyUp,
	ActionMenuDown:    ebiten.KeyDown,
	ActionMenuLeft:    ebiten.KeyLeft,
	ActionMenuRight:    ebiten.KeyRight,
	ActionToggleDebug:  ebiten.KeyF1,
}

var DefaultChordBindings = map[Action]Chord{
	ActionToggleFullscreen: {Mod: ebiten.KeyAlt, Key: ebiten.KeyEnter},
}

var capturableKeys = []ebiten.Key{
	ebiten.KeyA, ebiten.KeyB, ebiten.KeyC, ebiten.KeyD,
	ebiten.KeyE, ebiten.KeyF, ebiten.KeyG, ebiten.KeyH,
	ebiten.KeyI, ebiten.KeyJ, ebiten.KeyK, ebiten.KeyL,
	ebiten.KeyM, ebiten.KeyN, ebiten.KeyO, ebiten.KeyP,
	ebiten.KeyQ, ebiten.KeyR, ebiten.KeyS, ebiten.KeyT,
	ebiten.KeyU, ebiten.KeyV, ebiten.KeyW, ebiten.KeyX,
	ebiten.KeyY, ebiten.KeyZ,
	ebiten.Key0, ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4,
	ebiten.Key5, ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9,
	ebiten.KeyUp, ebiten.KeyDown, ebiten.KeyLeft, ebiten.KeyRight,
	ebiten.KeyEnter, ebiten.KeyEscape, ebiten.KeySpace, ebiten.KeyShift,
	ebiten.KeyTab, ebiten.KeyBackspace, ebiten.KeyControl, ebiten.KeyAlt,
	ebiten.KeyF1, ebiten.KeyF2, ebiten.KeyF3, ebiten.KeyF4,
	ebiten.KeyF5, ebiten.KeyF6, ebiten.KeyF7, ebiten.KeyF8,
	ebiten.KeyF9, ebiten.KeyF10, ebiten.KeyF11, ebiten.KeyF12,
}

type Manager struct {
	mu             sync.RWMutex
	bindings       map[Action]ebiten.Key
	chordBindings  map[Action]Chord
	prevKeys       map[ebiten.Key]bool
	prevMouseLeft  bool
	prevMouseRight bool
}

func NewManager() *Manager {
	b := make(map[Action]ebiten.Key, len(DefaultBindings))
	for k, v := range DefaultBindings {
		b[k] = v
	}
	cb := make(map[Action]Chord, len(DefaultChordBindings))
	for k, v := range DefaultChordBindings {
		cb[k] = v
	}
	return &Manager{
		bindings:      b,
		chordBindings: cb,
		prevKeys:      make(map[ebiten.Key]bool),
	}
}

func (m *Manager) Update() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, key := range m.bindings {
		m.prevKeys[key] = ebiten.IsKeyPressed(key)
	}
	for _, chord := range m.chordBindings {
		m.prevKeys[chord.Mod] = ebiten.IsKeyPressed(chord.Mod)
		m.prevKeys[chord.Key] = ebiten.IsKeyPressed(chord.Key)
	}
	m.prevMouseLeft = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	m.prevMouseRight = ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
}

func (m *Manager) SetBinding(action Action, key ebiten.Key) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bindings[action] = key
}

func (m *Manager) GetBinding(action Action) ebiten.Key {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bindings[action]
}

func (m *Manager) SetChordBinding(action Action, chord Chord) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chordBindings[action] = chord
}

func (m *Manager) GetChordBinding(action Action) (Chord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	chord, ok := m.chordBindings[action]
	return chord, ok
}

func (m *Manager) IsPressed(action Action) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key, ok := m.bindings[action]
	if !ok {
		return false
	}
	return ebiten.IsKeyPressed(key)
}

func (m *Manager) IsJustPressed(action Action) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key, ok := m.bindings[action]
	if !ok {
		return false
	}
	return ebiten.IsKeyPressed(key) && !m.prevKeys[key]
}

func (m *Manager) IsChordJustPressed(action Action) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	chord, ok := m.chordBindings[action]
	if !ok {
		return false
	}
	modHeld := ebiten.IsKeyPressed(chord.Mod)
	keyNow := ebiten.IsKeyPressed(chord.Key)
	keyPrev := m.prevKeys[chord.Key]
	return modHeld && keyNow && !keyPrev
}

func (m *Manager) CursorPosition() (int, int) {
	return ebiten.CursorPosition()
}

func (m *Manager) IsLeftClick() bool {
	return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func (m *Manager) IsLeftClickJustPressed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !m.prevMouseLeft
}

func (m *Manager) CaptureAnyKey(knownKeys map[ebiten.Key]bool) (ebiten.Key, bool) {
	for _, k := range capturableKeys {
		pressed := ebiten.IsKeyPressed(k)
		was := knownKeys[k]
		if pressed && !was {
			return k, true
		}
		knownKeys[k] = pressed
	}
	return 0, false
}

func ActionName(a Action) string {
	switch a {
	case ActionMoveLeft: return "MoveLeft"
	case ActionMoveRight: return "MoveRight"
	case ActionJump: return "Jump"
	case ActionDodge: return "Dodge"
	case ActionAttack: return "Attack"
	case ActionSkill: return "Skill"
	case ActionTransform: return "Transform"
	case ActionMenuConfirm: return "MenuConfirm"
	case ActionMenuBack: return "MenuBack"
	case ActionMenuUp: return "MenuUp"
	case ActionMenuDown: return "MenuDown"
	case ActionMenuLeft: return "MenuLeft"
	case ActionMenuRight: return "MenuRight"
	case ActionToggleFullscreen: return "ToggleFullscreen"
	case ActionToggleDebug: return "Debug"
	}
	return ""
}

func ParseAction(name string) (Action, bool) {
	for a := ActionMoveLeft; a <= ActionToggleDebug; a++ {
		if ActionName(a) == name {
			return a, true
		}
	}
	return 0, false
}

func (m *Manager) LoadBindings(data map[string]int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for a, k := range DefaultBindings {
		m.bindings[a] = k
	}
	for name, keyCode := range data {
		if a, ok := ParseAction(name); ok {
			m.bindings[a] = ebiten.Key(keyCode)
		}
	}
}
