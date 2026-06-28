package editor

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (a *EditorApp) syncTitleBarScale() {
	a.win.barLogicH = titleBarH
	a.win.btnLogicW = btnPhysW
}

func (a *EditorApp) clientCursorPos() (cx, cy int, ok bool) {
	cx, cy = ebiten.CursorPosition()
	return cx, cy, true
}

func (a *EditorApp) detectEdge(cx, cy int) resizeEdge {
	ow, oh := a.win.outsideWidth, a.win.outsideHeight
	if cx < 0 || cy < 0 || cx >= ow || cy >= oh {
		return edgeNone
	}
	onTop := cy <= resizeEdgeThick
	onBot := cy >= oh-1-resizeEdgeThick
	onLef := cx <= resizeEdgeThick
	onRig := cx >= ow-1-resizeEdgeThick

	switch {
	case onTop && onLef:
		return edgeTopLeft
	case onTop && onRig:
		return edgeTopRight
	case onBot && onLef:
		return edgeBottomLeft
	case onBot && onRig:
		return edgeBottomRight
	case onTop:
		return edgeTop
	case onBot:
		return edgeBottom
	case onLef:
		return edgeLeft
	case onRig:
		return edgeRight
	}
	return edgeNone
}

func (a *EditorApp) computeResize(e resizeEdge, dx, dy int) (newW, newH, newX, newY int) {
	iW, iH := a.win.resizing.initW, a.win.resizing.initH
	iX, iY := a.win.resizing.initX, a.win.resizing.initY

	switch e {
	case edgeRight:
		return max(minWindowW, iW+dx), iH, iX, iY
	case edgeLeft:
		nw := max(minWindowW, iW-dx)
		return nw, iH, iX + (iW - nw), iY
	case edgeBottom:
		return iW, iH + dy, iX, iY
	case edgeTop:
		nh := iH - dy
		return iW, nh, iX, iY + (iH - nh)
	case edgeTopLeft:
		nw := max(minWindowW, iW-dx)
		nh := iH - dy
		return nw, nh, iX + (iW - nw), iY + (iH - nh)
	case edgeTopRight:
		nw := max(minWindowW, iW+dx)
		nh := iH - dy
		return nw, nh, iX, iY + (iH - nh)
	case edgeBottomLeft:
		nw := max(minWindowW, iW-dx)
		return nw, iH + dy, iX + (iW - nw), iY
	case edgeBottomRight:
		return max(minWindowW, iW+dx), iH + dy, iX, iY
	}
	return iW, iH, iX, iY
}

func cursorForEdge(e resizeEdge) ebiten.CursorShapeType {
	switch e {
	case edgeLeft, edgeRight:
		return ebiten.CursorShapeEWResize
	case edgeTop, edgeBottom:
		return ebiten.CursorShapeNSResize
	case edgeTopLeft, edgeBottomRight:
		return ebiten.CursorShapeNWSEResize
	case edgeTopRight, edgeBottomLeft:
		return ebiten.CursorShapeNESWResize
	}
	return ebiten.CursorShapeDefault
}

func (a *EditorApp) titleBarButtonAt(mx int) titleBarBtn {
	bW := a.win.btnLogicW
	aw := a.win.outsideWidth
	switch {
	case mx >= aw-bW:
		return btnClose
	case mx >= aw-bW*2:
		return btnMaximize
	case mx >= aw-bW*3:
		return btnMinimize
	}
	return btnNone
}
