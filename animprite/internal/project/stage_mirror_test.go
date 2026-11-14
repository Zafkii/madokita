package project

import "testing"

func TestFlipXMirrorsAuthoredBoxes(t *testing.T) {
	obj := StageObject{
		Category:       StageCategoryProp,
		Width:          40,
		Height:         20,
		FlipX:          true,
		CollisionBoxes: []StageBox{{X: 6, Y: 2, Width: 10, Height: 8, Rotation: 15}},
		Hitboxes:       []StageBox{{X: -3, Y: 1, Width: 4, Height: 5, Rotation: -20}},
	}

	boxes := obj.EffectiveCollisionBoxes()
	if len(boxes) != 1 {
		t.Fatalf("collision boxes = %+v, want one", boxes)
	}
	if boxes[0].X != -6 || boxes[0].Y != 2 || boxes[0].Rotation != -15 {
		t.Fatalf("mirrored collision box = %+v, want X -6, Y 2, Rotation -15", boxes[0])
	}

	hitboxes := obj.EffectiveHitboxes()
	if len(hitboxes) != 1 {
		t.Fatalf("hitboxes = %+v, want one", hitboxes)
	}
	if hitboxes[0].X != 3 || hitboxes[0].Rotation != 20 {
		t.Fatalf("mirrored hitbox = %+v, want X 3, Rotation 20", hitboxes[0])
	}

	if obj.CollisionBoxes[0].X != 6 || obj.Hitboxes[0].X != -3 {
		t.Fatalf("authored boxes changed: collision %+v hitboxes %+v", obj.CollisionBoxes, obj.Hitboxes)
	}
}

func TestFlipXKeepsDefaultBlockBoxCentered(t *testing.T) {
	obj := StageObject{Category: StageCategoryBlock, Width: 50, Height: 20, FlipX: true}

	boxes := obj.EffectiveCollisionBoxes()
	if len(boxes) != 1 || boxes[0] != (StageBox{Width: 50, Height: 20}) {
		t.Fatalf("flipped block boxes = %+v, want centered 50x20", boxes)
	}
}
