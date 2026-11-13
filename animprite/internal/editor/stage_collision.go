package editor

import (
	"fmt"

	"animprite/internal/project"
	"animprite/internal/theme"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

const stageCollisionTopPanelH = 190

func (a *EditorApp) initStageCollisionWidgets() {
	th := a.th

	a.st.collisionViewBtn = ui.NewButton(0, 0, 1, 20, "Show collisions", th)
	a.st.collisionViewBtn.BtnType = ui.BtnDefault
	a.st.collisionViewBtn.OnClick = func() {
		a.st.showCollisions = !a.st.showCollisions
		if a.st.showCollisions {
			a.setStatus("Showing collision and hit boxes")
		} else {
			a.setStatus("Collision boxes hidden")
		}
	}

	a.st.collisionEditBtn = ui.NewButton(0, 0, 1, 22, "Edit collision", th)
	a.st.collisionEditBtn.BtnType = ui.BtnBlue
	a.st.collisionEditBtn.OnClick = func() {
		a.openStageCollisionEditor()
	}

	a.st.collisionToggleBtn = ui.NewButton(0, 0, 1, 22, "Collision: On", th)
	a.st.collisionToggleBtn.BtnType = ui.BtnDefault
	a.st.collisionToggleBtn.OnClick = func() {
		a.toggleStageCollisionForSelection()
	}

	a.st.collisionBackBtn = ui.NewButton(0, 0, 1, 20, "Back", th)
	a.st.collisionBackBtn.BtnType = ui.BtnDefault
	a.st.collisionBackBtn.OnClick = func() {
		a.closeStageCollisionEditor()
	}

	a.st.collisionTabBtns = []*ui.Button{
		ui.NewButton(0, 0, 1, 20, "Collision", th),
		ui.NewButton(0, 0, 1, 20, "Hitbox", th),
	}
	for i, btn := range a.st.collisionTabBtns {
		idx := i
		btn.OnClick = func() {
			a.setStageCollisionTab(idx)
		}
	}

	a.st.collisionTable = ui.NewTable("Collision boxes", []ui.TableColumn{
		{Label: "#", Width: 26},
		{Label: "X"},
		{Label: "Y"},
		{Label: "W"},
		{Label: "H"},
		{Label: "Rot"},
	}, 4, th)

	addBtn := ui.NewButton(0, 0, ui.ButtonWidth("+"), topPanelBtnH, "+", th)
	addBtn.BtnType = ui.BtnBlue
	addBtn.Tooltip = "Add box"
	addBtn.OnClick = func() {
		a.addStageCollisionBox()
	}
	delBtn := ui.NewButton(0, 0, ui.ButtonWidth("-"), topPanelBtnH, "-", th)
	delBtn.BtnType = ui.BtnRed
	delBtn.Tooltip = "Remove selected box"
	delBtn.OnClick = func() {
		a.removeStageCollisionBox()
	}
	a.st.collisionTable.AddBtn = addBtn
	a.st.collisionTable.RemoveBtn = delBtn
	a.st.collisionTable.DrawRow = func(screen *ebiten.Image, p theme.Palette, tx, rowY, cw int, xs []int, idx int) {
		ui.DrawText(screen, fmt.Sprintf("%d", idx+1), tx+xs[0], rowY+3, 1, p.TextPrimary)
		inputs := []*ui.TextInput{
			a.st.collisionXInputs[idx],
			a.st.collisionYInputs[idx],
			a.st.collisionWInputs[idx],
			a.st.collisionHInputs[idx],
			a.st.collisionRotInputs[idx],
		}
		for col, inp := range inputs {
			x := tx + xs[col+1]
			w := (tx + cw - tableColPad) - x - 2
			if col < len(inputs)-1 {
				w = xs[col+2] - xs[col+1] - 2
			}
			if w < 20 {
				w = 20
			}
			inp.X = x
			inp.Y = rowY + (topTableRowH-topPanelBtnH)/2
			inp.W = w
			inp.H = topPanelBtnH
			inp.Visible = true
			inp.Draw(screen)
		}
	}
}

func (a *EditorApp) stageCollisionList() *[]project.StageBox {
	obj := a.selectedStageObject()
	if obj == nil {
		return nil
	}
	if !project.StageCategoryHasCollision(obj.Category) {
		return nil
	}
	if a.st.collisionTab == "hitbox" {
		return &obj.Hitboxes
	}
	return &obj.CollisionBoxes
}

func (a *EditorApp) stageCollisionInputList() []*ui.TextInput {
	all := make([]*ui.TextInput, 0, len(a.st.collisionXInputs)*5)
	all = append(all, a.st.collisionXInputs...)
	all = append(all, a.st.collisionYInputs...)
	all = append(all, a.st.collisionWInputs...)
	all = append(all, a.st.collisionHInputs...)
	all = append(all, a.st.collisionRotInputs...)
	return all
}

func (a *EditorApp) syncStageCollisionInputs() {
	if a.st.collisionTable == nil {
		return
	}
	listPtr := a.stageCollisionList()
	var list []project.StageBox
	if listPtr != nil {
		list = *listPtr
	}
	n := len(list)
	a.st.collisionTable.RowCount = n

	newInput := func(min, step float64) *ui.TextInput {
		inp := ui.NewTextInput(0, 0, 50, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = min
		inp.Step = step
		return inp
	}
	for len(a.st.collisionXInputs) < n {
		a.st.collisionXInputs = append(a.st.collisionXInputs, newInput(-99999, 1))
	}
	for len(a.st.collisionYInputs) < n {
		a.st.collisionYInputs = append(a.st.collisionYInputs, newInput(-99999, 1))
	}
	for len(a.st.collisionWInputs) < n {
		a.st.collisionWInputs = append(a.st.collisionWInputs, newInput(1, 1))
	}
	for len(a.st.collisionHInputs) < n {
		a.st.collisionHInputs = append(a.st.collisionHInputs, newInput(1, 1))
	}
	for len(a.st.collisionRotInputs) < n {
		a.st.collisionRotInputs = append(a.st.collisionRotInputs, newInput(-360, 0.5))
	}
	a.st.collisionXInputs = a.st.collisionXInputs[:n]
	a.st.collisionYInputs = a.st.collisionYInputs[:n]
	a.st.collisionWInputs = a.st.collisionWInputs[:n]
	a.st.collisionHInputs = a.st.collisionHInputs[:n]
	a.st.collisionRotInputs = a.st.collisionRotInputs[:n]

	for idx := 0; idx < n; idx++ {
		i := idx
		a.st.collisionXInputs[i].SetNumeric(list[i].X)
		a.st.collisionXInputs[i].OnChange = func(_ string) {
			if list := a.stageCollisionList(); list != nil && i < len(*list) {
				(*list)[i].X = a.st.collisionXInputs[i].NumericValue()
				a.dirty = true
			}
		}
		a.st.collisionYInputs[i].SetNumeric(list[i].Y)
		a.st.collisionYInputs[i].OnChange = func(_ string) {
			if list := a.stageCollisionList(); list != nil && i < len(*list) {
				(*list)[i].Y = a.st.collisionYInputs[i].NumericValue()
				a.dirty = true
			}
		}
		a.st.collisionWInputs[i].SetNumeric(list[i].Width)
		a.st.collisionWInputs[i].OnChange = func(_ string) {
			if list := a.stageCollisionList(); list != nil && i < len(*list) {
				(*list)[i].Width = a.st.collisionWInputs[i].NumericValue()
				a.dirty = true
			}
		}
		a.st.collisionHInputs[i].SetNumeric(list[i].Height)
		a.st.collisionHInputs[i].OnChange = func(_ string) {
			if list := a.stageCollisionList(); list != nil && i < len(*list) {
				(*list)[i].Height = a.st.collisionHInputs[i].NumericValue()
				a.dirty = true
			}
		}
		a.st.collisionRotInputs[i].SetNumeric(list[i].Rotation)
		a.st.collisionRotInputs[i].OnChange = func(_ string) {
			if list := a.stageCollisionList(); list != nil && i < len(*list) {
				(*list)[i].Rotation = a.st.collisionRotInputs[i].NumericValue()
				a.dirty = true
			}
		}
	}

	if n == 0 {
		a.st.collisionTable.SelectedIdx = -1
	} else if a.st.collisionTable.SelectedIdx >= n {
		a.st.collisionTable.SelectedIdx = n - 1
	}
}

func (a *EditorApp) ensureStageCollisionDefault() {
	obj := a.selectedStageObject()
	if obj == nil || !project.StageCategoryIsBlock(obj.Category) {
		return
	}
	if len(obj.CollisionBoxes) > 0 {
		return
	}
	a.saveSnapshotNoDirty()
	obj.CollisionBoxes = []project.StageBox{obj.DefaultCollisionBox()}
	a.dirty = true
}

func (a *EditorApp) openStageCollisionEditor() {
	obj := a.selectedStageObject()
	if obj == nil || !project.StageCategoryHasCollision(obj.Category) {
		a.setStatus("Select a block or prop to edit its collision")
		return
	}
	a.clearStageInputFocus()
	a.st.collisionEdit = true
	a.st.collisionTab = "collision"
	a.st.showCollisions = true
	a.ensureStageCollisionDefault()
	a.st.collisionTable.SelectedIdx = -1
	a.st.collisionTable.Scroll = 0
	a.syncStageCollisionInputs()
	a.syncLayout()
	a.setStatus("Editing collision boxes")
}

func (a *EditorApp) closeStageCollisionEditor() {
	a.clearStageInputFocus()
	a.st.collisionEdit = false
	if a.st.collisionTable != nil {
		a.st.collisionTable.Visible = false
	}
	a.syncLayout()
}

func (a *EditorApp) setStageCollisionTab(tab int) {
	if tab < 0 || tab > 1 {
		return
	}
	a.clearStageInputFocus()
	if tab == 1 {
		a.st.collisionTab = "hitbox"
	} else {
		a.st.collisionTab = "collision"
		a.ensureStageCollisionDefault()
	}
	a.st.collisionTable.SelectedIdx = -1
	a.st.collisionTable.Scroll = 0
	a.syncStageCollisionInputs()
}

func (a *EditorApp) addStageCollisionBox() {
	listPtr := a.stageCollisionList()
	if listPtr == nil {
		return
	}
	obj := a.selectedStageObject()
	a.saveSnapshot()
	box := project.StageBox{Width: project.DefaultBlockSize, Height: project.DefaultBlockSize}
	if obj != nil {
		box = obj.DefaultCollisionBox()
	}
	*listPtr = append(*listPtr, box)
	if obj != nil {
		obj.CollisionDisabled = false
	}
	a.st.collisionTable.SelectedIdx = len(*listPtr) - 1
	a.st.collisionTable.Scroll = max(0, a.st.collisionTable.SelectedIdx-a.st.collisionTable.MaxVisible+1)
	a.dirty = true
	a.syncStageCollisionInputs()
	a.setStatus("Added box")
}

func (a *EditorApp) removeStageCollisionBox() {
	listPtr := a.stageCollisionList()
	if listPtr == nil {
		return
	}
	idx := a.st.collisionTable.SelectedIdx
	if idx < 0 || idx >= len(*listPtr) {
		a.setStatus("Select a box to remove")
		return
	}
	a.saveSnapshot()
	*listPtr = append((*listPtr)[:idx], (*listPtr)[idx+1:]...)
	obj := a.selectedStageObject()
	if a.st.collisionTab == "collision" && len(*listPtr) == 0 && obj != nil {
		if project.StageCategoryIsBlock(obj.Category) {
			*listPtr = []project.StageBox{obj.DefaultCollisionBox()}
		}
	}
	if obj != nil && project.StageCategoryIsProp(obj.Category) &&
		len(obj.CollisionBoxes) == 0 && len(obj.Hitboxes) == 0 {
		obj.CollisionDisabled = true
	}
	if idx >= len(*listPtr) {
		idx = len(*listPtr) - 1
	}
	a.st.collisionTable.SelectedIdx = idx
	a.dirty = true
	a.syncStageCollisionInputs()
	a.setStatus("Removed box")
}

func (a *EditorApp) toggleStageCollisionForSelection() {
	indices := a.stageSelectedIndices()
	if len(indices) == 0 {
		return
	}
	anyEnabled := false
	anySupported := false
	for _, idx := range indices {
		obj := &a.proj.Stage.Objects[idx]
		if !project.StageCategoryHasCollision(obj.Category) {
			continue
		}
		anySupported = true
		if !obj.CollisionDisabled {
			anyEnabled = true
		}
	}
	if !anySupported {
		a.setStatus("Only blocks and props have collision")
		return
	}
	a.saveSnapshot()
	for _, idx := range indices {
		obj := &a.proj.Stage.Objects[idx]
		if !project.StageCategoryHasCollision(obj.Category) {
			continue
		}
		if !anyEnabled && project.StageCategoryIsProp(obj.Category) &&
			len(obj.CollisionBoxes) == 0 && len(obj.Hitboxes) == 0 {
			obj.CollisionBoxes = []project.StageBox{obj.DefaultCollisionBox()}
		}
		obj.CollisionDisabled = anyEnabled
	}
	a.dirty = true
	a.syncStageCollisionInputs()
	if anyEnabled {
		a.setStatus("Collision disabled for selection")
	} else {
		a.setStatus("Collision enabled for selection")
	}
}

func (a *EditorApp) drawStageCollisionEditor(screen *ebiten.Image, p theme.Palette, x, top, right int) {
	if a.st.collisionTable == nil {
		return
	}
	width := right - x
	if width <= 0 {
		return
	}

	tabW := 82
	for i, btn := range a.st.collisionTabBtns {
		if btn == nil {
			continue
		}
		active := (i == 0 && a.st.collisionTab == "collision") || (i == 1 && a.st.collisionTab == "hitbox")
		btn.X = x + i*(tabW+4)
		btn.Y = top
		btn.W = tabW
		btn.H = 20
		btn.BtnType = ui.BtnDefault
		if active {
			btn.BtnType = ui.BtnBlue
		}
		btn.Visible = true
		btn.Draw(screen)
	}
	if btn := a.st.collisionBackBtn; btn != nil {
		btn.W = ui.ButtonWidth(btn.Text)
		btn.X = right - btn.W
		btn.Y = top
		btn.H = 20
		btn.Visible = true
		btn.Draw(screen)
	}

	title := "Collision boxes"
	if a.st.collisionTab == "hitbox" {
		title = "Hitboxes"
	}
	a.st.collisionTable.Title = title
	a.st.collisionTable.Visible = true
	a.st.collisionTable.Draw(screen, p, x, top+28, width)
}

func (a *EditorApp) handleStageCollisionMouse(mx, my int, justL bool) bool {
	if !a.st.collisionEdit || a.st.collisionTable == nil {
		return false
	}

	for _, btn := range a.st.collisionTabBtns {
		if btn == nil {
			continue
		}
		if justL && btn.HitTest(mx, my) {
			btn.HandleMouse(mx, my, true)
			return true
		}
		btn.HandleMouse(mx, my, false)
	}
	if btn := a.st.collisionBackBtn; btn != nil {
		if justL && btn.HitTest(mx, my) {
			btn.HandleMouse(mx, my, true)
			return true
		}
		btn.HandleMouse(mx, my, false)
	}

	tbl := a.st.collisionTable
	if justL && tbl.AddBtn != nil && tbl.AddBtn.HitTest(mx, my) {
		tbl.AddBtn.HandleMouse(mx, my, true)
		return true
	}
	if justL && tbl.RemoveBtn != nil && tbl.RemoveBtn.HitTest(mx, my) {
		tbl.RemoveBtn.HandleMouse(mx, my, true)
		return true
	}
	if tbl.AddBtn != nil {
		tbl.AddBtn.HandleMouse(mx, my, false)
	}
	if tbl.RemoveBtn != nil {
		tbl.RemoveBtn.HandleMouse(mx, my, false)
	}

	hitInput := false
	start, end := tbl.VisibleRange()
	for i := start; i < end; i++ {
		if i >= len(a.st.collisionXInputs) {
			break
		}
		inputs := []*ui.TextInput{
			a.st.collisionXInputs[i],
			a.st.collisionYInputs[i],
			a.st.collisionWInputs[i],
			a.st.collisionHInputs[i],
			a.st.collisionRotInputs[i],
		}
		for _, inp := range inputs {
			if stageInfoInputHit(inp, mx, my) {
				hitInput = true
			}
			inp.HandleMouse(mx, my, justL)
		}
	}
	if justL && hitInput {
		return true
	}

	if justL {
		if idx := tbl.HitRow(mx, my); idx >= 0 {
			a.clearStageInputFocus()
			tbl.SelectedIdx = idx
			return true
		}
	}
	return false
}
