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
	switch asset.Category {
	case project.StageCategoryBG, project.StageCategoryProp,
		project.StageCategoryAnimProp, project.StageCategoryAnimBlock:
		return asset.Category
	default:
		return project.StageCategoryBlock
	}
}

func stageCategoryLabel(category string) string {
	switch category {
	case project.StageCategoryBG:
		return "Backgrounds"
	case project.StageCategoryProp:
		return "Props"
	case project.StageCategoryAnimProp:
		return "Anim Props"
	case project.StageCategoryBlock:
		return "Blocks"
	case project.StageCategoryAnimBlock:
		return "Anim Blocks"
	default:
		return category
	}
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

func (a *EditorApp) hideStageObjectInputs() {
	for i := range a.props {
		if a.props[i] != nil {
			a.props[i].Visible = false
		}
	}
	if a.st.opacityInput != nil {
		a.st.opacityInput.Visible = false
	}
}

func (a *EditorApp) drawStageInfoInput(screen *ebiten.Image, inp *ui.TextInput, x, y, w int) {
	inp.Visible = true
	inp.X = x
	inp.Y = y
	inp.W = w
	inp.Draw(screen)
}

func (a *EditorApp) drawStageInfo(screen *ebiten.Image, p theme.Palette) {
	top := a.topPanelY() + topPanelPad
	x := a.stageToolContentX()
	width := a.win.outsideWidth - panelPad - x
	colGap := 10
	rowH := rightPanelInputH + 2

	name := a.proj.Stage.Name
	if name == "" {
		name = "Stage1"
	}
	obj := a.selectedStageObject()
	if obj == nil {
		a.hideStageObjectInputs()
		ui.DrawText(screen, fmt.Sprintf("Stage: %s    Objects: %d    Assets: %d", name, len(a.proj.Stage.Objects), len(a.proj.Stage.Assets)), x, top+2, 1, p.LabelColor)
		ui.DrawText(screen, "No object selected. Click an object to edit its properties here.", x, top+22, 1, p.TextMuted)
		ui.DrawText(screen, "CapsLock + drag moves. Tab + click or box selects. Wheel zooms.", x, top+42, 1, p.TextMuted)
		return
	}

	base := ui.TruncateText(filepath.Base(obj.File), 42)
	selected := fmt.Sprintf("Stage: %s    Objects: %d    Assets: %d", name, len(a.proj.Stage.Objects), len(a.proj.Stage.Assets))
	if count := a.stageSelectionCount(); count > 1 {
		selected = fmt.Sprintf("%d selected    Primary: %s", count, base)
	} else {
		selected = fmt.Sprintf("Selected: %s    %s", base, obj.Category)
	}
	ui.DrawText(screen, selected, x, top+2, 1, p.LabelColor)

	cols := 3
	if width < 620 {
		cols = 2
	}
	cellW := (width - (cols-1)*colGap) / cols
	if cellW < 80 {
		cellW = 80
	}
	inputs := []*ui.TextInput{a.props[0], a.props[1], a.props[2], a.props[3], a.props[4], a.st.opacityInput}
	for i, inp := range inputs {
		if inp == nil {
			continue
		}
		col := i % cols
		row := i / cols
		a.drawStageInfoInput(screen, inp, x+col*(cellW+colGap), top+22+row*rowH, cellW)
	}

	rows := (len(inputs) + cols - 1) / cols
	hintY := top + 22 + rows*rowH + 4
	ui.DrawText(screen, "CapsLock + drag moves. Tab + click or box selects. Wheel zooms.", x, hintY, 1, p.TextMuted)
}

var stageToolTabs = []struct{ Key, Label, Icon string }{
	{"build", "Build", "hammer"},
	{"edit", "Edit", "square-pen"},
	{"info", "Info", "info"},
}

const (
	stageToolTabCols    = 3
	stageToolTabW       = 62
	stageToolTabH       = 20
	stageToolTabGap     = 4
	stageToolTabIcon    = 13
	stageToolTabIconGap = 4

	stageToolSectionGap = 12
)

func stageToolTabsWidth() int {
	cols := stageToolTabCols
	if len(stageToolTabs) < cols {
		cols = len(stageToolTabs)
	}
	return cols*stageToolTabW + (cols-1)*stageToolTabGap
}

func stageToolTabsHeight() int {
	rows := (len(stageToolTabs) + stageToolTabCols - 1) / stageToolTabCols
	return rows*stageToolTabH + (rows-1)*stageToolTabGap
}

func stageToolTabOrigin(index, tabsX, top int) (int, int) {
	col := index % stageToolTabCols
	row := index / stageToolTabCols
	return tabsX + col*(stageToolTabW+stageToolTabGap), top + row*(stageToolTabH+stageToolTabGap)
}

// stageToolAreaX is where the stage toolbar starts: after the left-side
// project controls (mode dropdown and Open/Open Dir/Save/Theme buttons).
func (a *EditorApp) stageToolAreaX() int {
	return panelPad + dropdownW + panelPad
}

func (a *EditorApp) stageToolContentX() int {
	return a.stageToolAreaX() + stageToolTabsWidth() + stageToolSectionGap
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
	stageToolSubMinW = 70
	stageToolSubH    = 20
	stageToolSubGap  = 3

	// stageInfoHoldFrames keeps the auto-selected Info tab visible for a short
	// moment after the last live edit before returning to Edit.
	stageInfoHoldFrames = 45
)

// stageBuildColumns returns how many build toggles fit in width. The grid
// stretches its cells to fill the whole content area, so the toolbar uses the
// available horizontal space instead of clustering on the right.
func stageBuildColumns(width int) int {
	if width <= 0 {
		return 1
	}
	cols := (width + stageToolSubGap) / (stageToolSubMinW + stageToolSubGap)
	if cols < 1 {
		cols = 1
	}
	if cols > len(stageBuildToggles) {
		cols = len(stageBuildToggles)
	}
	return cols
}

func stageBuildCellWidth(width, cols int) int {
	if cols < 1 {
		cols = 1
	}
	cellW := (width - (cols-1)*stageToolSubGap) / cols
	if cellW < 1 {
		cellW = 1
	}
	return cellW
}

// stageBuildColumnWidth gives the last column the remainder so the grid ends
// exactly at the right edge of the content area.
func stageBuildColumnWidth(width, cols, col int) int {
	base := stageBuildCellWidth(width, cols)
	if col == cols-1 {
		return width - (cols-1)*(base+stageToolSubGap)
	}
	return base
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

func (a *EditorApp) drawStageEditActions(screen *ebiten.Image, p theme.Palette, x, y int) {
	if a.st.deleteBtn == nil || a.st.duplicateBtn == nil {
		return
	}
	hideCollision := func() {
		for _, btn := range []*ui.Button{a.st.collisionEditBtn, a.st.collisionToggleBtn} {
			if btn != nil {
				btn.Visible = false
			}
		}
	}
	obj := a.selectedStageObject()
	if obj == nil {
		a.st.deleteBtn.Visible = false
		a.st.duplicateBtn.Visible = false
		hideCollision()
		ui.DrawText(screen, "Select one or more objects on the canvas to edit them.", x, y+4, 1, p.TextMuted)
		return
	}

	buttons := []*ui.Button{a.st.deleteBtn, a.st.duplicateBtn}
	if project.StageCategoryHasCollision(obj.Category) {
		if a.st.collisionToggleBtn != nil {
			a.st.collisionToggleBtn.Text = "Collision: On"
			if obj.CollisionDisabled {
				a.st.collisionToggleBtn.Text = "Collision: Off"
			}
			a.st.collisionToggleBtn.BtnType = ui.BtnDefault
			if !obj.CollisionDisabled {
				a.st.collisionToggleBtn.BtnType = ui.BtnBlue
			}
		}
		if a.st.collisionEditBtn != nil {
			buttons = append(buttons, a.st.collisionEditBtn)
		}
		if a.st.collisionToggleBtn != nil {
			buttons = append(buttons, a.st.collisionToggleBtn)
		}
	} else {
		hideCollision()
	}

	maxRight := a.win.outsideWidth - panelPad
	cx, cy := x, y
	for _, btn := range buttons {
		if btn == nil {
			continue
		}
		btn.W = ui.ButtonWidth(btn.Text)
		btn.H = 22
		if cx > x && cx+btn.W > maxRight {
			cx = x
			cy += 26
		}
		btn.X = cx
		btn.Y = cy
		btn.Visible = true
		btn.Draw(screen)
		cx += btn.W + 4
	}
}

func (a *EditorApp) drawStageBuildToggles(screen *ebiten.Image, p theme.Palette, mx, my, x, top, width int) {
	if width <= 0 {
		return
	}
	cols := stageBuildColumns(width)
	cellW := stageBuildCellWidth(width, cols)
	for i, toggle := range stageBuildToggles {
		col := i % cols
		row := i / cols
		bx := x + col*(cellW+stageToolSubGap)
		by := top + row*(stageToolSubH+stageToolSubGap)

		on := a.st.toolToggles[toggle.Key]
		bg := p.InputBG
		textClr := p.TextPrimary
		if on {
			bg = p.BtnGreen
			textClr = p.ButtonText
		}
		ui.FillRect(screen, bx, by, cellW, stageToolSubH, bg)
		if mx >= bx && mx <= bx+cellW && my >= by && my <= by+stageToolSubH {
			ui.FillRect(screen, bx, by, cellW, stageToolSubH, p.BtnHover)
		}
		ui.FillBorder(screen, bx, by, cellW, stageToolSubH, p.InputBorder)
		a.drawStageToolContent(screen, bx, by, cellW, stageToolSubH, toggle.Label, toggle.Icon, textClr)
		a.st.toolToggleRects = append(a.st.toolToggleRects, stageToolRect{Key: toggle.Key, X: bx, Y: by, W: cellW, H: stageToolSubH})
	}
}

func (a *EditorApp) drawStageToolbar(screen *ebiten.Image, p theme.Palette) {
	if a.st.toolTab != "info" {
		a.hideStageObjectInputs()
	}
	a.st.toolTabRects = a.st.toolTabRects[:0]
	a.st.toolToggleRects = a.st.toolToggleRects[:0]
	if a.st.deleteBtn != nil {
		a.st.deleteBtn.Visible = false
	}
	if a.st.duplicateBtn != nil {
		a.st.duplicateBtn.Visible = false
	}
	for _, btn := range []*ui.Button{a.st.collisionEditBtn, a.st.collisionToggleBtn, a.st.collisionBackBtn} {
		if btn != nil {
			btn.Visible = false
		}
	}
	if a.st.collisionTable != nil {
		a.st.collisionTable.Visible = false
	}

	right := a.win.outsideWidth - panelPad
	top := a.topPanelY() + topPanelPad
	tabsX := a.stageToolAreaX()
	contentX := a.stageToolContentX()
	tabsH := stageToolTabsHeight()

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

	ui.FillRect(screen, contentX-stageToolSectionGap/2, top, 1, tabsH, p.InputBorder)

	switch a.st.toolTab {
	case "edit":
		if a.st.collisionEdit && a.selectedStageObject() != nil {
			a.drawStageCollisionEditor(screen, p, contentX, top, right)
		} else {
			a.drawStageEditActions(screen, p, contentX, top)
		}
	case "build":
		a.drawStageBuildToggles(screen, p, mx, my, contentX, top, right-contentX)
	}
}

func (a *EditorApp) drawStageLayers(screen *ebiten.Image, p theme.Palette, px, iy int) int {
	a.st.layerRects = a.st.layerRects[:0]
	mx, my := ebiten.CursorPosition()
	iy = a.drawSectionHeader(screen, p, px, iy, "Layers")

	a.ensureActiveStageLayer()
	activeIdx := a.activeStageLayerIndex()
	general := a.st.activeLayer == ""
	if general {
		ui.DrawText(screen, "All layers visible. New objects: "+a.stageWriteLayer(), px, iy, 1, p.TextMuted)
	} else {
		ui.DrawText(screen, "Others dimmed. Top renders in front.", px, iy, 1, p.TextMuted)
	}
	iy += 16

	if a.st.layerAddBtn != nil && a.st.layerDeleteBtn != nil && a.st.layerUpBtn != nil && a.st.layerDownBtn != nil {
		buttons := []*ui.Button{a.st.layerAddBtn, a.st.layerDeleteBtn, a.st.layerUpBtn, a.st.layerDownBtn}
		a.st.layerAddBtn.Visible = true
		a.st.layerDeleteBtn.Visible = true
		a.st.layerUpBtn.Visible = true
		a.st.layerDownBtn.Visible = true
		a.st.layerDeleteBtn.Enabled = !general && len(a.proj.Stage.Layers) > 1
		a.st.layerUpBtn.Enabled = !general && activeIdx < len(a.proj.Stage.Layers)-1
		a.st.layerDownBtn.Enabled = !general && activeIdx > 0

		x := px
		for _, btn := range buttons {
			btn.X = x
			btn.Y = iy
			btn.W = ui.ButtonWidth(btn.Text)
			btn.H = 20
			btn.Draw(screen)
			x += btn.W + 4
		}
	}
	iy += 24

	if len(a.proj.Stage.Layers) == 0 {
		ui.DrawText(screen, "No layers yet.", px, iy, 1, p.TextMuted)
		iy += 22
	} else {
		drawRow := func(index int, label string, active bool, hovered bool) {
			rowY := iy
			if active {
				ui.FillRect(screen, px, rowY, rightPanelInner, 18, p.BtnBlue)
			} else if hovered {
				ui.FillRect(screen, px, rowY, rightPanelInner, 18, p.BtnHover)
			}
			ui.FillBorder(screen, px, rowY, rightPanelInner, 18, p.InputBorder)
			textClr := p.TextPrimary
			if active {
				textClr = p.ButtonText
			}
			ui.DrawText(screen, ui.TruncateText(label, rightPanelInner-8), px+4, rowY+2, 1, textClr)
			a.st.layerRects = append(a.st.layerRects, stageLayerRect{Index: index, X: px, Y: rowY, W: rightPanelInner, H: 18})
			iy += 20
		}

		generalHovered := mx >= px && mx <= px+rightPanelInner && my >= iy && my <= iy+18
		drawRow(-1, fmt.Sprintf("General  (%d)", len(a.proj.Stage.Objects)), general, generalHovered)

		for idx := len(a.proj.Stage.Layers) - 1; idx >= 0; idx-- {
			layer := a.proj.Stage.Layers[idx]
			rowY := iy
			hovered := mx >= px && mx <= px+rightPanelInner && my >= rowY && my <= rowY+18
			count := 0
			for _, obj := range a.proj.Stage.Objects {
				if a.proj.Stage.LayerNameFor(obj) == layer.Name {
					count++
				}
			}
			drawRow(idx, fmt.Sprintf("%s  (%d)", layer.Name, count), !general && idx == activeIdx, hovered)
		}
	}

	if btn := a.st.collisionViewBtn; btn != nil {
		btn.Text = "Show collisions"
		btn.BtnType = ui.BtnDefault
		if a.st.showCollisions {
			btn.Text = "Hide collisions"
			btn.BtnType = ui.BtnBlue
		}
		btn.X = px
		btn.Y = iy
		btn.W = ui.ButtonWidth(btn.Text)
		btn.H = 20
		btn.Visible = true
		btn.Draw(screen)
		iy += 24
	}
	return iy + 4
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

	if a.st.stageWidthInput != nil && a.st.stageHeightInput != nil {
		gap := 6
		width := (rightPanelInner - gap) / 2
		a.st.stageWidthInput.X = px
		a.st.stageWidthInput.Y = iy
		a.st.stageWidthInput.W = width
		a.st.stageWidthInput.H = rightPanelInputH
		a.st.stageWidthInput.Draw(screen)

		a.st.stageHeightInput.X = px + width + gap
		a.st.stageHeightInput.Y = iy
		a.st.stageHeightInput.W = rightPanelInner - width - gap
		a.st.stageHeightInput.H = rightPanelInputH
		a.st.stageHeightInput.Draw(screen)
		iy += rightPanelInputH + 4
	}

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

	categories := []string{
		project.StageCategoryBG,
		project.StageCategoryProp,
		project.StageCategoryAnimProp,
		project.StageCategoryBlock,
		project.StageCategoryAnimBlock,
	}
	for _, cat := range categories {
		count := 0
		for _, asset := range stage.Assets {
			if stageAssetCategory(asset) == cat {
				count++
			}
		}

		label := stageCategoryLabel(cat)
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
		ui.DrawText(screen, label, px+16, iy+2, 1, p.TextPrimary)
		ui.DrawText(screen, fmt.Sprintf("(%d)", count), px+16+ui.TextWidth(label)+6, iy+2, 1, p.TextMuted)
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

			img := a.stageAssetThumb(asset)
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
	iy = a.drawStageLayers(screen, p, px, iy)
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
