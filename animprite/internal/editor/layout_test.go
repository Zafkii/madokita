package editor

import "testing"

func TestLayoutClampsToMinimumWindow(t *testing.T) {
	a := NewEditorApp()

	w, h := a.Layout(400, 300)
	if w != MinWinW || h != MinWinH {
		t.Fatalf("Layout(400, 300) = (%d, %d), want (%d, %d)", w, h, MinWinW, MinWinH)
	}
	if a.win.outsideWidth != MinWinW || a.win.outsideHeight != MinWinH {
		t.Fatalf("window = (%d, %d), want (%d, %d)", a.win.outsideWidth, a.win.outsideHeight, MinWinW, MinWinH)
	}

	// ui.Table adds 6px of padding on each side. At the minimum window the
	// top-panel table grid must still leave 264px for the fixed columns.
	const uiTablePad = 6
	const minTableContentW = 264
	halfW := (a.tableGridW() - topTableGap) / 2
	if got := halfW - uiTablePad*2; got < minTableContentW {
		t.Fatalf("table content width at minimum window = %d, want at least %d", got, minTableContentW)
	}
}
