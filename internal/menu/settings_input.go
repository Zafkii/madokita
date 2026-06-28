package menu

import (
	"fmt"
	"madokita/internal/input"
	"madokita/internal/localization"
	"madokita/internal/settings"
	"madokita/internal/ui"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

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
	ebiten.KeyTab, ebiten.KeyBackspace,
	ebiten.KeyF1, ebiten.KeyF2, ebiten.KeyF3, ebiten.KeyF4,
	ebiten.KeyF5, ebiten.KeyF6, ebiten.KeyF7, ebiten.KeyF8,
	ebiten.KeyF9, ebiten.KeyF10, ebiten.KeyF11, ebiten.KeyF12,
}

func (s *SettingsScene) bindingHitBounds(e bindingEntry, idx int, sc float64) (x0, y0, x1, y1 int) {
	y := s.pageRowY(idx)
	if chord, ok := s.inputMgr.GetChordBinding(e.action); ok {
		lineH := int(float64(ui.TextHeight()) * sc)
		_, _, xa, _ := s.textBounds(e.label+":", s.actionX, y, sc)
		_, _, xb, yb := s.textBounds("  "+keyDisplayName(chord.Mod)+"+"+keyDisplayName(chord.Key), s.actionX, y+lineH, sc)
		if xb > xa {
			xa = xb
		}
		return s.actionX, y, xa, yb
	}
	return s.textBounds(s.bindingDisplayString(e), s.actionX, y, sc)
}

func (s *SettingsScene) Update(dt float64) error {
	if !s.active {
		return nil
	}
	if s.animTitle != nil {
		s.animTitle.Update(dt)
	}

	mx, my := s.inputMgr.CursorPosition()
	gw, gh := s.mgr.GameSize()
	mx = mx * DesignWidth / gw
	my = my * DesignHeight / gh
	clicked := s.inputMgr.IsLeftClickJustPressed()
	back := s.inputMgr.IsJustPressed(input.ActionMenuBack)
	sc := 2.0

	if s.showResetDialog {
		if back {
			s.showResetDialog = false
			return nil
		}
		if clicked {
			boxW, boxH := 420, 160
			boxX := (DesignWidth - boxW) / 2
			boxY := (DesignHeight - boxH) / 2
			yesW := int(float64(ui.TextWidth("Yes")) * sc)
			noW := int(float64(ui.TextWidth("No")) * sc)
			gap := 40
			totalW := yesW + gap + noW
			btnY := boxY + boxH - int(float64(ui.TextHeight())*sc) - 16
			startX := (DesignWidth - totalW) / 2
			yesX := startX
			noX := startX + yesW + gap

			_, _, yx1, yy1 := s.textBounds("Yes", yesX, btnY, sc)
			if mx >= yesX && mx <= yx1 && my >= btnY && my <= yy1 {
				s.doReset()
				return nil
			}
			_, _, nx1, ny1 := s.textBounds("No", noX, btnY, sc)
			if mx >= noX && mx <= nx1 && my >= btnY && my <= ny1 {
				s.showResetDialog = false
				return nil
			}
			if mx < boxX || mx > boxX+boxW || my < boxY || my > boxY+boxH {
				s.showResetDialog = false
				return nil
			}
		}
		return nil
	}

	s.hoverNav = -1
	for _, nt := range navTabKeys {
		y := s.pageRowY(int(nt))
		w := int(float64(ui.TextWidth(navLabels[nt])) * sc)
		x0 := s.infoX - w
		if mx >= x0 && mx <= s.infoX && my >= y && my <= y+int(float64(ui.TextHeight())*sc) {
			s.hoverNav = int(nt)
		}
	}
	s.hoverMiddle = -1
	for i, lbl := range s.middleItems() {
		y := s.pageRowY(i)
		cx := (s.infoX + s.actionX) / 2
		w := int(float64(ui.TextWidth(lbl)) * sc)
		x0 := cx - w/2
		x1 := cx + w/2
		if mx >= x0 && mx <= x1 && my >= y && my <= y+int(float64(ui.TextHeight())*sc) {
			s.hoverMiddle = i
		}
	}
	s.hoverRight = -1
	if s.rightType == rcBindings {
		bindings := s.currentBindings()
		for i, e := range bindings {
			_, y0, x1, y1 := s.bindingHitBounds(e, i, sc)
			if mx >= s.actionX && mx <= x1 && my >= y0 && my <= y1 {
				s.hoverRight = i
			}
		}
	} else {
		for i, lbl := range s.rightOptionLabels() {
			y := s.pageRowY(i)
			_, _, x1, y1 := s.textBounds(lbl, s.actionX, y, sc)
			if mx >= s.actionX && mx <= x1 && my >= y && my <= y1 {
				s.hoverRight = i
			}
		}
	}

	s.inNav = mx >= s.navX && mx < s.infoX
	s.inMiddle = mx >= s.infoX && mx < s.actionX
	s.inRight = mx >= s.actionX

	if s.rightType == rcVolumeBar && s.colFocus == focusRight && back {
		s.colFocus = focusMiddle
		return nil
	}
	if back {
		s.mgr.SwitchTo("main-menu")
		return nil
	}

	if s.capturing {
		if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			s.capturing = false
			return nil
		}
		for _, k := range capturableKeys {
			if k != ebiten.KeyEscape && ebiten.IsKeyPressed(k) {
				bindings := s.currentBindings()
				if s.captureIdx >= 0 && s.captureIdx < len(bindings) {
					e := bindings[s.captureIdx]
					s.inputMgr.SetBinding(e.action, k)
					settings.SetKeyBinding(input.ActionName(e.action), int(k))
				}
				s.capturing = false
				return nil
			}
		}
		return nil
	}

	for _, nt := range navTabKeys {
		y := s.pageRowY(int(nt))
		w := int(float64(ui.TextWidth(navLabels[nt])) * sc)
		x0 := s.infoX - w
		if mx >= x0 && mx <= s.infoX && my >= y && my <= y+int(float64(ui.TextHeight())*sc) && clicked {
			s.setNav(nt)
		}
	}

	for i, lbl := range s.middleItems() {
		y := s.pageRowY(i)
		cx := (s.infoX + s.actionX) / 2
		w := int(float64(ui.TextWidth(lbl)) * sc)
		x0 := cx - w/2
		x1 := cx + w/2
		if mx >= x0 && mx <= x1 && my >= y && my <= y+int(float64(ui.TextHeight())*sc) && clicked {
			s.toggleMiddle(i)
			if s.selMiddle >= 0 {
				s.colFocus = focusRight
			}
		}
	}

	backLabel := localization.Get("MENU.BACK")
	if s.selNav == navProfile {
		backY := s.backY()
		bx := DesignWidth/2 - int(float64(ui.TextWidth(backLabel))*sc/2)
		if bx, by, x1, y1 := s.textBounds(backLabel, bx, backY, sc); mx >= bx && mx <= x1 && my >= by && my <= y1 && clicked {
			s.mgr.SwitchTo("main-menu")
			return nil
		}
	} else {
		resetLabel := localization.Get("MENU.RESET")
		gap := 40
		backW := int(float64(ui.TextWidth(backLabel)) * sc)
		resetW := int(float64(ui.TextWidth(resetLabel)) * sc)
		totalW := backW + gap + resetW
		startX := (DesignWidth - totalW) / 2
		bx := startX
		if bx, by, x1, y1 := s.textBounds(backLabel, bx, s.backY(), sc); mx >= bx && mx <= x1 && my >= by && my <= y1 && clicked {
			s.mgr.SwitchTo("main-menu")
			return nil
		}
		rx := startX + backW + gap
		if rx, ry, x1, y1 := s.textBounds(resetLabel, rx, s.backY(), sc); mx >= rx && mx <= x1 && my >= ry && my <= y1 && clicked {
			s.showResetDialog = true
			s.resetDialogTab = s.selNav
			s.resetFocusYes = false
		}
	}

	if s.rightType == rcBindings {
		bindings := s.currentBindings()
		for i, e := range bindings {
			_, y0, x1, y1 := s.bindingHitBounds(e, i, sc)
			if mx >= s.actionX && mx <= x1 && my >= y0 && my <= y1 && clicked {
				s.selRight = i
				if _, ok := s.inputMgr.GetChordBinding(e.action); ok {
					continue
				}
				s.startCapture(i)
			}
		}
	} else {
		for i, lbl := range s.rightOptionLabels() {
			y := s.pageRowY(i)
			_, _, x1, y1 := s.textBounds(lbl, s.actionX, y, sc)
			if mx >= s.actionX && mx <= x1 && my >= y && my <= y1 && clicked {
				s.selRight = i
				s.selectRightOption(i)
			}
		}
	}

	if s.rightType == rcVolumeBar {
		barW := 200
		y := s.pageRowY(0)
		if mx >= s.actionX && mx <= s.actionX+barW && my >= y && my <= y+int(float64(ui.TextHeight())*sc) && clicked {
			block := (mx - s.actionX) * 20 / barW
			if block < 0 {
				block = 0
			}
			if block > 20 {
				block = 20
			}
			s.selRight = block
			s.applyVolumeFromSelRight()
		}
		_, wy := ebiten.Wheel()
		if wy != 0 {
			steps := int(math.Round(wy))
			s.selRight += steps
			if s.selRight < 0 {
				s.selRight = 0
			}
			if s.selRight > 20 {
				s.selRight = 20
			}
			s.applyVolumeFromSelRight()
		}
	}

	if s.rightType == rcVolumeBar && s.colFocus == focusRight {
		if s.inputMgr.IsJustPressed(input.ActionMenuLeft) && s.selRight > 0 {
			s.selRight--
			s.applyVolumeFromSelRight()
		}
		if s.inputMgr.IsJustPressed(input.ActionMenuRight) && s.selRight < 20 {
			s.selRight++
			s.applyVolumeFromSelRight()
		}
	} else {
		if s.inputMgr.IsJustPressed(input.ActionMenuLeft) {
			if s.colFocus > focusNav {
				s.colFocus--
				if s.colFocus == focusMiddle && len(s.middleItems()) == 0 {
					s.colFocus = focusNav
				}
			}
		}
		if s.inputMgr.IsJustPressed(input.ActionMenuRight) {
			if s.colFocus < focusRight {
				s.colFocus++
				if s.colFocus == focusMiddle && len(s.middleItems()) == 0 {
					s.colFocus = focusRight
				}
			}
		}
	}

	if s.inputMgr.IsJustPressed(input.ActionMenuUp) {
		switch s.colFocus {
		case focusNav:
			s.selNav = navTab((int(s.selNav) - 1 + int(navCount)) % int(navCount))
			s.setNav(s.selNav)
		case focusMiddle:
			items := s.middleItems()
			if len(items) > 0 {
				s.selMiddle = (s.selMiddle - 1 + len(items)) % len(items)
				s.buildRight()
			}
		case focusRight:
			labels := s.rightOptionLabels()
			if len(labels) > 0 {
				s.selRight = (s.selRight - 1 + len(labels)) % len(labels)
			}
		}
	}
	if s.inputMgr.IsJustPressed(input.ActionMenuDown) {
		switch s.colFocus {
		case focusNav:
			s.selNav = navTab((int(s.selNav) + 1) % int(navCount))
			s.setNav(s.selNav)
		case focusMiddle:
			items := s.middleItems()
			if len(items) > 0 {
				s.selMiddle = (s.selMiddle + 1) % len(items)
				s.buildRight()
			}
		case focusRight:
			labels := s.rightOptionLabels()
			if len(labels) > 0 {
				s.selRight = (s.selRight + 1) % len(labels)
			}
		}
	}

	if s.inputMgr.IsJustPressed(input.ActionMenuConfirm) {
		switch s.colFocus {
		case focusNav:
			items := s.middleItems()
			if len(items) > 0 {
				s.selMiddle = 0
				s.buildRight()
				s.colFocus = focusMiddle
			} else {
				s.colFocus = focusRight
			}
		case focusMiddle:
			if s.selMiddle >= 0 && len(s.middleItems()) > 0 {
				s.buildRight()
				s.colFocus = focusRight
			}
		case focusRight:
			if s.rightType == rcBindings {
				if s.selRight >= 0 {
					bindings := s.currentBindings()
					if s.selRight < len(bindings) {
						if _, ok := s.inputMgr.GetChordBinding(bindings[s.selRight].action); ok {
							break
						}
					}
					s.startCapture(s.selRight)
				}
			} else {
				s.selectRightOption(s.selRight)
			}
		}
	}

	return nil
}

func (s *SettingsScene) startCapture(entryIdx int) {
	s.capturing = true
	s.captureIdx = entryIdx
}

func (s *SettingsScene) bindingDisplayString(e bindingEntry) string {
	if chord, ok := s.inputMgr.GetChordBinding(e.action); ok {
		return fmt.Sprintf("%s: %s+%s", e.label, keyDisplayName(chord.Mod), keyDisplayName(chord.Key))
	}
	return fmt.Sprintf("%s: %s", e.label, keyDisplayName(s.inputMgr.GetBinding(e.action)))
}

func (s *SettingsScene) doReset() {
	s.showResetDialog = false
	switch s.resetDialogTab {
	case navDisplay:
		settings.ResetDisplay()
		s.buildRight()
	case navControllers:
		settings.ResetControllers()
		s.inputMgr.LoadBindings(settings.GetData().KeyBindings)
		s.buildRight()
	case navVolume:
		settings.ResetVolume()
		s.buildRight()
	case navSystem:
		settings.ResetSystem()
		s.buildRight()
	}
}
