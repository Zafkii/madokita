package stage

import "testing"

func TestEffectiveCollisionBoxes(t *testing.T) {
	block := Object{Category: CategoryBlock, Width: 50, Height: 20}
	boxes := block.EffectiveCollisionBoxes()
	if len(boxes) != 1 || boxes[0] != (Box{Width: 50, Height: 20}) {
		t.Fatalf("block boxes = %+v, want default 50x20", boxes)
	}

	block.CollisionBoxes = []Box{{X: 3, Y: 4, Width: 10, Height: 12}}
	boxes = block.EffectiveCollisionBoxes()
	if len(boxes) != 1 || boxes[0] != (Box{X: 3, Y: 4, Width: 10, Height: 12}) {
		t.Fatalf("authored block boxes = %+v, want authored box", boxes)
	}

	prop := Object{Category: CategoryProp, Width: 20, Height: 20}
	if boxes := prop.EffectiveCollisionBoxes(); len(boxes) != 0 {
		t.Fatalf("prop without authored box = %+v, want none", boxes)
	}
	prop.CollisionBoxes = []Box{{Width: 5, Height: 6}}
	if boxes := prop.EffectiveCollisionBoxes(); len(boxes) != 1 {
		t.Fatalf("prop with authored box = %+v, want one", boxes)
	}

	for _, obj := range []Object{
		{Category: CategoryBG, Width: 20, Height: 20},
		{Category: CategoryStart},
		{Category: CategoryBlock, Width: 20, Height: 20, CollisionDisabled: true},
	} {
		if boxes := obj.EffectiveCollisionBoxes(); len(boxes) != 0 {
			t.Errorf("%q boxes = %+v, want none", obj.Category, boxes)
		}
	}
}

func TestDefaultCollisionBoxFallsBackToBlockSize(t *testing.T) {
	boxes := (Object{Category: CategoryBlock}).EffectiveCollisionBoxes()
	if len(boxes) != 1 || boxes[0].Width != DefaultBlockSize || boxes[0].Height != DefaultBlockSize {
		t.Fatalf("boxes = %+v, want %vx%v", boxes, DefaultBlockSize, DefaultBlockSize)
	}
}
