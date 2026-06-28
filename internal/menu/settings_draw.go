package menu

import (
	"fmt"
	"image/color"
	"madokita/internal/localization"
	"madokita/internal/settings"
	"madokita/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func (s *SettingsScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})
	if s.animTitle != nil {
		s.animTitle.Draw(screen)
	}

	ui.DrawTextCentered(screen, "Settings",
		DesignWidth/2, s.titleY, 3.0, color.RGBA{255, 255, 255, 255})

	cNrm := color.RGBA{255, 255, 255, 255}
	cAct := color.RGBA{255, 65, 255, 255}
	cMuted := color.RGBA{136, 136, 136, 255}
	sc := 2.0

	// ── Left: nav ──
	for _, nt := range navTabKeys {
		clr := cMuted
		if s.hoverNav >= 0 {
			if int(nt) == s.hoverNav {
				clr = cAct
			} else if nt == s.selNav {
				clr = cAct
			} else {
				clr = cNrm
			}
		} else if s.colFocus == focusNav || s.inNav {
			if nt == s.selNav {
				clr = cAct
			} else {
				clr = cNrm
			}
		} else if nt == s.selNav {
			clr = cAct
		}
		ui.DrawTextRight(screen, navLabels[nt], s.infoX, s.pageRowY(int(nt)), sc, clr)
	}

	// ── Middle: content ──
	for i, lbl := range s.middleItems() {
		clr := cMuted
		ds := sc
		if s.hoverMiddle >= 0 {
			if i == s.hoverMiddle {
				clr = cAct
				ds = sc * 1.05
			} else if i == s.selMiddle {
				clr = cAct
				ds = sc * 1.05
			} else {
				clr = cNrm
			}
		} else if s.colFocus == focusMiddle || s.inMiddle {
			if i == s.selMiddle {
				clr = cAct
				ds = sc * 1.05
			} else {
				clr = cNrm
			}
		} else if i == s.selMiddle {
			clr = cAct
			ds = sc * 1.05
		}
		midW := int(float64(ui.TextWidth(lbl)) * ds)
		midX := (s.infoX + s.actionX - midW) / 2
		ui.DrawText(screen, lbl, midX, s.pageRowY(i), ds, clr)
	}

	// ── Footer: back / reset ──
	backLabel := localization.Get("MENU.BACK")
	if s.selNav == navProfile {
		ui.DrawTextCentered(screen, backLabel,
			DesignWidth/2, s.backY(), sc, cMuted)
	} else {
		resetLabel := localization.Get("MENU.RESET")
		gap := 40
		backW := int(float64(ui.TextWidth(backLabel)) * sc)
		resetW := int(float64(ui.TextWidth(resetLabel)) * sc)
		totalW := backW + gap + resetW
		startX := (DesignWidth - totalW) / 2
		ui.DrawText(screen, backLabel, startX, s.backY(), sc, cMuted)
		ui.DrawText(screen, resetLabel, startX+backW+gap, s.backY(), sc, cMuted)
	}

	// ── Right: content ──
	switch s.rightType {
	case rcNone:

	case rcResolution, rcFPS, rcMode:
		data := settings.GetData()
		var currentValue string
		switch s.rightType {
		case rcResolution:
			currentValue = fmt.Sprintf("%dx%d", data.Resolution.Width, data.Resolution.Height)
		case rcFPS:
			currentValue = fmt.Sprintf("%d", data.FPSLimit)
		case rcMode:
			if data.Fullscreen {
				currentValue = "Fullscreen"
			} else {
				currentValue = "Windowed"
			}
		}
		for i, lbl := range s.rightOptionLabels() {
			clr := cMuted
			ds := sc
			if s.hoverRight >= 0 {
				if i == s.hoverRight {
					clr = cAct
					ds = sc * 1.05
				} else if i == s.selRight {
					clr = cAct
					ds = sc * 1.05
				} else {
					clr = cNrm
				}
			} else if s.colFocus == focusRight || s.inRight {
				if i == s.selRight {
					clr = cAct
					ds = sc * 1.05
				} else {
					clr = cNrm
				}
			} else {
				if lbl == currentValue {
					clr = cAct
				}
			}
			ui.DrawText(screen, lbl, s.actionX, s.pageRowY(i), ds, clr)
		}

	case rcProfile:
		profileData := []string{
			"Name: Sayaka",
			"Level: 12",
			"Location: Stage 3",
			"Completion: 45%",
			"Fav Char: Sayaka",
		}
		for i, txt := range profileData {
			clr := cMuted
			ds := sc
			if s.hoverRight >= 0 {
				if i == s.hoverRight {
					clr = cAct
					ds = sc * 1.05
				} else if i == s.selRight {
					clr = cAct
					ds = sc * 1.05
				} else {
					clr = cNrm
				}
			} else if s.colFocus == focusRight || s.inRight {
				if i == s.selRight {
					clr = cAct
					ds = sc * 1.05
				} else {
					clr = cNrm
				}
			} else if i == s.selRight {
				clr = cAct
				ds = sc * 1.05
			}
			ui.DrawText(screen, txt, s.actionX, s.pageRowY(i), ds, clr)
		}

	case rcVolumeBar:
		filled := s.selRight
		pct := (s.selRight * 100) / 20
		clr := cMuted
		ds := sc
		if s.colFocus == focusRight || s.inRight {
			clr = cAct
			ds = sc * 1.05
		}
		barW := 200
		barH := int(float64(ui.TextHeight()) * ds)
		barX := s.actionX
		barY := s.pageRowY(0)
		fillW := barW * filled / 20

		ui.FillRect(screen, barX, barY, barW, barH, color.RGBA{60, 60, 60, 255})
		if fillW > 0 {
			ui.FillRect(screen, barX, barY, fillW, barH, clr)
		}
		ui.DrawText(screen, fmt.Sprintf("%d%%", pct), barX+barW+8, barY, ds, clr)

	case rcBindings:
		bindings := s.currentBindings()
		for i, e := range bindings {
			clr := cMuted
			ds := sc
			if s.capturing && s.captureIdx == i {
				clr = cAct
				ds = sc * 1.05
			} else if s.hoverRight >= 0 {
				if i == s.hoverRight {
					clr = cAct
					ds = sc * 1.05
				} else if i == s.selRight {
					clr = cAct
					ds = sc * 1.05
				} else {
					clr = cNrm
				}
			} else if s.colFocus == focusRight || s.inRight {
				if i == s.selRight {
					clr = cAct
					ds = sc * 1.05
				} else {
					clr = cNrm
				}
			} else if i == s.selRight {
				clr = cAct
				ds = sc * 1.05
			}
			if chord, ok := s.inputMgr.GetChordBinding(e.action); ok {
				labelLine := e.label + ":"
				keyLine := "  " + keyDisplayName(chord.Mod) + "+" + keyDisplayName(chord.Key)
				if s.capturing && s.captureIdx == i {
					keyLine = "  ..."
				}
				lineH := int(float64(ui.TextHeight()) * ds)
				ui.DrawText(screen, labelLine, s.actionX, s.pageRowY(i), ds, clr)
				ui.DrawText(screen, keyLine, s.actionX, s.pageRowY(i)+lineH, ds, clr)
			} else {
				display := s.bindingDisplayString(e)
				if s.capturing && s.captureIdx == i {
					display = e.label + ": ..."
				}
				ui.DrawText(screen, display, s.actionX, s.pageRowY(i), ds, clr)
			}
		}
	}

	if s.showResetDialog {
		// overlay
		ui.FillRect(screen, 0, 0, DesignWidth, DesignHeight, color.RGBA{0, 0, 0, 180})

		boxW, boxH := 420, 140
		boxX := (DesignWidth - boxW) / 2
		boxY := (DesignHeight - boxH) / 2

		ui.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{48, 48, 48, 255})
		ui.FillRect(screen, boxX, boxY, boxW, 2, color.RGBA{180, 180, 180, 255})
		ui.FillRect(screen, boxX, boxY+boxH-2, boxW, 2, color.RGBA{180, 180, 180, 255})
		ui.FillRect(screen, boxX, boxY, 2, boxH, color.RGBA{180, 180, 180, 255})
		ui.FillRect(screen, boxX+boxW-2, boxY, 2, boxH, color.RGBA{180, 180, 180, 255})

		tabName := navLabels[s.resetDialogTab]
		msg := fmt.Sprintf("Reset %s settings to default?", tabName)
		ui.DrawTextCentered(screen, msg, DesignWidth/2, boxY+50, sc, color.RGBA{255, 255, 255, 255})

		gap := 40
		yesW := int(float64(ui.TextWidth("Yes")) * sc)
		noW := int(float64(ui.TextWidth("No")) * sc)
		totalW := yesW + gap + noW
		btnY := boxY + boxH - int(float64(ui.TextHeight())*sc) - 16
		startX := (DesignWidth - totalW) / 2
		ui.DrawText(screen, "Yes", startX, btnY, sc, cMuted)
		ui.DrawText(screen, "No", startX+yesW+gap, btnY, sc, cMuted)
	}
}
