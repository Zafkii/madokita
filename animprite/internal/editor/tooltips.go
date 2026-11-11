package editor

import (
	"animprite/internal/theme"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

// drawTooltips is called after the panels so a tooltip is never clipped by the
// canvas or right panel. Only the topmost hovered control draws a tooltip.
func (a *EditorApp) drawTooltips(screen *ebiten.Image, p theme.Palette) {
	if a.confirm.visible {
		return
	}
	mx, my := ebiten.CursorPosition()
	if a.modeDropdown.PopupHit(mx, my) {
		return
	}
	for _, btn := range a.tooltipButtons() {
		if btn == nil || btn.Tooltip == "" {
			continue
		}
		if btn.HitTest(mx, my) {
			ui.DrawTooltip(screen, p, btn.Tooltip, mx, my, a.win.outsideWidth, a.win.outsideHeight)
			return
		}
	}
}

func (a *EditorApp) tooltipButtons() []*ui.Button {
	btns := []*ui.Button{a.openBtn, a.openDirBtn, a.saveBtn, a.themeBtn}
	if a.isStageMode() {
		return btns
	}
	for _, table := range []*ui.Table{a.animTable, a.spriteTable, a.hurtboxTable, a.hitboxTable} {
		if table == nil {
			continue
		}
		btns = append(btns, table.AddBtn, table.RemoveBtn)
		btns = append(btns, table.ExtraBtns...)
	}
	btns = appendVisibleRowBtns(btns, a.animTable,
		a.animAddFrameBtns, a.animRemoveFrameBtns, a.animFramePrevBtns, a.animFrameNextBtns)
	btns = appendVisibleRowBtns(btns, a.spriteTable,
		a.spriteBrowseBtns, a.spriteFramePrevBtns, a.spriteFrameNextBtns)
	return btns
}

func appendVisibleRowBtns(dst []*ui.Button, table *ui.Table, groups ...[]*ui.Button) []*ui.Button {
	if table == nil {
		return dst
	}
	start, end := table.VisibleRange()
	if start < 0 {
		start = 0
	}
	for _, group := range groups {
		if start >= len(group) {
			continue
		}
		groupEnd := end
		if groupEnd > len(group) {
			groupEnd = len(group)
		}
		if groupEnd > start {
			dst = append(dst, group[start:groupEnd]...)
		}
	}
	return dst
}
