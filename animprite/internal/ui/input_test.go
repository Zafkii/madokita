package ui

import (
	"testing"

	"animprite/internal/theme"
)

func TestTextInputDoubleClickSelectsWord(t *testing.T) {
	inp := NewTextInput(0, 0, 80, 20, theme.NewManager())
	inp.Text = "256"

	inp.HandleMouse(10, 10, true)
	if !inp.Focused {
		t.Fatal("first click did not focus the input")
	}
	inp.HandleMouse(11, 10, true)
	if got := inp.selectedText(); got != "256" {
		t.Fatalf("double-click selected %q, want %q", got, "256")
	}
}

func TestTextInputTripleClickSelectsAll(t *testing.T) {
	inp := NewTextInput(0, 0, 80, 20, theme.NewManager())
	inp.Text = "hello world"

	inp.HandleMouse(10, 10, true)
	inp.HandleMouse(11, 10, true)
	inp.HandleMouse(12, 10, true)
	if got := inp.selectedText(); got != "hello world" {
		t.Fatalf("triple-click selected %q, want %q", got, "hello world")
	}
}

func TestTextInputPasteTextNumericFilters(t *testing.T) {
	inp := NewTextInput(0, 0, 80, 20, theme.NewManager())
	inp.Numeric = true

	if got := inp.pasteText("  -12.5abc "); got != "-12.5" {
		t.Fatalf("numeric paste produced %q, want %q", got, "-12.5")
	}
	if got := inp.pasteText("1.2.3"); got != "1.23" {
		t.Fatalf("numeric paste produced %q, want %q", got, "1.23")
	}
	if got := inp.pasteText("- 5"); got != "-5" {
		t.Fatalf("numeric paste produced %q, want %q", got, "-5")
	}
}

func TestTextInputPasteTextKeepsPlainText(t *testing.T) {
	inp := NewTextInput(0, 0, 80, 20, theme.NewManager())

	if got := inp.pasteText("hello world"); got != "hello world" {
		t.Fatalf("plain paste produced %q, want %q", got, "hello world")
	}
}
