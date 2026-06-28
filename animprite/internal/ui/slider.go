package ui

import (
	"animprite/internal/theme"

	"github.com/hajimehoshi/ebiten/v2"
)

type Slider struct {
	X, Y, W, H int
	Min, Max   float64
	Value      float64
	Step       float64
	Visible    bool
	Enabled    bool
	OnChange   func(float64)
	th         *theme.Manager
	dragging   bool
	hovered    bool
	thumbW     int
}

func NewSlider(x, y, w, h int, th *theme.Manager) *Slider {
	return &Slider{
		X: x, Y: y, W: w, H: h,
		Min:     0,
		Max:     1,
		Step:    0.01,
		Visible: true,
		Enabled: true,
		th:      th,
		thumbW:  12,
	}
}

const handleH = 8
const trackH = 4

func (s *Slider) trackX() int { return s.X }
func (s *Slider) trackW() int { return s.W - s.thumbW }
func (s *Slider) thumbX() int {
	t := 0.0
	if s.Max > s.Min {
		t = (s.Value - s.Min) / (s.Max - s.Min)
	}
	return s.X + int(t*float64(s.trackW()))
}

func (s *Slider) thumbRect() (x, y, w, h int) {
	return s.thumbX(), s.Y + (s.H-handleH)/2, s.thumbW, handleH
}

func (s *Slider) valueAt(cx int) float64 {
	ratio := float64(cx-s.X) / float64(s.trackW())
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	v := s.Min + ratio*(s.Max-s.Min)
	if s.Step > 0 {
		steps := (v - s.Min) / s.Step
		v = s.Min + float64(int(steps))*s.Step
	}
	return v
}

func (s *Slider) HandleMouse(cx, cy int, justPressed bool) {
	if !s.Visible || !s.Enabled {
		s.dragging = false
		s.hovered = false
		return
	}
	tx, ty, tw, th := s.thumbRect()
	s.hovered = cx >= tx && cx <= tx+tw && cy >= ty && cy <= ty+th
	if justPressed {
		if cx >= tx && cx <= tx+tw && cy >= ty && cy <= ty+th {
			s.dragging = true
		} else if cx >= s.X && cx <= s.X+s.W && cy >= s.Y && cy <= s.Y+s.H {
			v := s.valueAt(cx)
			if v != s.Value {
				s.Value = v
				if s.OnChange != nil {
					s.OnChange(v)
				}
			}
		}
	}
	if !justPressed && s.dragging {
		v := s.valueAt(cx)
		if v != s.Value {
			s.Value = v
			if s.OnChange != nil {
				s.OnChange(v)
			}
		}
	}
}

func (s *Slider) HandleRelease() {
	s.dragging = false
}

func (s *Slider) Draw(screen *ebiten.Image) {
	if !s.Visible {
		return
	}
	p := s.th.Current

	ty := s.Y + (s.H-trackH)/2
	tx, ty2, tw, th2 := s.thumbRect()

	FillRect(screen, s.X, ty, s.W, trackH, p.InputBorder)
	filledW := tx - s.X
	if filledW > 0 {
		FillRect(screen, s.X, ty, filledW, trackH, p.BtnBlue)
	}

	clr := p.BtnBlue
	if s.dragging || !s.Enabled {
		clr = p.BtnOrange
	}
	FillRect(screen, tx, ty2, tw, th2, clr)
	if s.hovered && !s.dragging && s.Enabled {
		FillRect(screen, tx, ty2, tw, th2, p.BtnHover)
	}
}
