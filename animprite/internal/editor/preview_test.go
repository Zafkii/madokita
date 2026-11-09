package editor

import (
	"testing"

	"animprite/internal/project"
)

func TestAdvancePreviewMovesToNextFrame(t *testing.T) {
	a := newScaleTestApp()
	a.mode = modeMovement
	a.animTable.SelectedIdx = 0
	a.proj.Animations[0].FPS = 10
	a.proj.Animations[0].Loop = true
	a.proj.Animations[0].CurrentIdx = 0
	a.proj.Animations[0].Frames = []project.AnimationFrame{
		{Sprites: a.defaultFrameSprites()},
		{Sprites: a.defaultFrameSprites()},
		{Sprites: a.defaultFrameSprites()},
		{Sprites: a.defaultFrameSprites()},
	}
	a.prev.previewPlaying = true
	a.prev.loopChecked = true
	a.prev.previewSpeed = 1
	a.prev.previewAnimIdx = a.animTable.SelectedIdx
	a.prev.previewAccumulator = 0

	for i := 0; i < 12; i++ {
		a.advancePreview()
	}

	if got := a.proj.Animations[0].CurrentIdx; got != 1 {
		t.Fatalf("CurrentIdx = %d, want 1 after 12 ticks at 10 FPS", got)
	}
}
