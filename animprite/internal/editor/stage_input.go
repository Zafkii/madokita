package editor

import (
	"math"

	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (a *EditorApp) clearStageInputFocus() {
	for i := range a.props {
		a.props[i].Focused = false
	}
	if a.st.opacityInput != nil {
		a.st.opacityInput.Focused = false
	}
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

func (a *EditorApp) handleStageToolbarMouse(mx, my int, justL bool) {
	if !justL {
		return
	}
	for _, r := range a.st.toolTabRects {
		if mx >= r.X && mx <= r.X+r.W && my >= r.Y && my <= r.Y+r.H {
			a.st.toolTab = r.Key
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

	hasSelection := a.selectedStageObject() != nil
	if hasSelection {
		for i := range a.props {
			a.props[i].HandleMouse(mx, my, justL)
		}
		if a.st.opacityInput != nil {
			a.st.opacityInput.HandleMouse(mx, my, justL)
		}
	}

	buttons := []*ui.Button{}
	if hasSelection {
		buttons = append(buttons, a.st.deleteBtn, a.st.duplicateBtn)
	}
	clickedButton := false
	for _, btn := range buttons {
		if btn == nil {
			continue
		}
		btn.HandleMouse(mx, my, justL)
		if justL && btn.HitTest(mx, my) {
			clickedButton = true
		}
	}
	if clickedButton || !justL {
		return
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

func (a *EditorApp) handleStageCanvasMouse(mx, my int, leftDown, justPressed bool) {
	if a.st.dragAsset >= 0 {
		return
	}

	middleDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)
	rightDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	spaceDown := ebiten.IsKeyPressed(ebiten.KeySpace)
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

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
			a.scaleHandle = -1
			a.st.draggingObj = false
			a.st.dragUndoSaved = false
			a.panCanvasByMouse(mx, my)
		}
		return
	}

	if !leftDown {
		a.scaleHandle = -1
		a.rotationDrag.active = false
		a.st.draggingObj = false
		a.st.dragUndoSaved = false
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
			a.selectStageObject(idx)
			if obj := a.selectedStageObject(); obj != nil {
				a.st.dragObjStartX = obj.X
				a.st.dragObjStartY = obj.Y
				a.st.dragWorldStartX = wx
				a.st.dragWorldStartY = wy
				a.st.draggingObj = true
			}
			return
		}
		a.selectStageObject(-1)
		a.st.panning = true
		return
	}

	if a.scaleHandle >= 0 && ctrl {
		a.applyHandleScale(mx, my)
		return
	}

	if a.st.draggingObj {
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
		if newX == obj.X && newY == obj.Y {
			return
		}
		if !a.st.dragUndoSaved {
			a.saveSnapshot()
			a.st.dragUndoSaved = true
		}
		obj.X = newX
		obj.Y = newY
		a.syncStageInputs()
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
	case ctrl && obj != nil:
		factor := 1 + wy*0.05
		if factor < 0.01 {
			factor = 0.01
		}
		obj.Width = math.Max(1, math.Round(obj.Width*factor*100)/100)
		obj.Height = math.Max(1, math.Round(obj.Height*factor*100)/100)
		a.props[2].SetNumeric(obj.Width)
		a.props[3].SetNumeric(obj.Height)
	default:
		a.canvas.Cam.ZoomAt(1+wy*0.1, float64(mx-a.canvas.X), float64(my-a.canvas.Y))
	}
}

func (a *EditorApp) handleStageKeys() {
	if a.focusedInput() != nil {
		return
	}
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyD) {
		a.duplicateSelectedStageObject()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		a.deleteSelectedStageObject()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.selectStageObject(-1)
	}
}
