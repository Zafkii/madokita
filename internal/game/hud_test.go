package game

import "testing"

func TestStatsHUDLayout(t *testing.T) {
	icon, bars := hudLayout()

	wantBarX := hudMargin + hudIconSize + hudIconGap
	stackH := 3*hudBarH + 2*hudBarGap
	wantTop := hudMargin + (hudIconSize-stackH)/2
	for i, bar := range bars {
		if bar.W != hudBarW || bar.H != hudBarH {
			t.Fatalf("bar %d size = %vx%v, want %vx%v", i, bar.W, bar.H, hudBarW, hudBarH)
		}
		if bar.X != wantBarX {
			t.Fatalf("bar %d left edge = %v, want %v (to the right of the icon)", i, bar.X, wantBarX)
		}
		wantY := wantTop + float64(i)*(hudBarH+hudBarGap)
		if bar.Y != wantY {
			t.Fatalf("bar %d top = %v, want %v (health, soul, stamina order)", i, bar.Y, wantY)
		}
	}

	if icon.W != hudIconSize || icon.H != hudIconSize {
		t.Fatalf("icon size = %vx%v, want %vx%v", icon.W, icon.H, hudIconSize, hudIconSize)
	}
	if icon.X != hudMargin {
		t.Fatalf("icon left edge = %v, want the top left margin %v", icon.X, hudMargin)
	}
	if gap := bars[0].X - (icon.X + icon.W); gap != hudIconGap {
		t.Fatalf("icon gap = %v, want %v", gap, hudIconGap)
	}
	iconMid := icon.Y + icon.H/2
	stackMid := bars[0].Y + (bars[2].Y+bars[2].H-bars[0].Y)/2
	if iconMid != stackMid {
		t.Fatalf("icon centre = %v, want the bar stack centre %v", iconMid, stackMid)
	}
}

func TestStatsHUDFillWidth(t *testing.T) {
	tests := []struct {
		ratio float64
		want  int
	}{
		{-1, 0},
		{0, 0},
		{0.5, 198},
		{1, 396},
		{2, 396},
	}
	for _, tt := range tests {
		if got := hudFillWidth(tt.ratio); got != tt.want {
			t.Fatalf("hudFillWidth(%v) = %v, want %v", tt.ratio, got, tt.want)
		}
	}

	if got := hudRatio(50, 100); got != 0.5 {
		t.Fatalf("hudRatio(50, 100) = %v, want 0.5", got)
	}
	if got := hudRatio(120, 100); got != 1 {
		t.Fatalf("hudRatio(120, 100) = %v, want a clamped 1", got)
	}
	if got := hudRatio(50, 0); got != 0 {
		t.Fatalf("hudRatio(50, 0) = %v, want 0", got)
	}
}
