package ui

import "testing"

func TestTooltipPlacementStaysInsideScreen(t *testing.T) {
	tests := []struct {
		name                 string
		mx, my, w, h, sw, sh int
	}{
		{"normal", 10, 10, 80, 20, 400, 300},
		{"flips_left", 380, 10, 80, 20, 400, 300},
		{"flips_up", 10, 290, 80, 20, 400, 300},
		{"clamps_origin", -100, -100, 80, 20, 400, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := tooltipPlacement(tt.mx, tt.my, tt.w, tt.h, tt.sw, tt.sh)
			if x < tooltipMargin || y < tooltipMargin {
				t.Fatalf("tooltip origin (%d,%d) is outside margin", x, y)
			}
			if x+tt.w > tt.sw-tooltipMargin || y+tt.h > tt.sh-tooltipMargin {
				t.Fatalf("tooltip rect (%d,%d,%d,%d) exceeds screen %dx%d", x, y, tt.w, tt.h, tt.sw, tt.sh)
			}
		})
	}
}
