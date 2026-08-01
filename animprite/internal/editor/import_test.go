package editor

import (
	"testing"

	"animprite/internal/project"
)

func TestMergeProjectSprites(t *testing.T) {
	imported := []project.SpriteRow{{Name: "Sprite 0", File: "a.png"}, {Name: "Sprite 1", File: "b.png"}}
	session := []project.SpriteRow{{Name: "Sprite 0", File: "old.png"}}

	got := mergeProjectSprites(imported, session)
	if len(got) != 2 || got[1].File != "b.png" {
		t.Fatalf("file sprites should win: got %+v", got)
	}
	if &got[0] != &imported[0] {
		t.Errorf("expected the imported slice itself, got a different backing array")
	}

	legacy := mergeProjectSprites(nil, session)
	if len(legacy) != 1 || legacy[0].File != "old.png" {
		t.Fatalf("session sprites should survive legacy files: got %+v", legacy)
	}

	empty := mergeProjectSprites(nil, nil)
	if len(empty) != 0 {
		t.Fatalf("expected empty result, got %+v", empty)
	}
}
