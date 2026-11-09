package editor

import (
	"math"
	"testing"
)

func TestRotationGuideHandleFollowsRotation(t *testing.T) {
	a := newScaleTestApp()
	a.buildSpriteRenders()

	a.canvas.Selection.Rotation = 0
	cx, cy, r, hx, hy, ok := a.canvas.RotationGuide()
	if !ok {
		t.Fatal("expected a rotation guide for the visible selection")
	}
	if math.Abs(hx-cx) > 0.01 || math.Abs(hy-(cy-r)) > 0.01 {
		t.Fatalf("at rotation 0 handle = (%v, %v), want top of circle (%v, %v)", hx, hy, cx, cy-r)
	}

	a.canvas.Selection.Rotation = math.Pi / 2
	_, _, _, hx, hy, _ = a.canvas.RotationGuide()
	if math.Abs(hx-(cx+r)) > 0.01 || math.Abs(hy-cy) > 0.01 {
		t.Fatalf("at rotation 90 handle = (%v, %v), want right of circle (%v, %v)", hx, hy, cx+r, cy)
	}
}

func TestRotationHandleHitTest(t *testing.T) {
	a := newScaleTestApp()
	a.panelMode = panelModeHurtbox
	a.buildSpriteRenders()

	_, _, _, hx, hy, ok := a.canvas.RotationGuide()
	if !ok {
		t.Fatal("expected a rotation guide for the visible selection")
	}
	mx := int(math.Round(hx)) + a.canvas.X
	my := int(math.Round(hy)) + a.canvas.Y
	if !a.canvas.RotationHandleHitTest(mx, my) {
		t.Fatalf("expected rotation handle at (%d, %d) to hit", mx, my)
	}
	if a.canvas.RotationHandleHitTest(mx+50, my+50) {
		t.Fatal("rotation handle hit test is too broad")
	}
}

func TestRotationDragRotatesHurtbox(t *testing.T) {
	a := newScaleTestApp()
	a.panelMode = panelModeHurtbox
	a.buildSpriteRenders()
	a.syncPanelToSelection()

	sel := a.canvas.Selection
	cx := sel.X + sel.W/2
	cy := sel.Y + sel.H/2
	startX, startY := a.canvas.Cam.WorldToCanvas(cx+100, cy)
	endX, endY := a.canvas.Cam.WorldToCanvas(cx, cy+100)

	if !a.beginRotationDrag(int(startX)+a.canvas.X, int(startY)+a.canvas.Y) {
		t.Fatal("expected rotation drag to start on the hurtbox")
	}
	a.applyRotationDrag(int(endX)+a.canvas.X, int(endY)+a.canvas.Y)
	a.flushInputsToData()

	if got := hurtbox(a).Rotation; math.Abs(got-90) > 1 {
		t.Fatalf("hurtbox rotation = %v, want 90", got)
	}
	if got := spriteEntry(a).Rotation; got != 0 {
		t.Fatalf("sprite entry rotation = %v, want 0", got)
	}
}

// A mode change mid-drag (preview advancing) must not redirect the rotation to
// the frame sprite instead of the hurtbox that started the drag.
func TestRotationDragKeepsTargetAcrossModeChange(t *testing.T) {
	a := newScaleTestApp()
	a.panelMode = panelModeAnimFrame
	a.buildSpriteRenders()
	a.syncPanelToSelection()

	sel := a.canvas.Selection
	cx := sel.X + sel.W/2
	cy := sel.Y + sel.H/2
	startX, startY := a.canvas.Cam.WorldToCanvas(cx+100, cy)
	endX, endY := a.canvas.Cam.WorldToCanvas(cx, cy+100)

	if !a.beginRotationDrag(int(startX)+a.canvas.X, int(startY)+a.canvas.Y) {
		t.Fatal("expected rotation drag to start on the hurtbox")
	}
	a.panelMode = panelModeAnimFrame
	a.applyRotationDrag(int(endX)+a.canvas.X, int(endY)+a.canvas.Y)
	a.flushInputsToData()

	if got := hurtbox(a).Rotation; math.Abs(got-90) > 1 {
		t.Fatalf("hurtbox rotation = %v, want 90", got)
	}
	if got := spriteEntry(a).Rotation; got != 0 {
		t.Fatalf("sprite entry rotation = %v, want 0", got)
	}
}
