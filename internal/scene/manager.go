package scene

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

type Manager struct {
	mu      sync.Mutex
	scenes  map[string]Interface
	current Interface
	name    string
	next    string
	fade    *Fade
	w       int
	h       int
	canvas  *ebiten.Image
}

const (
	canvasWidth  = 1280
	canvasHeight = 720
)

func NewManager(w, h int) *Manager {
	return &Manager{
		scenes: make(map[string]Interface),
		w:      w,
		h:      h,
	}
}

func (m *Manager) SetGameSize(w, h int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.w = w
	m.h = h
}

func (m *Manager) GameSize() (int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.w, m.h
}

func (m *Manager) Add(name string, s Interface) {
	m.scenes[name] = s
}

func (m *Manager) Get(name string) (Interface, bool) {
	s, ok := m.scenes[name]
	return s, ok
}

func (m *Manager) SwitchTo(name string) error {
	m.mu.Lock()
	s, ok := m.scenes[name]
	if !ok {
		m.mu.Unlock()
		return nil
	}
	if m.current != nil {
		m.current.Exit()
	}
	m.name = name
	m.current = s
	m.mu.Unlock()
	return s.Enter()
}

func (m *Manager) SwitchWithFade(name string, duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.scenes[name]; !ok {
		return
	}
	m.next = name
	m.fade = NewFade(duration, color.RGBA{0, 0, 0, 255})
	m.fade.StartOut()
}

func (m *Manager) Current() Interface {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.current
}

func (m *Manager) CurrentName() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.name
}

func (m *Manager) IsTransitioning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.fade != nil
}

func (m *Manager) Update(dt float64) error {
	m.mu.Lock()
	f := m.fade
	m.mu.Unlock()

	if f != nil {
		done, state := f.Update(dt)
		if done && state == FadeNone {
			m.mu.Lock()
			if m.next != "" {
				if m.current != nil {
					m.current.Exit()
				}
				m.name = m.next
				m.current = m.scenes[m.next]
				m.current.Enter()
				m.next = ""
				m.fade = NewFade(f.Duration, f.Color)
				m.fade.StartIn()
				m.mu.Unlock()
				return nil
			}
			m.fade = nil
			m.mu.Unlock()
			return nil
		}
		return nil
	}

	m.mu.Lock()
	cur := m.current
	m.mu.Unlock()

	if cur != nil && cur.IsActive() {
		return cur.Update(dt)
	}
	return nil
}

func (m *Manager) Draw(screen *ebiten.Image) {
	m.mu.Lock()
	f := m.fade
	cur := m.current
	m.mu.Unlock()

	if m.canvas == nil || m.canvas.Bounds().Dx() != canvasWidth || m.canvas.Bounds().Dy() != canvasHeight {
		m.canvas = ebiten.NewImage(canvasWidth, canvasHeight)
	}
	m.canvas.Clear()

	if cur != nil {
		cur.Draw(m.canvas)
	}
	if f != nil {
		f.Draw(m.canvas)
	}

	op := &ebiten.DrawImageOptions{}
	sx := float64(screen.Bounds().Dx()) / canvasWidth
	sy := float64(screen.Bounds().Dy()) / canvasHeight
	op.GeoM.Scale(sx, sy)
	screen.DrawImage(m.canvas, op)

	if cur != nil {
		if od, ok := cur.(OverlayDrawer); ok {
			od.DrawOverlay(screen)
		}
	}
}
