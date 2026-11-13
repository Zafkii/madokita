package project

import "testing"

func TestApplySizeFromFirstBackground(t *testing.T) {
	stage := StageData{
		Width:  DefaultStageWidth,
		Height: DefaultStageHeight,
		Assets: []StageAsset{
			{Category: StageCategoryProp, Name: "tree", Width: 640, Height: 480},
			{Category: StageCategoryBG, Name: "no-size"},
			{Category: StageCategoryBG, Name: "sky", Width: 3200, Height: 900},
			{Category: StageCategoryBG, Name: "later", Width: 4000, Height: 1000},
		},
	}

	if !stage.ApplySizeFromFirstBackground() {
		t.Fatal("ApplySizeFromFirstBackground() = false, want true")
	}
	if stage.Width != 3200 || stage.Height != 900 {
		t.Fatalf("stage size = %vx%v, want 3200x900", stage.Width, stage.Height)
	}
}

func TestApplySizeFromFirstBackgroundKeepsDefaultsWhenNoValidBackground(t *testing.T) {
	stage := StageData{
		Width:  2560,
		Height: 720,
		Assets: []StageAsset{
			{Category: StageCategoryProp, Name: "tree", Width: 640, Height: 480},
			{Category: StageCategoryBG, Name: "no-size"},
		},
	}

	if stage.ApplySizeFromFirstBackground() {
		t.Fatal("ApplySizeFromFirstBackground() = true, want false")
	}
	if stage.Width != 2560 || stage.Height != 720 {
		t.Fatalf("stage size = %vx%v, want unchanged 2560x720", stage.Width, stage.Height)
	}
}

func TestApplySizeFromFirstBackgroundInfersCategoryFromFile(t *testing.T) {
	stage := StageData{
		Width:  DefaultStageWidth,
		Height: DefaultStageHeight,
		Assets: []StageAsset{
			{Name: "tree", File: "prop_tree.png", Width: 640, Height: 480},
			{Name: "sky", File: "background_sky.png", Width: 3000, Height: 800},
		},
	}

	if !stage.ApplySizeFromFirstBackground() {
		t.Fatal("ApplySizeFromFirstBackground() = false, want true")
	}
	if stage.Width != 3000 || stage.Height != 800 {
		t.Fatalf("stage size = %vx%v, want 3000x800", stage.Width, stage.Height)
	}
}
