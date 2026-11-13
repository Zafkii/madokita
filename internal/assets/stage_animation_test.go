package assets

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"madokita-gems/internal/stage"
	"madokita-gems/internal/ui"
)

func TestPreloadStageComposesAnimatedObject(t *testing.T) {
	dir := t.TempDir()
	writeStageSheet(t, filepath.Join(dir, "sprite_prop_tree.png"))

	layout := &stage.Stage{
		Width:  800,
		Height: 400,
		Layers: []stage.Layer{{Name: "Layer 1"}},
		Assets: []stage.Asset{{
			Name:     "tree",
			File:     "sprite_prop_tree.png",
			Category: "anim_prop",
			Width:    400,
			Height:   400,
			Sprites: []stage.SpriteSheet{{
				File:       "sprite_prop_tree.png",
				FrameW:     400,
				FrameH:     400,
				FrameCount: 2,
			}},
			Animation: &stage.Animation{
				Name:   "idle",
				FPS:    1,
				Loop:   true,
				FrameW: 400,
				FrameH: 400,
				Frames: []stage.AnimFrame{
					{Sprites: []stage.AnimSprite{{
						SpriteIdx: 0, SpriteFrameIdx: 0,
						ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5,
					}}},
					{Sprites: []stage.AnimSprite{{
						SpriteIdx: 0, SpriteFrameIdx: 1,
						ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5,
					}}},
				},
			},
		}},
		Objects: []stage.Object{{
			Layer:    "Layer 1",
			Category: "anim_prop",
			File:     "sprite_prop_tree.png",
			X:        200,
			Y:        200,
			Width:    400,
			Height:   400,
			Opacity:  1,
		}},
	}

	mgr := NewAssetManager(ui.NewImageCache(dir))
	if err := mgr.PreloadStage(&StageDef{ID: "test", Layout: layout}); err != nil {
		t.Fatalf("PreloadStage() error: %v", err)
	}

	obj := &layout.Objects[0]
	anim := mgr.StageObjectAnimation("test", obj)
	if anim == nil {
		t.Fatal("animated stage object has no animation")
	}
	if len(anim.Frames) != 2 {
		t.Fatalf("frames = %d, want 2", len(anim.Frames))
	}
	if got := anim.Frames[0].Bounds(); got != image.Rect(0, 0, 400, 400) {
		t.Fatalf("frame bounds = %v, want 400x400", got)
	}
	if anim.Frame(0) != anim.Frames[0] || anim.Frame(1.1) != anim.Frames[1] {
		t.Fatal("frame selection does not follow the animation clock")
	}
	if anim.Frame(2.1) != anim.Frames[0] {
		t.Fatal("looping animation did not wrap")
	}
	anim.Loop = false
	if anim.Frame(5) != anim.Frames[1] {
		t.Fatal("non-looping animation did not clamp to the last frame")
	}
	if static := (&stage.Object{Category: "prop", File: "sprite_prop_tree.png"}); mgr.StageObjectAnimation("test", static) != nil {
		t.Fatal("static object sharing an animated sheet received an animation")
	}

	mgr.UnloadStage("test")
	if mgr.StageObjectAnimation("test", obj) != nil {
		t.Fatal("animation still available after UnloadStage")
	}
}

func TestPreloadStageSkipsStartMarker(t *testing.T) {
	layout := &stage.Stage{
		Width:  800,
		Height: 400,
		Layers: []stage.Layer{{Name: "Layer 1"}},
		Objects: []stage.Object{{
			Category: stage.CategoryStart,
			X:        320,
			Y:        480,
			Width:    32,
			Height:   32,
		}},
	}

	mgr := NewAssetManager(ui.NewImageCache(t.TempDir()))
	if err := mgr.PreloadStage(&StageDef{ID: "test", Layout: layout}); err != nil {
		t.Fatalf("PreloadStage() error: %v", err)
	}
	if len(mgr.pathRefs) != 0 {
		t.Fatalf("start marker loaded image paths: %+v", mgr.pathRefs)
	}
}

func writeStageSheet(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 800, 400))
	for y := 0; y < 400; y++ {
		for x := 0; x < 800; x++ {
			c := color.RGBA{R: 255, A: 255}
			if x >= 400 {
				c = color.RGBA{B: 255, A: 255}
			}
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create sheet: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode sheet: %v", err)
	}
}
