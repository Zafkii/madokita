package main

import (
	"time"

	"madokita/internal/input"
	"madokita/internal/settings"
	"madokita/internal/windrag"

	"github.com/hajimehoshi/ebiten/v2"
)

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *GameApp) clientCursorPos() (cx, cy int, ok bool) {
	mx, my := ebiten.CursorPosition()
	if g.outsideWidth <= 0 || g.outsideHeight <= 0 || g.gameWidth <= 0 || g.gameHeight <= 0 {
		return 0, 0, false
	}
	cx = mx * g.outsideWidth / g.gameWidth
	cy = my * g.outsideHeight / g.gameHeight
	return cx, cy, true
}

func (g *GameApp) detectEdge(cx, cy int) resizeEdge {
	ow, oh := g.outsideWidth, g.outsideHeight
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

func (g *GameApp) computeResize(e resizeEdge, dx, dy int) (newW, newH, newX, newY int) {
	iW, iH := g.resizing.initW, g.resizing.initH
	iX, iY := g.resizing.initX, g.resizing.initY

	right := func(w int) (int, int) { return w, w * 9 / 16 }
	left := func(w int) (int, int, int) { nw := w; nh := w * 9 / 16; return nw, nh, iX + (iW - nw) }

	switch e {
	case edgeRight:
		newW = max(minWindowW, iW+dx)
		newW, newH = right(newW)
		return newW, newH, iX, iY
	case edgeLeft:
		newW = max(minWindowW, iW-dx)
		newW, newH, newX = left(newW)
		return newW, newH, newX, iY
	case edgeBottom:
		newH = max(minWindowW*9/16, iH+dy)
		newW = max(minWindowW, newH*16/9)
		newH = newW * 9 / 16
		return newW, newH, iX, iY
	case edgeTop:
		newH = max(minWindowW*9/16, iH-dy)
		newW = max(minWindowW, newH*16/9)
		newH = newW * 9 / 16
		return newW, newH, iX, iY + (iH - newH)
	case edgeTopLeft:
		newW = max(minWindowW, iW-dx)
		newW, newH, newX = left(newW)
		return newW, newH, newX, iY + (iH - newH)
	case edgeTopRight:
		newW = max(minWindowW, iW+dx)
		newW, newH = right(newW)
		return newW, newH, iX, iY + (iH - newH)
	case edgeBottomLeft:
		newW = max(minWindowW, iW-dx)
		newW, newH, newX = left(newW)
		return newW, newH, newX, iY
	case edgeBottomRight:
		newW = max(minWindowW, iW+dx)
		newW, newH = right(newW)
		return newW, newH, iX, iY
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

func (g *GameApp) Update() error {
	if g.pendingWindow {
		g.pendingWindow = false
		d := g.pendingData
		if d.Fullscreen {
			ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
			if mon := ebiten.Monitor(); mon != nil {
				mw, mh := mon.Size()
				scale := mon.DeviceScaleFactor()
				ebiten.SetWindowSize(int(float64(mw)*scale), int(float64(mh)*scale))
			}
			ebiten.SetFullscreen(true)
		} else {
			ebiten.SetWindowSize(d.Resolution.Width, d.Resolution.Height)
			ebiten.SetFullscreen(false)
		}
		ebiten.SetMaxTPS(d.FPSLimit)
	}

	g.syncTitleBarScale()

	leftDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	justPressed := leftDown && !g.prevLeftBtn
	isMaxed := ebiten.IsWindowMaximized()

	mx, my := ebiten.CursorPosition()

	if g.resizing.active {
		if !leftDown {
			g.resizing.active = false
			wx, wy := ebiten.WindowPosition()
			settings.SetWindowPosition(wx, wy)
		} else {
			sx, sy := windrag.ScreenCursorPos()
			dx := sx - g.resizing.startSX
			dy := sy - g.resizing.startSY
			nw, nh, nx, ny := g.computeResize(g.resizing.edge, dx, dy)
			ebiten.SetWindowSize(nw, nh)
			ebiten.SetWindowPosition(nx, ny)
		}
		g.prevLeftBtn = leftDown
		goto endUpdate
	}

	if !ebiten.IsFullscreen() && justPressed && my >= 0 && my < g.barLogicH && mx >= g.gameWidth-g.btnLogicW*3 {
		switch g.titleBarButtonAt(mx) {
		case btnClose:
			return ErrWindowClose
		case btnMaximize:
			if isMaxed {
				g.restoreState = rsPending
				g.restoreW = g.prevW
				g.restoreH = g.prevH
				ebiten.RestoreWindow()
			} else {
				g.prevW, g.prevH = ebiten.WindowSize()
				ebiten.MaximizeWindow()
			}
		case btnMinimize:
			ebiten.MinimizeWindow()
		}
		g.prevLeftBtn = leftDown
		goto endUpdate
	}

	if !ebiten.IsFullscreen() && !isMaxed && justPressed {
		if ccx, ccy, cok := g.clientCursorPos(); cok {
			if e := g.detectEdge(ccx, ccy); e != edgeNone {
				sx, sy := windrag.ScreenCursorPos()
				wx, wy := ebiten.WindowPosition()
				g.resizing = resizeInfo{
					active:  true,
					edge:    e,
					startSX: sx,
					startSY: sy,
					initW:   g.outsideWidth,
					initH:   g.outsideHeight,
					initX:   wx,
					initY:   wy,
				}
				g.prevLeftBtn = leftDown
				goto endUpdate
			}
		}
	}

	if g.dragMgr.IsDragging() {
		if !leftDown {
			g.dragMgr.StopDrag()
			wx, wy := ebiten.WindowPosition()
			settings.SetWindowPosition(wx, wy)
		} else {
			if x, y, ok := g.dragMgr.Update(); ok && !ebiten.IsFullscreen() {
				ebiten.SetWindowPosition(x, y)
			}
		}
	} else if g.dragPending {
		if leftDown {
			g.dragPending = false
			wx, wy := ebiten.WindowPosition()
			g.dragMgr.StartDrag(wx, wy)
		} else {
			g.dragPending = false
		}
	} else if !ebiten.IsFullscreen() && justPressed && my >= 0 && my < g.barLogicH {
		now := time.Now()
		isDouble := !g.clickTimer.IsZero() && now.Sub(g.clickTimer) < 500*time.Millisecond &&
			absInt(mx-g.lastClickMX) < 8 && absInt(my-g.lastClickMY) < 8
		g.clickTimer = now
		g.lastClickMX = mx
		g.lastClickMY = my

		if isDouble {
			if isMaxed {
				g.restoreState = rsPending
				g.restoreW = g.prevW
				g.restoreH = g.prevH
				ebiten.RestoreWindow()
			} else {
				g.prevW, g.prevH = ebiten.WindowSize()
				ebiten.MaximizeWindow()
			}
		} else if isMaxed {
			g.restoreState = rsPending
			g.restoreW = g.prevW
			g.restoreH = g.prevH
			ebiten.RestoreWindow()
			g.dragPending = true
		} else {
			wx, wy := ebiten.WindowPosition()
			g.dragMgr.StartDrag(wx, wy)
		}
	}
	g.prevLeftBtn = leftDown

endUpdate:
	if !ebiten.IsFullscreen() && !g.resizing.active && my >= 0 && my < g.barLogicH && mx >= g.gameWidth-g.btnLogicW*3 {
		g.hoveredBtn = g.titleBarButtonAt(mx)
	} else {
		g.hoveredBtn = btnNone
	}

	if !ebiten.IsFullscreen() && !isMaxed {
		if ccx, ccy, cok := g.clientCursorPos(); cok {
			if g.resizing.active {
				ebiten.SetCursorShape(cursorForEdge(g.resizing.edge))
			} else if e := g.detectEdge(ccx, ccy); e != edgeNone {
				ebiten.SetCursorShape(cursorForEdge(e))
			} else {
				ebiten.SetCursorShape(ebiten.CursorShapeDefault)
			}
		}
	} else {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}

	if g.inputMgr.IsChordJustPressed(input.ActionToggleFullscreen) {
		settings.SetFullscreen(!settings.GetData().Fullscreen)
	}

	switch g.restoreState {
	case rsPending:
		if !ebiten.IsWindowMaximized() {
			g.restoreState = rsApply
		}
	case rsApply:
		ebiten.SetWindowSize(g.restoreW, g.restoreH)
		g.restoreState = rsIdle
	}

	if !ebiten.IsFullscreen() && !ebiten.IsWindowMaximized() && g.restoreState == rsIdle {
		w, h := ebiten.WindowSize()
		if targetH := w * 9 / 16; h != targetH {
			ebiten.SetWindowSize(w, targetH)
		}
	}
	return nil
}

func (g *GameApp) Draw(screen *ebiten.Image) {
	g.sceneMgr.Draw(screen)
	if !ebiten.IsFullscreen() {
		g.drawTitleBar(screen)
	}
}

func (g *GameApp) Layout(w, h int) (int, int) {
	g.outsideWidth = w
	g.outsideHeight = h
	if ebiten.IsWindowMaximized() {
		g.gameWidth = w
		g.gameHeight = h
		g.sceneMgr.SetGameSize(w, h)
		return w, h
	}
	gw, gh := settings.GetResolution()
	g.gameWidth = gw
	g.gameHeight = gh
	g.sceneMgr.SetGameSize(gw, gh)
	return gw, gh
}
