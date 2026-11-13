package editor

import (
	"math"

	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (a *EditorApp) clearStageInputFocus() {
	if a.stageSizeInputFocused() {
		a.commitStageSizeInputs()
	}
	for _, inp := range a.stageSizeInputs() {
		inp.Focused = false
	}
	for i := range a.props {
		if a.props[i] != nil {
			a.props[i].Focused = false
		}
	}
	if a.st.opacityInput != nil {
		a.st.opacityInput.Focused = false
	}
	for _, inp := range a.stageCollisionInputList() {
		if inp != nil {
			inp.Focused = false
			inp.Visible = false
		}
	}
}

func (a *EditorApp) stageSizeInputs() []*ui.TextInput {
	if a.st.stageWidthInput == nil || a.st.stageHeightInput == nil {
		return nil
	}
	return []*ui.TextInput{a.st.stageWidthInput, a.st.stageHeightInput}
}

func (a *EditorApp) stageSizeInputFocused() bool {
	for _, inp := range a.stageSizeInputs() {
		if inp.Focused {
			return true
		}
	}
	return false
}

func (a *EditorApp) isStageSizeInput(inp *ui.TextInput) bool {
	return inp != nil && (inp == a.st.stageWidthInput || inp == a.st.stageHeightInput)
}

func (a *EditorApp) syncStageSizeInputs() {
	for _, inp := range a.stageSizeInputs() {
		if inp.Focused {
			continue
		}
		if inp == a.st.stageWidthInput {
			inp.SetNumeric(a.proj.Stage.Width)
		} else {
			inp.SetNumeric(a.proj.Stage.Height)
		}
	}
}

func (a *EditorApp) commitStageSizeInputs() {
	widthInput, heightInput := a.st.stageWidthInput, a.st.stageHeightInput
	if widthInput == nil || heightInput == nil {
		return
	}

	width := math.Round(widthInput.NumericValue())
	height := math.Round(heightInput.NumericValue())
	if width <= 0 {
		width = a.proj.Stage.Width
	}
	if height <= 0 {
		height = a.proj.Stage.Height
	}
	widthInput.SetNumeric(width)
	heightInput.SetNumeric(height)

	if width == a.proj.Stage.Width && height == a.proj.Stage.Height {
		return
	}

	a.saveSnapshot()
	a.proj.Stage.Width = width
	a.proj.Stage.Height = height
	a.setStageCanvasBounds()
}

func (a *EditorApp) pointInCanvas(mx, my int) bool {
	return mx >= a.canvas.X && mx <= a.canvas.X+a.canvas.Width &&
		my >= a.canvas.Y && my <= a.canvas.Y+a.canvas.Height
}

func (a *EditorApp) canvasWorld(mx, my int) (float64, float64) {
	return a.canvas.Cam.CanvasToWorld(float64(mx-a.canvas.X), float64(my-a.canvas.Y))
}

func (a *EditorApp) panCanvasByMouse(mx, my int) {
	a.canvas.Cam.Pan(float64(mx-a.prevMouseX), float64(my-a.prevMouseY))
}

func stageInfoInputHit(inp *ui.TextInput, mx, my int) bool {
	if inp == nil || !inp.Visible {
		return false
	}
	ix := inp.InputX()
	return mx >= ix && mx <= ix+inp.InputW() && my >= inp.Y && my <= inp.Y+inp.H
}

func (a *EditorApp) handleStageToolbarMouse(mx, my int, justL bool) {
	if a.st.collisionEdit {
		if a.handleStageCollisionMouse(mx, my, justL) {
			return
		}
	}
	if a.st.toolTab == "edit" && a.selectedStageObject() != nil {
		for _, btn := range []*ui.Button{a.st.deleteBtn, a.st.duplicateBtn, a.st.collisionEditBtn, a.st.collisionToggleBtn} {
			if btn == nil {
				continue
			}
			hit := justL && btn.HitTest(mx, my)
			btn.HandleMouse(mx, my, justL)
			if hit {
				return
			}
		}
	}
	if a.st.toolTab == "info" && a.selectedStageObject() != nil {
		focusedBefore := a.focusedInput()
		hitInput := false
		for i := range a.props {
			if stageInfoInputHit(a.props[i], mx, my) {
				hitInput = true
			}
			a.props[i].HandleMouse(mx, my, justL)
		}
		if a.st.opacityInput != nil {
			if stageInfoInputHit(a.st.opacityInput, mx, my) {
				hitInput = true
			}
			a.st.opacityInput.HandleMouse(mx, my, justL)
		}
		if focused := a.focusedInput(); focused != nil && focused != focusedBefore {
			a.markStageObjectLiveEdit()
		}
		if justL && hitInput {
			return
		}
	}

	if !justL {
		return
	}
	for _, r := range a.st.toolTabRects {
		if mx >= r.X && mx <= r.X+r.W && my >= r.Y && my <= r.Y+r.H {
			a.clearStageInputFocus()
			if a.st.collisionEdit {
				a.st.collisionEdit = false
				if a.st.collisionTable != nil {
					a.st.collisionTable.Visible = false
				}
				a.syncLayout()
			}
			a.st.toolTab = r.Key
			a.st.toolInfoAuto = false
			a.st.toolInfoHold = 0
			return
		}
	}
	for _, r := range a.st.toolToggleRects {
		if mx >= r.X && mx <= r.X+r.W && my >= r.Y && my <= r.Y+r.H {
			if a.st.toolToggles == nil {
				a.st.toolToggles = make(map[string]bool)
			}
			a.st.toolToggles[r.Key] = !a.st.toolToggles[r.Key]
			return
		}
	}
}

func (a *EditorApp) handleStagePanelMouse(mx, my int, justL bool) {
	if mx < a.win.outsideWidth-rightPanelW ||
		my < a.rightPanelY() || my >= a.rightPanelY()+a.rightPanelH() {
		return
	}

	focusedBefore := a.stageSizeInputFocused()
	sizeInputHit := false
	for _, inp := range a.stageSizeInputs() {
		if stageInfoInputHit(inp, mx, my) {
			sizeInputHit = true
		}
		inp.HandleMouse(mx, my, justL)
	}
	if justL && focusedBefore && !a.stageSizeInputFocused() {
		a.commitStageSizeInputs()
	}
	if justL && sizeInputHit {
		return
	}

	for _, btn := range []*ui.Button{a.st.layerAddBtn, a.st.layerDeleteBtn, a.st.layerUpBtn, a.st.layerDownBtn, a.st.collisionViewBtn} {
		if btn == nil {
			continue
		}
		hit := justL && btn.HitTest(mx, my)
		btn.HandleMouse(mx, my, justL)
		if hit {
			return
		}
	}

	if !justL {
		return
	}

	for _, row := range a.st.layerRects {
		if mx >= row.X && mx <= row.X+row.W && my >= row.Y && my <= row.Y+row.H {
			a.clearStageInputFocus()
			if row.Index < 0 {
				a.setGeneralStageView()
			} else {
				a.setActiveStageLayer(row.Index)
			}
			return
		}
	}

	chk := a.st.snapRect
	if chk.W > 0 && mx >= chk.X && mx <= chk.X+chk.W && my >= chk.Y && my <= chk.Y+chk.H {
		a.st.snap = !a.st.snap
		return
	}

	for _, cat := range a.st.categoryRects {
		if mx >= cat.X && mx <= cat.X+cat.W && my >= cat.Y && my <= cat.Y+cat.H {
			if a.st.expanded == nil {
				a.st.expanded = map[string]bool{}
			}
			a.st.expanded[cat.Category] = !a.st.expanded[cat.Category]
			return
		}
	}

	for _, ar := range a.st.assetRects {
		if mx >= ar.X && mx <= ar.X+ar.W && my >= ar.Y && my <= ar.Y+ar.H {
			a.clearStageInputFocus()
			a.st.dragAsset = ar.AssetIdx
			a.st.draggingFromPanel = false
			a.st.dragStartX = mx
			a.st.dragStartY = my
			cw := float64(a.canvas.Width)
			ch := float64(a.canvas.Height)
			a.st.ghostWX, a.st.ghostWY = a.canvas.Cam.CanvasToWorld(cw/2, ch/2)
			return
		}
	}
}

func stageMoveModifierDown() bool {
	return ebiten.IsKeyPressed(ebiten.KeyCapsLock)
}

func (a *EditorApp) handleStageCanvasMouse(mx, my int, leftDown, justPressed bool) {
	if a.st.dragAsset >= 0 {
		return
	}

	middleDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)
	rightDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	spaceDown := ebiten.IsKeyPressed(ebiten.KeySpace)
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	tab := ebiten.IsKeyPressed(ebiten.KeyTab)

	if a.st.panning {
		if leftDown || middleDown || rightDown {
			a.panCanvasByMouse(mx, my)
			return
		}
		a.st.panning = false
	}

	if middleDown || rightDown || (spaceDown && leftDown) {
		if a.pointInCanvas(mx, my) {
			a.st.panning = true
			a.st.stagePointerOnObject = false
			a.st.marqueeActive = false
			a.scaleHandle = -1
			a.st.draggingObj = false
			a.st.dragUndoSaved = false
			a.panCanvasByMouse(mx, my)
		}
		return
	}

	if !leftDown {
		if a.st.draggingObj && a.st.dragMoved {
			if obj := a.selectedStageObject(); obj != nil {
				wx, wy := a.canvasWorld(mx, my)
				newX, newY := dragStagePosition(
					a.st.dragObjStartX, a.st.dragObjStartY,
					a.st.dragWorldStartX, a.st.dragWorldStartY,
					wx, wy, obj.Width, obj.Height, a.st.snap,
				)
				a.applyStageObjectDrag(newX, newY)
			}
		}
		a.finishStageMarquee()
		a.scaleHandle = -1
		a.rotationDrag.active = false
		a.st.draggingObj = false
		a.st.dragMoved = false
		a.st.dragUndoSaved = false
		a.st.dragObjStarts = a.st.dragObjStarts[:0]
		a.st.stagePointerOnObject = false
		return
	}
	if a.st.marqueeActive {
		wx, wy := a.canvasWorld(mx, my)
		a.updateStageMarquee(wx, wy)
		return
	}
	if !a.pointInCanvas(mx, my) {
		return
	}

	if a.rotationDrag.active {
		a.applyRotationDrag(mx, my)
		return
	}

	if justPressed {
		a.clearStageInputFocus()
		a.st.draggingObj = false
		a.st.dragUndoSaved = false
		a.st.panning = false

		if shift && a.canvas.Selection.Visible && a.canvas.RotationHandleHitTest(mx, my) {
			a.syncPanelToSelection()
			if a.beginRotationDrag(mx, my) {
				a.saveSnapshot()
				return
			}
		}

		if ctrl && a.canvas.Selection.Visible {
			if h := a.canvas.HandleHitTest(mx, my); h >= 0 {
				a.saveSnapshot()
				a.scaleHandle = h
				a.panelMode = panelModeStage
				a.initScaleOrig(h)
				return
			}
		}

		wx, wy := a.canvasWorld(mx, my)
		if idx := a.stageObjectAtWorld(wx, wy); idx >= 0 {
			a.st.stagePointerOnObject = true
			switch {
			case tab || ctrl:
				a.toggleStageObjectSelection(idx)
				return
			case shift:
				a.addStageObjectSelection(idx)
			case !a.stageObjectSelected(idx):
				a.selectStageObject(idx)
			default:
				a.setStageObjectPrimary(idx)
			}
			a.beginStageObjectDrag(wx, wy)
			return
		}
		if tab {
			a.beginStageMarquee(wx, wy)
			return
		}
		a.selectStageObject(-1)
		a.st.stagePointerOnObject = false
		a.st.panning = true
		return
	}

	if a.scaleHandle >= 0 && ctrl {
		a.applyHandleScale(mx, my)
		return
	}

	if a.st.draggingObj {
		if !stageMoveModifierDown() {
			return
		}
		obj := a.selectedStageObject()
		if obj == nil {
			a.st.draggingObj = false
			return
		}
		wx, wy := a.canvasWorld(mx, my)
		newX, newY := dragStagePosition(
			a.st.dragObjStartX, a.st.dragObjStartY,
			a.st.dragWorldStartX, a.st.dragWorldStartY,
			wx, wy, obj.Width, obj.Height, a.st.snap,
		)
		a.applyStageObjectDrag(newX, newY)
		return
	}

	if a.st.stagePointerOnObject {
		return
	}
	a.panCanvasByMouse(mx, my)
}

func (a *EditorApp) handleStageCanvasWheel(mx, my int, wy float64, ctrl, shift bool) {
	if !a.pointInCanvas(mx, my) {
		return
	}
	obj := a.selectedStageObject()
	switch {
	case shift && obj != nil:
		obj.Rotation = math.Round(obj.Rotation + wy*5)
		a.props[4].SetNumeric(obj.Rotation)
		a.markStageObjectLiveEdit()
	case ctrl && obj != nil:
		factor := 1 + wy*0.05
		if factor < 0.01 {
			factor = 0.01
		}
		obj.Width = math.Max(1, math.Round(obj.Width*factor*100)/100)
		obj.Height = math.Max(1, math.Round(obj.Height*factor*100)/100)
		a.props[2].SetNumeric(obj.Width)
		a.props[3].SetNumeric(obj.Height)
		a.markStageObjectLiveEdit()
	default:
		a.canvas.Cam.ZoomAt(1+wy*0.1, float64(mx-a.canvas.X), float64(my-a.canvas.Y))
	}
}

func (a *EditorApp) handleStageKeys() {
	if a.focusedInput() != nil {
		return
	}
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyA) {
		a.selectAllStageObjects()
	}
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyD) {
		a.duplicateSelectedStageObject()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		a.deleteSelectedStageObject()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.st.marqueeActive = false
		a.selectStageObject(-1)
	}
}
