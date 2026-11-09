package editor

import (
	"fmt"
	"math"
	"path/filepath"

	"animprite/internal/project"
	"animprite/internal/theme"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func stageAssetCategory(asset project.StageAsset) string {
	if asset.Category == project.StageCategoryBG {
		return project.StageCategoryBG
	}
	return project.StageCategoryBlock
}

func drawImageFit(screen *ebiten.Image, img *ebiten.Image, x, y, w, h int) {
	if img == nil || w <= 0 || h <= 0 {
		return
	}
	b := img.Bounds()
	iw := float64(b.Dx())
	ih := float64(b.Dy())
	if iw <= 0 || ih <= 0 {
		return
	}
	scale := math.Min(float64(w)/iw, float64(h)/ih)
	dw := iw * scale
	dh := ih * scale

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x)+(float64(w)-dw)/2, float64(y)+(float64(h)-dh)/2)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

func (a *EditorApp) drawStageTopPanelInfo(screen *ebiten.Image, p theme.Palette) {
	a.drawStageToolbar(screen, p)
	if a.st.toolTab == "info" {
		a.drawStageInfo(screen, p)
	}
}

func (a *EditorApp) drawStageInfo(screen *ebiten.Image, p theme.Palette) {
	right := a.win.outsideWidth - panelPad
	y := a.topPanelY() + topPanelPad + 28
	name := a.proj.Stage.Name
	if name == "" {
		name = "Stage1"
	}
	info := fmt.Sprintf("Stage: %s    Objects: %d    Assets: %d", name, len(a.proj.Stage.Objects), len(a.proj.Stage.Assets))
	ui.DrawText(screen, info, right-ui.TextWidth(info), y, 1, p.LabelColor)
	y += 18
	hint := "Drag assets from the right panel onto the canvas. Pan: middle/right drag. Zoom: wheel."
	ui.DrawText(screen, hint, right-ui.TextWidth(hint), y, 1, p.TextMuted)
}

var stageToolTabs = []struct{ Key, Label string }{
	{"build", "Build"},
	{"edit", "Edit"},
	{"delete", "Delete"},
	{"info", "Info"},
}

var stageBuildToggles = []string{
	"col", "move", "stop", "pulse", "alpha", "toggle", "spawn", "rotate", "scale", "follow", "shake",
}

func (a *EditorApp) drawStageToolbar(screen *ebiten.Image, p theme.Palette) {
	a.st.toolTabRects = a.st.toolTabRects[:0]
	a.st.toolToggleRects = a.st.toolToggleRects[:0]

	const tabW, tabH, tabGap = 62, 20, 4
	const subW, subH, subGap = 52, 20, 3
	const sectionGap = 8

	right := a.win.outsideWidth - panelPad
	top := a.topPanelY() + topPanelPad
	tabsW := len(stageToolTabs)*tabW + (len(stageToolTabs)-1)*tabGap
	subsW := len(stageBuildToggles)*subW + (len(stageBuildToggles)-1)*subGap
	tabsX := right - (tabsW + sectionGap + subsW)

	x := tabsX
	mx, my := ebiten.CursorPosition()
	for _, tab := range stageToolTabs {
		active := a.st.toolTab == tab.Key
		bg := p.InputBG
		textClr := p.TextPrimary
		if active {
			bg = p.BtnBlue
			textClr = p.ButtonText
		}
		ui.FillRect(screen, x, top, tabW, tabH, bg)
		if mx >= x && mx <= x+tabW && my >= top && my <= top+tabH {
			ui.FillRect(screen, x, top, tabW, tabH, p.BtnHover)
		}
		ui.FillBorder(screen, x, top, tabW, tabH, p.InputBorder)
		ui.DrawTextCentered(screen, tab.Label, x+tabW/2, top+tabH/2, 1, textClr)
		a.st.toolTabRects = append(a.st.toolTabRects, stageToolRect{Key: tab.Key, X: x, Y: top, W: tabW, H: tabH})
		x += tabW + tabGap
	}

	if a.st.toolTab != "build" {
		return
	}

	ui.FillRect(screen, tabsX+tabsW+sectionGap/2, top, 1, tabH, p.InputBorder)
	x = tabsX + tabsW + sectionGap
	for _, label := range stageBuildToggles {
		on := a.st.toolToggles[label]
		bg := p.InputBG
		textClr := p.TextPrimary
		if on {
			bg = p.BtnGreen
			textClr = p.ButtonText
		}
		ui.FillRect(screen, x, top, subW, subH, bg)
		if mx >= x && mx <= x+subW && my >= top && my <= top+subH {
			ui.FillRect(screen, x, top, subW, subH, p.BtnHover)
		}
		ui.FillBorder(screen, x, top, subW, subH, p.InputBorder)
		ui.DrawTextCentered(screen, label, x+subW/2, top+subH/2, 1, textClr)
		a.st.toolToggleRects = append(a.st.toolToggleRects, stageToolRect{Key: label, X: x, Y: top, W: subW, H: subH})
		x += subW + subGap
	}
}

func (a *EditorApp) drawStageRightPanel(screen *ebiten.Image, p theme.Palette, px, iy int) int {
	a.st.categoryRects = a.st.categoryRects[:0]
	a.st.assetRects = a.st.assetRects[:0]
	mx, my := ebiten.CursorPosition()

	stage := &a.proj.Stage
	name := stage.Name
	if name == "" {
		name = "Stage1"
	}

	iy = a.drawSectionHeader(screen, p, px, iy, "Stage")
	ui.DrawText(screen, "Name: "+name, px, iy, 1, p.TextPrimary)
	iy += 16
	ui.DrawText(screen, fmt.Sprintf("Objects: %d", len(stage.Objects)), px, iy, 1, p.TextPrimary)
	iy += 18

	chkSize := 14
	chkY := iy + 3
	a.st.snapRect.X = px
	a.st.snapRect.Y = chkY
	a.st.snapRect.W = chkSize
	a.st.snapRect.H = chkSize
	ui.FillRect(screen, px, chkY, chkSize, chkSize, p.InputBG)
	if mx >= px && mx <= px+chkSize && my >= chkY && my <= chkY+chkSize {
		ui.FillRect(screen, px, chkY, chkSize, chkSize, p.BtnHover)
	}
	if a.st.snap {
		ui.FillRect(screen, px+3, chkY+3, chkSize-6, chkSize-6, p.BtnGreen)
	}
	ui.FillBorder(screen, px, chkY, chkSize, chkSize, p.InputBorder)
	ui.DrawText(screen, "Snap 50px", px+chkSize+6, iy+2, 1, p.TextPrimary)
	iy += rightPanelInputH + 4

	iy = a.drawSectionHeader(screen, p, px, iy, "Assets")

	categories := []string{project.StageCategoryBG, project.StageCategoryBlock}
	for _, cat := range categories {
		count := 0
		for _, asset := range stage.Assets {
			if stageAssetCategory(asset) == cat {
				count++
			}
		}

		headerH := 20
		a.st.categoryRects = append(a.st.categoryRects, stageCategoryRect{
			Category: cat,
			X:        px,
			Y:        iy,
			W:        rightPanelInner,
			H:        headerH,
		})
		if mx >= px && mx <= px+rightPanelInner && my >= iy && my <= iy+headerH {
			ui.FillRect(screen, px, iy, rightPanelInner, headerH, p.BtnHover)
		}
		arrow := "▸"
		if a.st.expanded[cat] {
			arrow = "▾"
		}
		ui.DrawText(screen, arrow, px+2, iy+2, 1, p.TextPrimary)
		ui.DrawText(screen, cat, px+16, iy+2, 1, p.TextPrimary)
		ui.DrawText(screen, fmt.Sprintf("(%d)", count), px+16+ui.TextWidth(cat)+6, iy+2, 1, p.TextMuted)
		iy += headerH

		if !a.st.expanded[cat] {
			continue
		}

		cardW := (rightPanelInner - 8) / 2
		cardH := 74
		col := 0
		cardX := px
		for assetIdx, asset := range stage.Assets {
			if stageAssetCategory(asset) != cat {
				continue
			}
			cardY := iy
			hovered := mx >= cardX && mx <= cardX+cardW && my >= cardY && my <= cardY+cardH
			ui.FillRect(screen, cardX, cardY, cardW, cardH, p.InputBG)
			border := p.InputBorder
			if hovered {
				border = p.InputFocusBorder
			}
			ui.FillBorder(screen, cardX, cardY, cardW, cardH, border)

			img := a.stageImage(asset.File)
			imgH := cardH - 26
			drawImageFit(screen, img, cardX+4, cardY+4, cardW-8, imgH)
			if img == nil {
				ui.DrawTextCentered(screen, "?", cardX+cardW/2, cardY+4+imgH/2, 1.5, p.TextMuted)
			}
			ui.DrawText(screen, ui.TruncateText(asset.Name, cardW-8), cardX+4, cardY+cardH-19, 1, p.TextPrimary)
			a.st.assetRects = append(a.st.assetRects, stageAssetRect{
				AssetIdx: assetIdx,
				X:        cardX,
				Y:        cardY,
				W:        cardW,
				H:        cardH,
			})

			col++
			if col == 2 {
				col = 0
				cardX = px
				iy += cardH + 6
			} else {
				cardX += cardW + 8
			}
		}
		if col != 0 {
			iy += cardH + 6
		}
	}

	iy += 4
	iy = a.drawSectionHeader(screen, p, px, iy, "Selected Object")
	obj := a.selectedStageObject()
	if obj == nil {
		ui.DrawText(screen, "No object selected.", px, iy, 1, p.TextMuted)
		for i := range a.props {
			a.props[i].Visible = false
		}
		if a.st.opacityInput != nil {
			a.st.opacityInput.Visible = false
		}
		if a.st.deleteBtn != nil {
			a.st.deleteBtn.Visible = false
		}
		if a.st.duplicateBtn != nil {
			a.st.duplicateBtn.Visible = false
		}
		return iy + 22
	}

	filename := ui.TruncateText(obj.Category+" / "+filepath.Base(obj.File), rightPanelInner)
	ui.DrawText(screen, filename, px, iy, 1, p.TextMuted)
	iy += 16

	for i := 0; i < 5; i++ {
		iy = a.drawLabeledInput(screen, a.props[i], iy)
	}
	if a.st.opacityInput != nil {
		iy = a.drawLabeledInput(screen, a.st.opacityInput, iy)
	}

	bw := (rightPanelInner - 4) / 2
	bh := 22
	if a.st.deleteBtn != nil {
		a.st.deleteBtn.X = px
		a.st.deleteBtn.Y = iy
		a.st.deleteBtn.W = bw
		a.st.deleteBtn.H = bh
		a.st.deleteBtn.Visible = true
		a.st.deleteBtn.Draw(screen)
	}
	if a.st.duplicateBtn != nil {
		a.st.duplicateBtn.X = px + bw + 4
		a.st.duplicateBtn.Y = iy
		a.st.duplicateBtn.W = bw
		a.st.duplicateBtn.H = bh
		a.st.duplicateBtn.Visible = true
		a.st.duplicateBtn.Draw(screen)
	}
	iy += bh + 6

	return iy
}
