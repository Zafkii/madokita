package editor

type editorMode int

const (
	modeMovement editorMode = iota
	modeAttack
)

type rightPanelMode int

const (
	panelModeSprite rightPanelMode = iota
	panelModeHurtbox
	panelModeAnimFrame
)

type titleBarBtn int

const (
	btnNone titleBarBtn = iota
	btnMinimize
	btnMaximize
	btnClose
)

type resizeEdge int

const (
	edgeNone resizeEdge = iota
	edgeLeft
	edgeRight
	edgeTop
	edgeBottom
	edgeTopLeft
	edgeTopRight
	edgeBottomLeft
	edgeBottomRight
)

type resizeInfo struct {
	active  bool
	edge    resizeEdge
	startSX int
	startSY int
	initW   int
	initH   int
	initX   int
	initY   int
}

const (
	rsIdle = iota
	rsPending
	rsApply
)

type scaleOrigData struct {
	handleIdx int
	anchorWx  float64
	anchorWy  float64
	origW     float64
	origH     float64
	origRot   float64
	hbSx, hbSy       float64
	hbOx, hbOy       float64
	hbCos, hbSin     float64
	hbOrigX, hbOrigY float64
	pxW, pxH             float64
	originX, originY     float64
	origScaleX, origScaleY   float64
	origOffsetX, origOffsetY float64
}

var oppositeHandles = [8]int{2, 3, 0, 1, 5, 4, 7, 6}
