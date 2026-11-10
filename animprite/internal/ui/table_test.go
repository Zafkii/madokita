package ui

import "testing"

// The editor's minimum window width leaves 264px of content per top-panel
// table. Fixed columns must fit inside that before the flexible columns.
func TestFixedColumnsFitAtMinimumTableWidth(t *testing.T) {
	const minTableW = 264
	tests := []struct {
		name    string
		cols    []TableColumn
		lastW   int
		fileCol int
	}{
		{
			name:  "animation",
			cols:  []TableColumn{{Label: "Name"}, {Label: "Frame", Width: 160}},
			lastW: 160,
		},
		{
			name: "sprite",
			cols: []TableColumn{
				{Label: "Grade", Width: 40},
				{Label: "File"},
				{Label: "W", Width: 34},
				{Label: "H", Width: 34},
				{Label: "Frame", Width: 84},
			},
			lastW:   84,
			fileCol: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{Columns: tt.cols, ColPad: 4}
			xs := tbl.calcColumnPositions(minTableW)
			if end := xs[len(xs)-1] + tt.lastW; end > minTableW-tbl.ColPad {
				t.Fatalf("last fixed column ends at %d, table content ends at %d", end, minTableW-tbl.ColPad)
			}
			if tt.fileCol > 0 {
				i := tt.fileCol
				if w := xs[i+1] - xs[i] - 2; w <= 0 {
					t.Fatalf("flexible file column width = %d, want > 0", w)
				}
			}
		})
	}
}
