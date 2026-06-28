package windrag

type DragManager struct {
	dragging bool
	offsetX  int
	offsetY  int
}

func (d *DragManager) StartDrag(windowX, windowY int) {
	cx, cy := ScreenCursorPos()
	d.offsetX = cx - windowX
	d.offsetY = cy - windowY
	d.dragging = true
}

func (d *DragManager) Update() (newX, newY int, ok bool) {
	if !d.dragging {
		return 0, 0, false
	}
	cx, cy := ScreenCursorPos()
	return cx - d.offsetX, cy - d.offsetY, true
}

func (d *DragManager) StopDrag() {
	d.dragging = false
}

func (d *DragManager) IsDragging() bool {
	return d.dragging
}
