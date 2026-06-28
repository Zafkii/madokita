package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	btnIdleBG    = color.RGBA{80, 80, 80, 200}
	btnHoverBG   = color.RGBA{120, 120, 120, 220}
	btnPressedBG = color.RGBA{60, 60, 60, 200}
	btnTextClr   = color.RGBA{255, 255, 255, 255}
	btnBorderClr = color.RGBA{180, 180, 180, 255}
)

type Button struct {
	X, Y, W, H int
	Text       string
	TextScale  float64
	Visible    bool
	Enabled    bool
	hovered    bool
	pressed    bool
	onClick    func()
	bgSurf     *ebiten.Image
	pixel      *ebiten.Image
}

var btnPixel *ebiten.Image

func init() {
	btnPixel = ebiten.NewImage(1, 1)
	btnPixel.Fill(color.White)
}

func NewButton(x, y, w, h int, text string) *Button {
	b := &Button{
		X: x, Y: y, W: w, H: h,
		Text:      text,
		TextScale: 2.0,
		Visible:   true,
		Enabled:   true,
	}
	b.rebuildBG()
	return b
}

func (b *Button) rebuildBG() {
	if b.W <= 0 || b.H <= 0 {
		return
	}
	b.bgSurf = ebiten.NewImage(b.W, b.H)
	fillSolid(b.bgSurf, 0, 0, b.W, b.H, btnIdleBG)
	// border
	fillSolid(b.bgSurf, 0, 0, b.W, 1, btnBorderClr)
	fillSolid(b.bgSurf, 0, b.H-1, b.W, 1, btnBorderClr)
	fillSolid(b.bgSurf, 0, 0, 1, b.H, btnBorderClr)
	fillSolid(b.bgSurf, b.W-1, 0, 1, b.H, btnBorderClr)
}

func fillSolid(dst *ebiten.Image, x, y, w, h int, clr color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	s := ebiten.NewImage(w, h)
	s.Fill(clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(s, op)
}

func (b *Button) SetOnClick(fn func()) {
	b.onClick = fn
}

func (b *Button) HandleMouse(cx, cy int, justPressed bool) {
	if !b.Visible || !b.Enabled {
		b.hovered = false
		b.pressed = false
		return
	}
	over := cx >= b.X && cx <= b.X+b.W && cy >= b.Y && cy <= b.Y+b.H
	b.hovered = over
	if over && justPressed {
		b.pressed = true
		if b.onClick != nil {
			b.onClick()
		}
	} else {
		b.pressed = false
	}
}

func (b *Button) HandleKeyboard(active bool, confirmPressed bool) {
	if !b.Visible || !b.Enabled {
		return
	}
	if active && confirmPressed && b.onClick != nil {
		b.onClick()
	}
}

func (b *Button) Draw(screen *ebiten.Image) {
	if !b.Visible {
		return
	}
	bg := btnIdleBG
	if b.pressed {
		bg = btnPressedBG
	} else if b.hovered {
		bg = btnHoverBG
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(b.W), float64(b.H))
	op.GeoM.Translate(float64(b.X), float64(b.Y))
	op.ColorScale.Scale(
		float32(bg.R)/255,
		float32(bg.G)/255,
		float32(bg.B)/255,
		float32(bg.A)/255,
	)
	screen.DrawImage(btnPixel, op)
	// border
	borderClr := btnBorderClr
	sc := func(c uint8) float32 { return float32(c) / 255 }
	op2 := &ebiten.DrawImageOptions{}
	op2.ColorScale.Scale(sc(borderClr.R), sc(borderClr.G), sc(borderClr.B), sc(borderClr.A))
	// top
	op2.GeoM.Scale(float64(b.W), 1)
	op2.GeoM.Translate(float64(b.X), float64(b.Y))
	screen.DrawImage(btnPixel, op2)
	op2.GeoM.Reset()
	op2.ColorScale.Scale(sc(borderClr.R), sc(borderClr.G), sc(borderClr.B), sc(borderClr.A))
	// bottom
	op2.GeoM.Scale(float64(b.W), 1)
	op2.GeoM.Translate(float64(b.X), float64(b.Y+b.H-1))
	screen.DrawImage(btnPixel, op2)
	op2.GeoM.Reset()
	op2.ColorScale.Scale(sc(borderClr.R), sc(borderClr.G), sc(borderClr.B), sc(borderClr.A))
	// left
	op2.GeoM.Scale(1, float64(b.H))
	op2.GeoM.Translate(float64(b.X), float64(b.Y))
	screen.DrawImage(btnPixel, op2)
	op2.GeoM.Reset()
	op2.ColorScale.Scale(sc(borderClr.R), sc(borderClr.G), sc(borderClr.B), sc(borderClr.A))
	// right
	op2.GeoM.Scale(1, float64(b.H))
	op2.GeoM.Translate(float64(b.X+b.W-1), float64(b.Y))
	screen.DrawImage(btnPixel, op2)

	DrawTextCentered(screen, b.Text, b.X+b.W/2, b.Y+b.H/2, b.TextScale, btnTextClr)
}
