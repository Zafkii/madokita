package project

import "testing"

func testStageWithLayers(objects ...StageObject) StageData {
	stage := NewStageData("Order Test")
	stage.Layers = []StageLayer{{Name: "Back"}, {Name: "Front"}}
	stage.Objects = objects
	return stage
}

func TestRenderOrderZBeatsCategoryInsideOneLayer(t *testing.T) {
	stage := testStageWithLayers(
		StageObject{Category: StageCategoryBlock, Z: 5},
		StageObject{Category: StageCategoryBG},
		StageObject{Category: StageCategoryProp, Z: 2},
	)

	order := stage.RenderOrder()

	want := []int{1, 2, 0}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v (background first, then Z ascending)", order, want)
		}
	}
}

func TestRenderOrderKeepsCategoryOrderAtEqualZ(t *testing.T) {
	stage := testStageWithLayers(
		StageObject{Category: StageCategoryBlock},
		StageObject{Category: StageCategoryAnimBlock},
		StageObject{Category: StageCategoryProp},
		StageObject{Category: StageCategoryAnimProp},
		StageObject{Category: StageCategoryBG},
	)

	order := stage.RenderOrder()

	want := []int{4, 2, 3, 0, 1}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v (background, prop, anim prop, block, anim block)", order, want)
		}
	}
}

func TestRenderOrderLayersDominateZ(t *testing.T) {
	stage := testStageWithLayers(
		StageObject{Layer: "Back", Category: StageCategoryBlock, Z: 100},
		StageObject{Layer: "Front", Category: StageCategoryBG, Z: -100},
	)

	order := stage.RenderOrder()

	if len(order) != 2 || order[0] != 0 || order[1] != 1 {
		t.Fatalf("order = %v, want the back layer first regardless of Z", order)
	}
}

func TestRenderOrderSortsStartMarkerLastForPicking(t *testing.T) {
	stage := testStageWithLayers(
		StageObject{Category: StageCategoryStart},
		StageObject{Category: StageCategoryBG},
	)

	order := stage.RenderOrder()

	if len(order) != 2 || order[0] != 1 || order[1] != 0 {
		t.Fatalf("order = %v, want the start marker last so picking reaches it first", order)
	}
	if StageCategoryDraws(StageCategoryStart) {
		t.Fatal("the start marker must never paint a sprite")
	}
}
