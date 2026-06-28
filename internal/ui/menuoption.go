package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type MenuOption struct {
	Cx, Cy      int
	Text        string
	Scale       float64
	Alpha       float64
	YOffset     float64
	NormalClr   color.RGBA
	ActiveClr   color.RGBA
	HoverScale  float64
	hovered     bool
	focused     bool
	onClick     func()
}

func NewMenuOption(cx, cy int, text string, scale float64) *MenuOption {
	return &MenuOption{
		Cx: cx, Cy: cy,
		Text:       text,
		Scale:      scale,
		Alpha:      1,
		NormalClr:  color.RGBA{255, 255, 255, 255},
		ActiveClr:  color.RGBA{255, 65, 255, 255},
		HoverScale: scale * 1.08,
	}
}

func (o *MenuOption) SetOnClick(fn func()) { o.onClick = fn }

func (o *MenuOption) Bounds() (x, y, w, h int) {
	tw := TextWidth(o.Text)
	th := TextHeight()
	sc := o.Scale
	w = int(float64(tw) * sc)
	h = int(float64(th) * sc)
	x = o.Cx - w/2
	y = o.Cy + int(o.YOffset) - h/2
	return
}

func (o *MenuOption) HandleMouse(cx, cy int, justPressed bool) bool {
	bx, by, bw, bh := o.Bounds()
	over := cx >= bx && cx <= bx+bw && cy >= by && cy <= by+bh
	o.hovered = over
	if over && justPressed && o.onClick != nil {
		o.onClick()
		return true
	}
	return false
}

func (o *MenuOption) SetFocused(v bool) { o.focused = v }

func (o *MenuOption) Click() {
	if o.onClick != nil {
		o.onClick()
	}
}

func (o *MenuOption) Draw(screen *ebiten.Image) {
	if o.Alpha < 0.01 {
		return
	}

	displayScale := o.Scale
	clr := o.NormalClr
	if o.hovered || o.focused {
		displayScale = o.HoverScale
		clr = o.ActiveClr
	}
	clr.A = uint8(math.Max(0, math.Min(255, float64(clr.A)*float64(o.Alpha))))

	DrawTextCentered(screen, o.Text, o.Cx, o.Cy+int(o.YOffset), displayScale, clr)
}
