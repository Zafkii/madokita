package combat

import (
	"testing"

	math2 "madokita-gems/internal/math"
)

// The editor authors the geometry per attack frame; the hitbox keeps the
// damage values from the config and resolves every authored box.
func TestHitboxRectsUseTheAuthoredGeometry(t *testing.T) {
	h := NewHitbox(HitboxConfig{W: 10, H: 10, Damage: 5}, "p1")

	if got := h.Rects(math2.Vec2{X: 5, Y: 5}, false); len(got) != 1 || got[0] != math2.NewRect(0, 0, 10, 10) {
		t.Fatalf("config fallback = %v, want one 10x10 box at the owner", got)
	}

	h.SetGeometry([]HitboxGeometry{
		{W: 40, H: 30, OffsetX: 60, OffsetY: -20},
		{W: 8, H: 8},
	})
	got := h.Rects(math2.Vec2{X: 100, Y: 200}, false)
	want := []math2.Rect{
		math2.NewRect(140, 165, 40, 30),
		math2.NewRect(96, 196, 8, 8),
	}
	if len(got) != len(want) {
		t.Fatalf("rects = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("rect %d = %v, want %v", i, got[i], want[i])
		}
	}

	flipped := h.Rects(math2.Vec2{X: 100, Y: 200}, true)
	if flipped[0] != math2.NewRect(20, 165, 40, 30) {
		t.Fatalf("flipped first box = %v, want the offset mirrored", flipped[0])
	}

	h.SetGeometry(nil)
	if got := h.Rects(math2.Vec2{X: 5, Y: 5}, false); len(got) != 1 || got[0] != math2.NewRect(0, 0, 10, 10) {
		t.Fatalf("cleared geometry = %v, want the config fallback again", got)
	}
}
