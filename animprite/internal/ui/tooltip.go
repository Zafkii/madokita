package ui

import (
	"animprite/internal/theme"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	tooltipPadX   = 6
	tooltipPadY   = 3
	tooltipGapX   = 14
	tooltipGapY   = 18
	tooltipMargin = 4
)

func tooltipSize(str string) (w, h int) {
	w = TextWidth(str) + tooltipPadX*2
	if w < 24 {
		w = 24
	}
	h = TextHeight() + tooltipPadY*2
	return w, h
}

// tooltipPlacement puts the tooltip near the cursor and flips or clamps it so
// the whole rectangle stays inside the window.
func tooltipPlacement(mx, my, w, h, screenW, screenH int) (x, y int) {
	x = mx + tooltipGapX
	y = my + tooltipGapY
	if x+w > screenW-tooltipMargin {
		x = mx - w - tooltipGapX
	}
	if y+h > screenH-tooltipMargin {
		y = my - h - tooltipGapY
	}
	if x < tooltipMargin {
		x = tooltipMargin
	}
	if y < tooltipMargin {
		y = tooltipMargin
	}
	return x, y
}

// DrawTooltip draws a small overlay near the cursor. The caller is responsible
// for selecting the hovered control and drawing this last so the tooltip stays
// above panels and tables.
func DrawTooltip(screen *ebiten.Image, p theme.Palette, str string, mx, my, screenW, screenH int) {
	if str == "" || screenW <= 0 || screenH <= 0 {
		return
	}

	maxTextW := screenW - 2*tooltipMargin - 2*tooltipPadX
	if maxTextW <= 0 {
		return
	}
	str = TruncateText(str, maxTextW)

	w, h := tooltipSize(str)
	x, y := tooltipPlacement(mx, my, w, h, screenW, screenH)

	FillRect(screen, x, y, w, h, p.DropdownPopupBG)
	FillRect(screen, x, y, w, 1, p.DropdownPopupBorder)
	FillRect(screen, x, y+h-1, w, 1, p.DropdownPopupBorder)
	FillRect(screen, x, y, 1, h, p.DropdownPopupBorder)
	FillRect(screen, x+w-1, y, 1, h, p.DropdownPopupBorder)
	DrawText(screen, str, x+tooltipPadX, y+tooltipPadY, 1, p.DropdownPopupText)
}
