package editor

import (
	"fmt"
	"image/color"
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
	y := a.topPanelY() + topPanelPad + stageToolTabsHeight() + 8
	name := a.proj.Stage.Name
	if name == "" {
		name = "Stage1"
	}
	info := fmt.Sprintf("Stage: %s    Objects: %d    Assets: %d", name, len(a.proj.Stage.Objects), len(a.proj.Stage.Assets))
	ui.DrawText(screen, info, right-ui.TextWidth(info), y, 1, p.LabelColor)
	y += 18

	obj := a.selectedStageObject()
	if obj != nil {
		base := ui.TruncateText(filepath.Base(obj.File), 52)
		selected := fmt.Sprintf("Selected: %s    Category: %s", base, obj.Category)
		if count := a.stageSelectionCount(); count > 1 {
			selected = fmt.Sprintf("%d objects selected    Primary: %s    Category: %s", count, base, obj.Category)
		}
		ui.DrawText(screen, selected, right-ui.TextWidth(selected), y, 1, p.TextPrimary)
		y += 18
		values := fmt.Sprintf(
			"X %g    Y %g    W %g    H %g    Rot %g°    Opacity %.2f",
			obj.X, obj.Y, obj.Width, obj.Height, obj.Rotation, obj.Opacity,
		)
		ui.DrawText(screen, values, right-ui.TextWidth(values), y, 1, p.TextMuted)
		y += 18
	}

	hint := "Move: hold CapsLock + drag. Multi-select: hold Tab + click or drag a box. Pan: middle/right. Zoom: wheel."
	ui.DrawText(screen, hint, right-ui.TextWidth(hint), y, 1, p.TextMuted)
}

var stageToolTabs = []struct{ Key, Label, Icon string }{
	{"build", "Build", "hammer"},
	{"edit", "Edit", "square-pen"},
	{"info", "Info", "info"},
}

const (
	stageToolTabCols    = 2
	stageToolTabW       = 62
	stageToolTabH       = 20
	stageToolTabGap     = 4
	stageToolTabIcon    = 13
	stageToolTabIconGap = 4
)

func stageToolTabsHeight() int {
	rows := (len(stageToolTabs) + stageToolTabCols - 1) / stageToolTabCols
	return rows*stageToolTabH + (rows-1)*stageToolTabGap
}

func stageToolTabOrigin(index, tabsX, top int) (int, int) {
	col := index % stageToolTabCols
	row := index / stageToolTabCols
	return tabsX + col*(stageToolTabW+stageToolTabGap), top + row*(stageToolTabH+stageToolTabGap)
}

var stageBuildToggles = []struct{ Key, Label, Icon string }{
	{"col", "Col", "palette"},
	{"move", "Move", "move"},
	{"stop", "Stop", "ban"},
	{"pulse", "Pulse", "heart-pulse"},
	{"alpha", "Alpha", "opacity"},
	{"toggle", "Toggle", "toggle-right"},
	{"spawn", "Spawn", "star"},
	{"rotate", "Rotate", "refresh-cw"},
	{"scale", "Scale", "square-dimensions"},
	{"follow", "Follow", "circle-chevrons-right"},
	{"shake", "Shake", "activity"},
}

const (
	stageToolSubW       = 61
	stageToolSubH       = 20
	stageToolSubGap     = 3
	stageToolPageSize   = 6
	stageToolChevronW   = 18
	stageToolChevronGap = 2

	// stageInfoHoldFrames keeps the auto-selected Info tab visible for a short
	// moment after the last live edit before returning to Edit.
	stageInfoHoldFrames = 45
)

func stageToolPageCount() int {
	if len(stageBuildToggles) == 0 {
		return 1
	}
	return (len(stageBuildToggles) + stageToolPageSize - 1) / stageToolPageSize
}

func stageToolPageBounds(page int) (start, end int) {
	pages := stageToolPageCount()
	if page < 0 {
		page = 0
	}
	if page >= pages {
		page = pages - 1
	}
	start = page * stageToolPageSize
	end = start + stageToolPageSize
	if end > len(stageBuildToggles) {
		end = len(stageBuildToggles)
	}
	return start, end
}

func (a *EditorApp) drawStageToolContent(screen *ebiten.Image, x, y, w, h int, label, icon string, clr color.Color) {
	labelW := ui.TextWidth(label)
	contentW := labelW
	hasIcon := a.iconSet.Has(icon)
	if hasIcon {
		contentW += stageToolTabIcon + stageToolTabIconGap
	}
	contentX := x + (w-contentW)/2
	if hasIcon {
		iconY := y + (h-stageToolTabIcon)/2
		a.iconSet.Draw(screen, icon, contentX, iconY, stageToolTabIcon, stageToolTabIcon, clr)
		contentX += stageToolTabIcon + stageToolTabIconGap
	}
	labelY := y + (h-ui.TextHeight())/2
	ui.DrawText(screen, label, contentX, labelY, 1, clr)
}

func (a *EditorApp) drawStagePageButton(screen *ebiten.Image, p theme.Palette, mx, my, x, y, w, h int, icon string, enabled bool) {
	hovered := enabled && mx >= x && mx <= x+w && my >= y && my <= y+h
	textClr := p.TextMuted
	if enabled {
		textClr = p.TextPrimary
	}
	ui.FillRect(screen, x, y, w, h, p.InputBG)
	if hovered {
		ui.FillRect(screen, x, y, w, h, p.BtnHover)
	}
	ui.FillBorder(screen, x, y, w, h, p.InputBorder)
	if a.iconSet.Draw(screen, icon, x, y, w, h, textClr) {
		return
	}
	fallback := "<"
	if icon == "chevron-right" {
		fallback = ">"
	}
	ui.DrawTextCentered(screen, fallback, x+w/2, y+h/2, 1, textClr)
}

func (a *EditorApp) drawStageEditActions(screen *ebiten.Image, p theme.Palette, x, y int) {
	if a.st.deleteBtn == nil || a.st.duplicateBtn == nil {
		return
	}
	if a.selectedStageObject() == nil {
		a.st.deleteBtn.Visible = false
		a.st.duplicateBtn.Visible = false
		ui.DrawText(screen, "Select one or more objects on the canvas to edit them.", x, y+4, 1, p.TextMuted)
		return
	}

	a.st.deleteBtn.X = x
	a.st.deleteBtn.Y = y
	a.st.deleteBtn.W = ui.ButtonWidth(a.st.deleteBtn.Text)
	a.st.deleteBtn.H = 22
	a.st.deleteBtn.Visible = true
	a.st.deleteBtn.Draw(screen)

	a.st.duplicateBtn.X = a.st.deleteBtn.X + a.st.deleteBtn.W + 4
	a.st.duplicateBtn.Y = y
	a.st.duplicateBtn.W = ui.ButtonWidth(a.st.duplicateBtn.Text)
	a.st.duplicateBtn.H = 22
	a.st.duplicateBtn.Visible = true
	a.st.duplicateBtn.Draw(screen)
}

func (a *EditorApp) drawStageToolbar(screen *ebiten.Image, p theme.Palette) {
	a.st.toolTabRects = a.st.toolTabRects[:0]
	a.st.toolToggleRects = a.st.toolToggleRects[:0]
	a.st.toolPagePrevRect = stageToolRect{}
	a.st.toolPageNextRect = stageToolRect{}
	if a.st.deleteBtn != nil {
		a.st.deleteBtn.Visible = false
	}
	if a.st.duplicateBtn != nil {
		a.st.duplicateBtn.Visible = false
	}

	const sectionGap = 8

	right := a.win.outsideWidth - panelPad
	top := a.topPanelY() + topPanelPad
	cols := stageToolTabCols
	if len(stageToolTabs) < cols {
		cols = len(stageToolTabs)
	}
	tabsW := cols*stageToolTabW + (cols-1)*stageToolTabGap
	tabsH := stageToolTabsHeight()
	pages := stageToolPageCount()
	if a.st.toolPage < 0 {
		a.st.toolPage = 0
	}
	if a.st.toolPage >= pages {
		a.st.toolPage = pages - 1
	}
	pageW := stageToolPageSize*stageToolSubW + (stageToolPageSize-1)*stageToolSubGap
	subsW := pageW
	if pages > 1 {
		subsW += 2*stageToolChevronW + 2*stageToolChevronGap
	}
	tabsX := right - (tabsW + sectionGap + subsW)

	mx, my := ebiten.CursorPosition()
	for i, tab := range stageToolTabs {
		tx, ty := stageToolTabOrigin(i, tabsX, top)

		active := a.st.toolTab == tab.Key
		bg := p.InputBG
		textClr := p.TextPrimary
		if active {
			bg = p.BtnBlue
			textClr = p.ButtonText
		}
		ui.FillRect(screen, tx, ty, stageToolTabW, stageToolTabH, bg)
		if mx >= tx && mx <= tx+stageToolTabW && my >= ty && my <= ty+stageToolTabH {
			ui.FillRect(screen, tx, ty, stageToolTabW, stageToolTabH, p.BtnHover)
		}
		ui.FillBorder(screen, tx, ty, stageToolTabW, stageToolTabH, p.InputBorder)
		a.drawStageToolContent(screen, tx, ty, stageToolTabW, stageToolTabH, tab.Label, tab.Icon, textClr)
		a.st.toolTabRects = append(a.st.toolTabRects, stageToolRect{Key: tab.Key, X: tx, Y: ty, W: stageToolTabW, H: stageToolTabH})
	}

	if a.st.toolTab == "edit" {
		a.drawStageEditActions(screen, p, tabsX, top+tabsH+6)
		return
	}
	if a.st.toolTab != "build" {
		return
	}

	ui.FillRect(screen, tabsX+tabsW+sectionGap/2, top, 1, tabsH, p.InputBorder)
	x := tabsX + tabsW + sectionGap
	if pages > 1 {
		a.drawStagePageButton(screen, p, mx, my, x, top, stageToolChevronW, stageToolSubH, "chevron-left", a.st.toolPage > 0)
		a.st.toolPagePrevRect = stageToolRect{Key: "prev", X: x, Y: top, W: stageToolChevronW, H: stageToolSubH}
		x += stageToolChevronW + stageToolChevronGap
	}

	start, end := stageToolPageBounds(a.st.toolPage)
	buttonsX := x
	for _, toggle := range stageBuildToggles[start:end] {
		on := a.st.toolToggles[toggle.Key]
		bg := p.InputBG
		textClr := p.TextPrimary
		if on {
			bg = p.BtnGreen
			textClr = p.ButtonText
		}
		ui.FillRect(screen, x, top, stageToolSubW, stageToolSubH, bg)
		if mx >= x && mx <= x+stageToolSubW && my >= top && my <= top+stageToolSubH {
			ui.FillRect(screen, x, top, stageToolSubW, stageToolSubH, p.BtnHover)
		}
		ui.FillBorder(screen, x, top, stageToolSubW, stageToolSubH, p.InputBorder)
		a.drawStageToolContent(screen, x, top, stageToolSubW, stageToolSubH, toggle.Label, toggle.Icon, textClr)
		a.st.toolToggleRects = append(a.st.toolToggleRects, stageToolRect{Key: toggle.Key, X: x, Y: top, W: stageToolSubW, H: stageToolSubH})
		x += stageToolSubW + stageToolSubGap
	}

	if pages > 1 {
		nextX := buttonsX + pageW + stageToolChevronGap
		a.drawStagePageButton(screen, p, mx, my, nextX, top, stageToolChevronW, stageToolSubH, "chevron-right", a.st.toolPage < pages-1)
		a.st.toolPageNextRect = stageToolRect{Key: "next", X: nextX, Y: top, W: stageToolChevronW, H: stageToolSubH}
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
	header := "Object Properties"
	if count := a.stageSelectionCount(); count > 1 {
		header = fmt.Sprintf("Object Properties (%d selected)", count)
	}
	iy = a.drawSectionHeader(screen, p, px, iy, header)
	obj := a.selectedStageObject()
	if obj == nil {
		ui.DrawText(screen, "No object selected.", px, iy, 1, p.TextMuted)
		for i := range a.props {
			a.props[i].Visible = false
		}
		if a.st.opacityInput != nil {
			a.st.opacityInput.Visible = false
		}
		return iy + 22
	}

	for i := 0; i < 5; i++ {
		iy = a.drawLabeledInput(screen, a.props[i], iy)
	}
	if a.st.opacityInput != nil {
		iy = a.drawLabeledInput(screen, a.st.opacityInput, iy)
	}

	iy += 4
	return iy
}

func (a *EditorApp) drawStageMarquee(screen *ebiten.Image, p theme.Palette) {
	if !a.st.marqueeActive {
		return
	}
	x1, y1 := a.canvas.Cam.WorldToCanvas(a.st.marqueeStartWX, a.st.marqueeStartWY)
	x2, y2 := a.canvas.Cam.WorldToCanvas(a.st.marqueeCurWX, a.st.marqueeCurWY)
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	left := a.canvas.X + int(math.Round(x1))
	top := a.canvas.Y + int(math.Round(y1))
	right := a.canvas.X + int(math.Round(x2))
	bottom := a.canvas.Y + int(math.Round(y2))
	left = max(left, a.canvas.X)
	top = max(top, a.canvas.Y)
	right = min(right, a.canvas.X+a.canvas.Width)
	bottom = min(bottom, a.canvas.Y+a.canvas.Height)
	if right <= left || bottom <= top {
		return
	}
	ui.FillRect(screen, left, top, right-left, bottom-top, color.RGBA{110, 170, 255, 36})
	ui.FillBorder(screen, left, top, right-left, bottom-top, p.CanvasRotationGuide)
}
