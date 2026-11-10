package ui

import "testing"

func TestButtonWidthKeepsLabelPadding(t *testing.T) {
	for _, label := range []string{"+", "-", "Copy Prev"} {
		if got := ButtonWidth(label); got < TextWidth(label)+8 {
			t.Fatalf("ButtonWidth(%q) = %d, want at least %d", label, got, TextWidth(label)+8)
		}
	}
}
