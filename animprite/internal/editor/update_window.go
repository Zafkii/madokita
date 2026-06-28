package editor

import (
	"time"

	"animprite/internal/windrag"

	"github.com/hajimehoshi/ebiten/v2"
)

func (a *EditorApp) handleWindowChrome(mx, my int, leftDown, justPressed, isMaxed bool) (skipMouse bool, err error) {
	if a.win.resizing.active {
		if !leftDown {
			a.win.resizing.active = false
			a.saveCurrentPrefs()
		} else {
			sx, sy := windrag.ScreenCursorPos()
			dx := sx - a.win.resizing.startSX
			dy := sy - a.win.resizing.startSY
			nw, nh, nx, ny := a.computeResize(a.win.resizing.edge, dx, dy)
			ebiten.SetWindowSize(nw, nh)
			ebiten.SetWindowPosition(nx, ny)
		}
		return true, nil
	}

	if !ebiten.IsFullscreen() && justPressed && my >= 0 && my < a.win.barLogicH && mx >= a.win.outsideWidth-a.win.btnLogicW*3 {
		switch a.titleBarButtonAt(mx) {
		case btnClose:
			a.saveCurrentPrefs()
			return true, ErrWindowClose
		case btnMaximize:
			if isMaxed {
				a.win.restoreState = rsPending
				a.win.restoreW = a.win.prevW
				a.win.restoreH = a.win.prevH
				ebiten.RestoreWindow()
			} else {
				a.win.prevW, a.win.prevH = ebiten.WindowSize()
				ebiten.MaximizeWindow()
			}
		case btnMinimize:
			ebiten.MinimizeWindow()
		}
		return true, nil
	}

	if !ebiten.IsFullscreen() && !isMaxed && justPressed {
		if ccx, ccy, cok := a.clientCursorPos(); cok {
			if e := a.detectEdge(ccx, ccy); e != edgeNone {
				sx, sy := windrag.ScreenCursorPos()
				wx, wy := ebiten.WindowPosition()
				a.win.resizing = resizeInfo{
					active:  true,
					edge:    e,
					startSX: sx,
					startSY: sy,
					initW:   a.win.outsideWidth,
					initH:   a.win.outsideHeight,
					initX:   wx,
					initY:   wy,
				}
				return true, nil
			}
		}
	}

	if a.win.dragMgr.IsDragging() {
		if !leftDown {
			a.win.dragMgr.StopDrag()
			a.saveCurrentPrefs()
		} else {
			if x, y, ok := a.win.dragMgr.Update(); ok && !ebiten.IsFullscreen() {
				ebiten.SetWindowPosition(x, y)
			}
		}
	} else if a.win.dragPending {
		if leftDown {
			a.win.dragPending = false
			wx, wy := ebiten.WindowPosition()
			a.win.dragMgr.StartDrag(wx, wy)
		} else {
			a.win.dragPending = false
		}
	} else if !ebiten.IsFullscreen() && justPressed && my >= 0 && my < a.win.barLogicH {
		now := time.Now()
		isDouble := !a.win.clickTimer.IsZero() && now.Sub(a.win.clickTimer) < 500*time.Millisecond &&
			absInt(mx-a.win.lastClickMX) < 8 && absInt(my-a.win.lastClickMY) < 8
		a.win.clickTimer = now
		a.win.lastClickMX = mx
		a.win.lastClickMY = my

		if isDouble {
			if isMaxed {
				a.win.restoreState = rsPending
				a.win.restoreW = a.win.prevW
				a.win.restoreH = a.win.prevH
				ebiten.RestoreWindow()
			} else {
				a.win.prevW, a.win.prevH = ebiten.WindowSize()
				ebiten.MaximizeWindow()
			}
		} else if isMaxed {
			a.win.restoreState = rsPending
			a.win.restoreW = a.win.prevW
			a.win.restoreH = a.win.prevH
			ebiten.RestoreWindow()
			a.win.dragPending = true
		} else {
			wx, wy := ebiten.WindowPosition()
			a.win.dragMgr.StartDrag(wx, wy)
		}
	}

	return false, nil
}

func (a *EditorApp) updateWindowCursor(isMaxed bool) {
	if !ebiten.IsFullscreen() && !isMaxed {
		if a.win.resizing.active {
			ebiten.SetCursorShape(cursorForEdge(a.win.resizing.edge))
		} else if ccx, ccy, cok := a.clientCursorPos(); cok {
			if e := a.detectEdge(ccx, ccy); e != edgeNone {
				ebiten.SetCursorShape(cursorForEdge(e))
			} else {
				ebiten.SetCursorShape(ebiten.CursorShapeDefault)
			}
		}
	} else {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}
}

func (a *EditorApp) handleWindowRestoreState() {
	switch a.win.restoreState {
	case rsPending:
		if !ebiten.IsWindowMaximized() {
			a.win.restoreState = rsApply
		}
	case rsApply:
		ebiten.SetWindowSize(a.win.restoreW, a.win.restoreH)
		a.win.restoreState = rsIdle
	}
}
