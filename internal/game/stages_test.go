package game

import (
	"testing"

	"madokita-gems/internal/stage"
)

func TestRuntimeStageAssetPath(t *testing.T) {
	cases := map[string]string{
		"../../../assets/images/stages/testStage/sky.png": "images/stages/testStage/sky.png",
		"assets/images/stages/testStage/floor.png":        "images/stages/testStage/floor.png",
		"/home/user/game/assets/images/stages/sky.png":    "images/stages/sky.png",
		"images/stages/sky.png":                           "images/stages/sky.png",
		"./images/../images/stages/sky.png":               "images/stages/sky.png",
		"":                                                "",
	}
	for in, want := range cases {
		if got := runtimeStageAssetPath(in); got != want {
			t.Errorf("runtimeStageAssetPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNewStageDefNormalizesLayout(t *testing.T) {
	layout := &stage.Stage{
		Layers: []stage.Layer{
			{Name: "Background"},
			{Name: "Background"},
			{Name: ""},
		},
		Assets: []stage.Asset{{
			Name: "sky",
			File: "../../../assets/images/stages/testStage/sky.png",
		}},
		Objects: []stage.Object{
			{Layer: "Missing", File: "../../../assets/images/stages/testStage/sky.png"},
			{Layer: "Background", File: "/home/user/game/assets/images/stages/floor.png"},
		},
	}

	def := NewStageDef("stage-id", "Stage Name", layout, 700, 50)

	if def.ID != "stage-id" || def.Name != "Stage Name" || def.GroundY != 700 || def.SpawnX != 50 {
		t.Fatalf("stage definition = %+v", def)
	}
	if def.Layout != layout {
		t.Fatal("layout was not attached")
	}
	if layout.Width != StageWidth || layout.Height != StageHeight {
		t.Fatalf("layout size = %vx%v, want %vx%v", layout.Width, layout.Height, StageWidth, StageHeight)
	}
	if len(layout.Layers) != 1 || layout.Layers[0].Name != "Background" {
		t.Fatalf("layers = %+v, want only Background", layout.Layers)
	}
	if layout.Objects[0].Layer != "Background" {
		t.Fatalf("object layer = %q, want Background", layout.Objects[0].Layer)
	}
	if layout.Assets[0].File != "images/stages/testStage/sky.png" {
		t.Fatalf("asset path = %q", layout.Assets[0].File)
	}
	if layout.Objects[1].File != "images/stages/floor.png" {
		t.Fatalf("object path = %q", layout.Objects[1].File)
	}
}

func TestNewStageDefUsesStartMarkerPosition(t *testing.T) {
	layout := &stage.Stage{
		Layers: []stage.Layer{{Name: "Layer 1"}},
		Objects: []stage.Object{
			{Category: stage.CategoryStart, X: 320, Y: 480},
			{Category: "block", File: "../../../assets/images/stages/testStage/floor.png"},
		},
	}

	def := NewStageDef("stage-id", "Stage Name", layout, 700, 50)

	if def.SpawnX != 320 || def.SpawnY != 480 {
		t.Fatalf("spawn = (%v, %v), want (320, 480)", def.SpawnX, def.SpawnY)
	}
	if def.GroundY != 700 {
		t.Fatalf("GroundY = %v, want 700", def.GroundY)
	}
}

func TestNewStageDefDefaultsSpawnYToGround(t *testing.T) {
	layout := &stage.Stage{Layers: []stage.Layer{{Name: "Layer 1"}}}

	def := NewStageDef("stage-id", "Stage Name", layout, 700, 50)

	if def.SpawnX != 50 || def.SpawnY != 700 {
		t.Fatalf("spawn = (%v, %v), want (50, 700)", def.SpawnX, def.SpawnY)
	}
}

func TestNewStageDefRewritesAnimationSpritePaths(t *testing.T) {
	layout := &stage.Stage{
		Layers: []stage.Layer{{Name: "Layer 1"}},
		Assets: []stage.Asset{{
			Name:     "tree",
			File:     "../../../assets/images/stages/test_stage/tree_moves/sprite_prop_tree.png",
			Category: "anim_prop",
			Sprites: []stage.SpriteSheet{{
				File:       "../../../assets/images/stages/test_stage/tree_moves/sprite_prop_tree.png",
				FrameW:     400,
				FrameH:     400,
				FrameCount: 3,
			}},
			Animation: &stage.Animation{
				Name:   "idle",
				FPS:    1,
				Loop:   true,
				FrameW: 400,
				FrameH: 400,
				Frames: []stage.AnimFrame{{}},
			},
		}},
	}

	NewStageDef("stage-id", "Stage Name", layout, 700, 50)

	want := "images/stages/test_stage/tree_moves/sprite_prop_tree.png"
	if got := layout.Assets[0].Sprites[0].File; got != want {
		t.Fatalf("sprite sheet path = %q, want %q", got, want)
	}
}
