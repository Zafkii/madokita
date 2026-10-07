package editor

import (
	"fmt"
	"image"

	"animprite/internal/theme"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func (a *EditorApp) Draw(screen *ebiten.Image) {
	p := a.th.Current
	a.drawModeIndicator(screen, p)
	a.drawTopPanel(screen, p)
	a.drawStatusbar(screen, p)

	a.buildSpriteRenders()
	a.canvas.Draw(screen)

	a.drawRightPanel(screen, p)
}

func (a *EditorApp) drawModeIndicator(screen *ebiten.Image, p theme.Palette) {
	ui.FillRect(screen, 0, 0, a.win.outsideWidth, a.modeIndH(), p.ModeIndBG)
	text := "Editing movement frames"
	if a.isAttackMode() {
		text = "Editing attack frames"
	}
	ui.DrawText(screen, text, 12, 4, 0.92, p.TextMuted)
}

func (a *EditorApp) modeIndH() int {
	return modeIndicatorH
}

func (a *EditorApp) drawTopPanel(screen *ebiten.Image, p theme.Palette) {
	ty := a.topPanelY()
	ui.FillRect(screen, 0, ty, a.win.outsideWidth, a.topPanelH, p.TopPanelBG)
	ui.FillRect(screen, 0, ty+a.topPanelH-1, a.win.outsideWidth, 1, p.Border)

	a.modeDropdown.Draw(screen)
	a.openBtn.Draw(screen)
	a.saveBtn.Draw(screen)
	a.themeBtn.Draw(screen)
	a.modeDropdown.DrawPopup(screen)

	tgX := a.tableGridX()
	tgW := a.tableGridW()
	halfW := (tgW - topTableGap) / 2

	ly := ty + topPanelPad
	ly = a.animTable.Draw(screen, p, tgX, ly, halfW)
	ly = a.spriteTable.Draw(screen, p, tgX, ly, halfW)

	rx := tgX + halfW + topTableGap
	ry := ty + topPanelPad
	ry = a.hurtboxTable.Draw(screen, p, rx, ry, halfW)
	ry = a.hitboxTable.Draw(screen, p, rx, ry, halfW)

	_ = max(ly, ry)
}

func (a *EditorApp) topPanelY() int {
	return a.modeIndH()
}

func (a *EditorApp) tableGridX() int {
	return panelPad + dropdownW + panelPad
}

func (a *EditorApp) tableGridW() int {
	return (a.win.outsideWidth - panelPad) - a.tableGridX()
}

func (a *EditorApp) drawStatusbar(screen *ebiten.Image, p theme.Palette) {
	y := a.win.outsideHeight - statusbarH
	ui.FillRect(screen, 0, y, a.win.outsideWidth, statusbarH, p.StatusbarBG)
	ui.FillRect(screen, 0, y, a.win.outsideWidth, 1, p.Border)

	switch {
	case a.hoveredFilePath != "":
		ui.DrawText(screen, "File: "+a.hoveredFilePath, 12, y+6, 1, p.TextPrimary)
	case a.statusMsg != "":
		ui.DrawText(screen, a.statusMsg, 12, y+6, 1, p.BtnGreen)
	default:
		zoomPct := int(a.canvas.Cam.Zoom*100 + 0.5)
		ui.DrawText(screen, "Zoom: "+itoa(zoomPct)+"%", 12, y+6, 1, p.TextPrimary)
	}

	a.win.resetViewBtnX = a.win.outsideWidth - rightPanelW - 100
	a.win.resetViewBtnY = y + 3
	a.win.resetViewBtnW = 90
	a.win.resetViewBtnH = statusbarH - 6
	ui.FillRect(screen, a.win.resetViewBtnX, a.win.resetViewBtnY, a.win.resetViewBtnW, a.win.resetViewBtnH, p.InputBG)
	ui.FillBorder(screen, a.win.resetViewBtnX, a.win.resetViewBtnY, a.win.resetViewBtnW, a.win.resetViewBtnH, p.InputBorder)
	ui.DrawTextCentered(screen, "Reset View", a.win.resetViewBtnX+45, a.win.resetViewBtnY+5, 1, p.TextPrimary)
}

func (a *EditorApp) drawRightPanel(screen *ebiten.Image, p theme.Palette) {
	x := a.win.outsideWidth - rightPanelW
	y := a.topPanelY() + a.topPanelH
	h := a.win.outsideHeight - y - statusbarH

	ui.FillRect(screen, x, y, rightPanelW, h, p.RightPanelBG)
	ui.FillRect(screen, x, y, 1, h, p.Border)

	fbW, fbH := a.win.outsideWidth, a.win.outsideHeight
	if fbW <= 0 || fbH <= 0 {
		return
	}
	if a.rp.buf == nil || a.rp.buf.Bounds().Dx() != fbW || a.rp.buf.Bounds().Dy() != fbH {
		if a.rp.buf != nil {
			a.rp.buf.Deallocate()
		}
		a.rp.buf = ebiten.NewImage(fbW, fbH)
	}
	a.rp.buf.Clear()

	px := x + rightPanelPad
	iy := y + 4 - a.rp.scroll

	iy = a.drawSelectedElementProps(a.rp.buf, p, px, iy)
	if a.hurtboxTable.SelectedIdx < 0 && a.hitboxTable.SelectedIdx < 0 {
		iy = a.drawBaseSpriteProps(a.rp.buf, p, px, iy)
	}
	if a.animTable.SelectedIdx >= 0 {
		iy = a.drawFrameSection(a.rp.buf, p, px, iy)
	}
	iy = a.drawAnimationSection(a.rp.buf, p, px, iy)
	iy = a.drawPreviewSection(a.rp.buf, p, px, iy)

	a.rp.contentH = iy - y + a.rp.scroll

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(a.rp.buf.SubImage(image.Rect(x, y, x+rightPanelW, y+h)).(*ebiten.Image), op)

	if a.isAttackMode() && a.phaseDropdown.IsOpen() {
		a.phaseDropdown.DrawPopup(screen)
	}

	if a.rp.contentH > h {
		sx := x + rightPanelW - 8
		sw := 4
		thumbH := max(int(h*h/a.rp.contentH), 10)
		maxScroll := a.rp.contentH - h
		pos := float64(a.rp.scroll) / float64(maxScroll)
		thumbY := y + int(pos*float64(h-thumbH))
		ui.FillRect(screen, sx, y, sw, h, p.ScrollbarTrack)
		ui.FillRect(screen, sx, thumbY, sw, thumbH, p.ScrollbarThumb)
	}
}

func (a *EditorApp) drawSectionHeader(screen *ebiten.Image, p theme.Palette, px, iy int, title string) int {
	ui.DrawText(screen, title, px, iy, 1.15, p.LabelColor)
	iy += 18
	ui.FillRect(screen, px, iy, rightPanelInner, 1, p.Border)
	iy += 6
	return iy
}

func (a *EditorApp) drawLabeledInput(screen *ebiten.Image, inp *ui.TextInput, iy int) int {
	inp.Visible = true
	inp.Y = iy
	inp.Draw(screen)
	return iy + rightPanelInputH + 2
}

func (a *EditorApp) drawLabeledDropdown(screen *ebiten.Image, dd *ui.Dropdown, iy int) int {
	dd.Visible = true
	dd.Enabled = true
	dd.Y = iy
	dd.Draw(screen)
	return iy + rightPanelInputH + 2
}

func (a *EditorApp) drawSelectedElementProps(screen *ebiten.Image, p theme.Palette, px, iy int) int {
	iy = a.drawSectionHeader(screen, p, px, iy, "Selected Element Properties")

	for i := 0; i < 5; i++ {
		iy = a.drawLabeledInput(screen, a.props[i], iy)
	}

	if a.isAttackMode() {
		iy = a.drawLabeledDropdown(screen, a.phaseDropdown, iy)
	}

	return iy + 4
}

func (a *EditorApp) drawBaseSpriteProps(screen *ebiten.Image, p theme.Palette, px, iy int) int {
	iy = a.drawSectionHeader(screen, p, px, iy, "Base Sprite Properties")

	iy = a.drawLabeledInput(screen, a.originInputs[0], iy)
	iy = a.drawLabeledInput(screen, a.originInputs[1], iy)
	iy = a.drawLabeledInput(screen, a.baseRotInput, iy)

	return iy + 4
}

func (a *EditorApp) drawFrameSection(screen *ebiten.Image, p theme.Palette, px, iy int) int {
	iy = a.drawSectionHeader(screen, p, px, iy, "Frame")

	iy = a.drawLabeledDropdown(screen, a.frameSpriteDropdown, iy)

	return iy + 4
}

func (a *EditorApp) drawAnimationSection(screen *ebiten.Image, p theme.Palette, px, iy int) int {
	iy = a.drawSectionHeader(screen, p, px, iy, "Animation")

	if a.isAttackMode() {
		ui.DrawText(screen, "Phase Durations (ms)", px, iy, 1, p.TextMuted)
		iy += 16
		for i := 0; i < 5; i++ {
			iy = a.drawLabeledInput(screen, a.atkTimingInputs[i], iy)
		}
	} else {
		iy = a.drawLabeledInput(screen, a.fpsInput, iy)
	}

	iy = a.drawLabeledInput(screen, a.loopInput, iy)

	iy += 4

	return iy + 4
}

func (a *EditorApp) drawPreviewSection(screen *ebiten.Image, p theme.Palette, px, iy int) int {
	iy = a.drawSectionHeader(screen, p, px, iy, "Preview")

	btnH := 22

	playW := 28
	playLabel := "\U000F040A"
	if a.prev.previewPlaying {
		playLabel = "\U000F03E4"
	}
	a.prev.previewPlayX, a.prev.previewPlayY = px, iy
	ui.FillRect(screen, px, iy, playW, btnH, p.BtnGreen)
	ui.FillBorder(screen, px, iy, playW, btnH, p.InputBorder)
	if a.prev.previewPlayHovered {
		ui.FillRect(screen, px, iy, playW, btnH, p.BtnHover)
	}
	ui.DrawTextCentered(screen, playLabel, px+playW/2, iy+btnH/2, 1, p.ButtonText)

	playTextX := px + playW + 4
	ui.DrawText(screen, "Play", playTextX, iy+4, 1, p.TextPrimary)

	chkX := px + playW + 4 + ui.TextWidth("Play") + 12
	chkSize := 14
	chkY := iy + (btnH-chkSize)/2
	a.prev.previewChkX, a.prev.previewChkY, a.prev.previewChkW, a.prev.previewChkH = chkX, chkY, chkSize, chkSize

	ui.FillRect(screen, chkX, chkY, chkSize, chkSize, p.InputBG)
	ui.FillBorder(screen, chkX, chkY, chkSize, chkSize, p.InputBorder)
	if a.prev.loopChecked {
		ui.FillRect(screen, chkX+3, chkY+3, chkSize-6, chkSize-6, p.BtnGreen)
	}
	if a.prev.previewLoopHovered {
		ui.FillRect(screen, chkX, chkY, chkSize, chkSize, p.BtnHover)
	}
	ui.DrawText(screen, "Loop", chkX+chkSize+4, iy+4, 1, p.TextPrimary)

	iy += btnH + 4

	a.prev.previewBtnH = btnH
	a.prev.previewRow2Y = iy

	speedLabel := "Speed:"
	ui.DrawText(screen, speedLabel, px, iy+4, 1, p.TextPrimary)
	speedLabelW := ui.TextWidth(speedLabel)

	btnSize := 20
	minusX := px + speedLabelW + 4
	minusY := iy + (btnH-btnSize)/2
	a.prev.previewMinusX, a.prev.previewMinusY = minusX, minusY

	ui.FillRect(screen, minusX, minusY, btnSize, btnSize, p.InputBG)
	ui.FillBorder(screen, minusX, minusY, btnSize, btnSize, p.InputBorder)
	if a.prev.previewMinusHovered {
		ui.FillRect(screen, minusX, minusY, btnSize, btnSize, p.BtnHover)
	}
	ui.DrawTextCentered(screen, "-", minusX+btnSize/2, minusY+btnSize/2, 1, p.TextPrimary)

	sliderX := minusX + btnSize + 2
	sliderW := rightPanelW - 16 - (sliderX - px) - btnSize - 2 - 4 - ui.TextWidth(" 1.0x")
	if sliderW < 40 {
		sliderW = 40
	}
	a.prev.speedSlider.X = sliderX
	a.prev.speedSlider.Y = iy + (btnH-16)/2
	a.prev.speedSlider.W = sliderW
	a.prev.speedSlider.H = 16
	a.prev.speedSlider.Visible = true
	a.prev.speedSlider.Draw(screen)

	plusX := sliderX + sliderW + 2
	plusY := iy + (btnH-btnSize)/2
	a.prev.previewPlusX, a.prev.previewPlusY = plusX, plusY

	ui.FillRect(screen, plusX, plusY, btnSize, btnSize, p.InputBG)
	ui.FillBorder(screen, plusX, plusY, btnSize, btnSize, p.InputBorder)
	if a.prev.previewPlusHovered {
		ui.FillRect(screen, plusX, plusY, btnSize, btnSize, p.BtnHover)
	}
	ui.DrawTextCentered(screen, "+", plusX+btnSize/2, plusY+btnSize/2, 1, p.TextPrimary)

	speedText := fmt.Sprintf("%.1fx", a.prev.previewSpeed)
	speedTextX := plusX + btnSize + 4
	ui.DrawText(screen, speedText, speedTextX, iy+4, 1, p.TextPrimary)

	iy += btnH + 6

	barW := rightPanelInner
	barH2 := 8
	ui.FillRect(screen, px, iy, barW, barH2, p.InputBG)
	ui.FillBorder(screen, px, iy, barW, barH2, p.InputBorder)
	ui.FillRect(screen, px, iy, 1, barH2, p.BtnBlue)

	iy += barH2 + 2
	ui.DrawTextCentered(screen, "1 / 0", px+barW/2, iy, 1, p.TextMuted)
	return iy
}

func (a *EditorApp) rightPanelY() int {
	return a.topPanelY() + a.topPanelH
}

func (a *EditorApp) rightPanelH() int {
	return a.win.outsideHeight - a.rightPanelY() - statusbarH
}
