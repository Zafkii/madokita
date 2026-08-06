package main

import (
	"math"
	"time"

	"madokita/internal/engine"
	"madokita/internal/input"
	"madokita/internal/settings"

	"github.com/hajimehoshi/ebiten/v2"
)

// maxFrameDt clamps the advance time per Update so a stutter (window drag,
// OS stall) never injects a huge dt into physics or menu timers.
const maxFrameDt = 0.1

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// setClientSize sizes the window so the OS client area (the region Ebitengine
// renders to) is exactly width x height, with the native title bar and borders
// added on top. GLFW measures the whole window, so the request must include
// the OS frame deltas once calibrated.
func (g *GameApp) setClientSize(width, height int) {
	if g.frameCalibrated {
		ebiten.SetWindowSize(width+g.frameW, height+g.frameH)
		return
	}
	ebiten.SetWindowSize(width, height)
}

// snapWindowAspect keeps the windowed client area at the game's 16:9 aspect
// ratio. The OS handles the actual resizing via the native window borders;
// whenever the client area drifts from 16:9 during a resize, it is snapped
// back so the fixed-resolution canvas always fills the client and no black
// letterbox bars appear. The tolerance avoids fighting the resize at the
// rounding boundary.
//
// The client size comes from g.lastClient (measured in Layout), not from
// ebiten.WindowSize(): that call returns the window DIP size, whose meaning
// flips between the requested total and the actual framebuffer depending on
// which callback updated it last, so it is not safe to derive either the
// client area or the frame deltas from it.
func (g *GameApp) snapWindowAspect() {
	if !g.frameCalibrated || ebiten.IsFullscreen() || ebiten.IsWindowMaximized() {
		return
	}
	cw, ch := g.lastClientW, g.lastClientH
	if cw <= 0 || ch <= 0 {
		return
	}
	targetH := int(math.Round(float64(cw) * 9 / 16))
	if targetH < 1 {
		targetH = 1
	}
	if absInt(ch-targetH) >= 1 {
		g.setClientSize(cw, targetH)
	}
}

func (g *GameApp) Update() error {
	if ebiten.IsWindowBeingClosed() {
		wx, wy := ebiten.WindowPosition()
		settings.SetWindowPosition(wx, wy)
		return ErrWindowClose
	}

	dt := time.Since(g.lastTick).Seconds()
	if dt <= 0 {
		dt = 1.0 / 60.0
	}
	if dt > maxFrameDt {
		dt = maxFrameDt
	}
	g.lastTick = time.Now()

	if g.runtime.Phase() == engine.PhaseRunning {
		if err := g.sceneMgr.Update(dt); err != nil {
			return err
		}
	}
	g.inputMgr.Update()

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
			g.setClientSize(d.Resolution.Width, d.Resolution.Height)
			ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
			ebiten.SetFullscreen(false)
		}
		ebiten.SetMaxTPS(d.FPSLimit)
	}

	g.snapWindowAspect()

	if g.inputMgr.IsChordJustPressed(input.ActionToggleFullscreen) {
		settings.SetFullscreen(!settings.GetData().Fullscreen)
	}

	return nil
}

func (g *GameApp) Draw(screen *ebiten.Image) {
	g.sceneMgr.Draw(screen)
}

func (g *GameApp) Layout(w, h int) (int, int) {
	gw, gh := settings.GetResolution()
	g.sceneMgr.SetGameSize(gw, gh)
	if !g.frameCalibrated && !ebiten.IsFullscreen() && !ebiten.IsWindowMaximized() {
		// First windowed layout: the client area is what the OS left after
		// taking its title bar and borders, so the frame size is the
		// difference between the requested resolution and what we got.
		g.frameW = gw - w
		g.frameH = gh - h
		g.frameCalibrated = true
	}
	g.lastClientW, g.lastClientH = w, h
	return gw, gh
}
