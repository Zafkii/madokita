package main

import (
	"image/color"
	"math"

	"madokita/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *GameApp) syncTitleBarScale() {
	ow := g.outsideWidth
	if ow <= 0 || g.gameWidth <= 0 {
		return
	}
	scale := float64(ow) / float64(g.gameWidth)
	bh := int(float64(titleBarPhysH) / scale)
	if bh < 1 {
		bh = 1
	}
	bw := int(float64(btnPhysW) / scale)
	if bw < 1 {
		bw = 1
	}
	if bh != g.barLogicH || bw != g.btnLogicW {
		g.barLogicH = bh
		g.btnLogicW = bw
		if g.titleBarImg != nil {
			g.titleBarImg.Deallocate()
			g.titleBarImg = nil
		}
	}

	targetSize := math.Round(14.0 / scale)
	if targetSize != g.prevTitleFontSize {
		g.prevTitleFontSize = targetSize
		ui.SetTitleFontPixelSize(targetSize)
	}
}

func (g *GameApp) drawTitleBar(screen *ebiten.Image) {
	w := g.gameWidth
	bh := g.barLogicH
	if w <= 0 || bh <= 0 || screen.Bounds().Dy() < bh {
		return
	}
	if g.titleBarImg == nil || g.titleBarW != w {
		if g.titleBarImg != nil {
			g.titleBarImg.Deallocate()
		}
		g.titleBarImg = ebiten.NewImage(w, bh)
		g.titleBarImg.Fill(color.RGBA{18, 18, 30, 200})
		g.titleBarW = w
	}
	screen.DrawImage(g.titleBarImg, nil)

	bW, bH := g.btnLogicW, g.barLogicH
	if w < bW*3 {
		return
	}

	leftX := 8

	if g.titleLogo != nil {
		logoTargetH := g.barLogicH - 4
		srcH := g.titleLogo.Bounds().Dy()
		srcW := g.titleLogo.Bounds().Dx()
		if srcH > 0 {
			scale := float64(logoTargetH) / float64(srcH)
			dstW := int(float64(srcW) * scale)
			dstY := (g.barLogicH - logoTargetH) / 2
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(leftX), float64(dstY))
			screen.DrawImage(g.titleLogo, op)
			leftX += dstW + 6
		}
	}

	title := "Madokita"
	tw := ui.TitleTextWidth(title)
	th := ui.TitleTextHeight()
	padY := (g.barLogicH-th)/2 + ui.TitleAscent()
	if leftX+tw < w-bW*3-4 {
		ui.DrawTitleText(screen, title, leftX, padY, color.RGBA{230, 230, 245, 255})
	}

	btnW := float32(bW)
	barH := float32(bH)

	closeX := float32(w - bW)
	maxiX := float32(w - bW*2)
	miniX := float32(w - bW*3)

	bg := color.RGBA{35, 35, 50, 220}
	bgHov := color.RGBA{55, 55, 75, 240}
	bgClose := color.RGBA{180, 40, 40, 220}
	bgCloseHov := color.RGBA{220, 60, 60, 240}

	if g.hoveredBtn == btnMinimize {
		vector.DrawFilledRect(screen, miniX, 0, btnW, barH, bgHov, false)
	} else {
		vector.DrawFilledRect(screen, miniX, 0, btnW, barH, bg, false)
	}
	if g.hoveredBtn == btnMaximize {
		vector.DrawFilledRect(screen, maxiX, 0, btnW, barH, bgHov, false)
	} else {
		vector.DrawFilledRect(screen, maxiX, 0, btnW, barH, bg, false)
	}
	if g.hoveredBtn == btnClose {
		vector.DrawFilledRect(screen, closeX, 0, btnW, barH, bgCloseHov, false)
	} else {
		vector.DrawFilledRect(screen, closeX, 0, btnW, barH, bgClose, false)
	}

	sep := color.RGBA{50, 50, 70, 255}
	vector.StrokeLine(screen, closeX, 0, closeX, barH, 1, sep, false)
	vector.StrokeLine(screen, maxiX, 0, maxiX, barH, 1, sep, false)
	vector.StrokeLine(screen, miniX, 0, miniX, barH, 1, sep, false)

	sym := color.RGBA{230, 230, 245, 255}
	sw := max(float32(1), btnW/10)
	midY := barH * 0.5

	// Maximize / Restore (leftmost button)
	mmx := maxiX + btnW*0.5
	rw := btnW * 0.48
	rh := barH * 0.44
	if ebiten.IsWindowMaximized() {
		off := btnW * 0.14
		vector.StrokeRect(screen, mmx-rw*0.5-off, midY-rh*0.5-off, rw, rh, sw, sym, false)
		vector.StrokeRect(screen, mmx-rw*0.5, midY-rh*0.5, rw, rh, sw, sym, false)
	} else {
		vector.StrokeRect(screen, mmx-rw*0.5, midY-rh*0.5, rw, rh, sw, sym, false)
	}

	// Minimize (middle button)
	cmx := miniX + btnW*0.5
	ly := barH * 0.68
	hl := btnW * 0.25
	vector.StrokeLine(screen, cmx-hl, ly, cmx+hl, ly, sw, sym, false)

	// Close (rightmost button)
	dmx := closeX + btnW*0.5
	d := btnW * 0.28
	vector.StrokeLine(screen, dmx-d, midY-d, dmx+d, midY+d, sw, sym, false)
	vector.StrokeLine(screen, dmx+d, midY-d, dmx-d, midY+d, sw, sym, false)
}

func (g *GameApp) titleBarButtonAt(mx int) titleBarBtn {
	bW := g.btnLogicW
	switch {
	case mx >= g.gameWidth-bW:
		return btnClose
	case mx >= g.gameWidth-bW*2:
		return btnMaximize
	case mx >= g.gameWidth-bW*3:
		return btnMinimize
	}
	return btnNone
}
