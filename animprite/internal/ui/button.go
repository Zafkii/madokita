package ui

import (
	"animprite/internal/theme"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type ButtonType int

const (
	BtnDefault ButtonType = iota
	BtnGreen
	BtnRed
	BtnBlue
	BtnOrange
)

type Button struct {
	X, Y, W, H int
	Text       string
	Visible    bool
	Enabled    bool
	OnClick    func()
	th         *theme.Manager
	BtnType    ButtonType
	hovered    bool
}

func NewButton(x, y, w, h int, text string, th *theme.Manager) *Button {
	return &Button{
		X: x, Y: y, W: w, H: h,
		Text:    text,
		Visible: true,
		Enabled: true,
		th:      th,
		BtnType: BtnGreen,
	}
}

func (b *Button) HitTest(cx, cy int) bool {
	return b.Visible && b.Enabled && cx >= b.X && cx <= b.X+b.W && cy >= b.Y && cy <= b.Y+b.H
}

func (b *Button) HandleMouse(cx, cy int, justPressed bool) {
	if !b.Visible || !b.Enabled {
		b.hovered = false
		return
	}
	b.hovered = b.HitTest(cx, cy)
	if b.hovered && justPressed && b.OnClick != nil {
		b.OnClick()
	}
}

func (b *Button) bgColor() color.Color {
	p := b.th.Current
	if !b.Enabled {
		return p.BtnDisabled
	}
	switch b.BtnType {
	case BtnGreen:
		return p.BtnGreen
	case BtnRed:
		return p.BtnRed
	case BtnBlue:
		return p.BtnBlue
	case BtnOrange:
		return p.BtnOrange
	default:
		return p.BtnGreen
	}
}

func (b *Button) Draw(screen *ebiten.Image) {
	if !b.Visible {
		return
	}
	p := b.th.Current
	FillRect(screen, b.X, b.Y, b.W, b.H, b.bgColor())
	if b.hovered && b.Enabled {
		FillRect(screen, b.X, b.Y, b.W, b.H, p.BtnHover)
	}
	textClr := p.ButtonText
	if !b.Enabled {
		textClr = p.TextMuted
	} else if b.hovered {
		textClr = p.BtnTextHover
	}
	DrawTextCentered(screen, b.Text, b.X+b.W/2, b.Y+b.H/2, 1, textClr)
}
