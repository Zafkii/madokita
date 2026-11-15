package menu

import (
	"errors"
	"testing"

	"madokita-gems/internal/input"
	"madokita-gems/internal/scene"
	"madokita-gems/internal/ui"
)

func newTestMainMenu(t *testing.T) *MainMenuScene {
	t.Helper()
	cache := ui.NewImageCache("testdata/no-assets")
	mgr := scene.NewManager(1280, 720)
	s := NewMainMenuScene(mgr, cache, input.NewManager(), nil)
	if err := s.Enter(); err != nil {
		t.Fatalf("enter: %v", err)
	}
	return s
}

// The EXIT option is the last entry of the main menu list.
func TestMainMenuExitOptionRequestsQuit(t *testing.T) {
	s := newTestMainMenu(t)
	if len(s.options) != 4 {
		t.Fatalf("options = %d, want 4", len(s.options))
	}

	if err := s.Update(0.016); err != nil {
		t.Fatalf("idle update: %v", err)
	}

	s.options[3].Click()

	if err := s.Update(0.016); !errors.Is(err, ErrExit) {
		t.Fatalf("after EXIT clicked, update = %v, want ErrExit", err)
	}
	if err := s.Update(0.016); !errors.Is(err, ErrExit) {
		t.Fatalf("update after quit request = %v, want ErrExit", err)
	}
}

func TestMainMenuEnterClearsQuitRequest(t *testing.T) {
	s := newTestMainMenu(t)
	s.quitRequested = true
	if err := s.Update(0.016); !errors.Is(err, ErrExit) {
		t.Fatalf("update = %v, want ErrExit", err)
	}

	if err := s.Enter(); err != nil {
		t.Fatalf("re-enter: %v", err)
	}
	if s.quitRequested {
		t.Fatal("re-entering the menu must clear a previous quit request")
	}
	if err := s.Update(0.016); err != nil {
		t.Fatalf("update after re-enter: %v", err)
	}
}
