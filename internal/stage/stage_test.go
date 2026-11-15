package stage

import (
	"math"
	"testing"
)

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

func TestFlipXMirrorsCollisionBoxes(t *testing.T) {
	obj := Object{
		Category:       CategoryProp,
		FlipX:          true,
		CollisionBoxes: []Box{{X: 6, Y: 2, Width: 10, Height: 8, Rotation: 15}},
	}

	boxes := obj.EffectiveCollisionBoxes()
	if len(boxes) != 1 || boxes[0].X != -6 || boxes[0].Rotation != -15 {
		t.Fatalf("boxes = %+v, want mirrored X -6, Rotation -15", boxes)
	}
	if obj.CollisionBoxes[0].X != 6 || obj.CollisionBoxes[0].Rotation != 15 {
		t.Fatalf("authored boxes changed: %+v", obj.CollisionBoxes)
	}
}

func TestDefaultCollisionBoxFallsBackToBlockSize(t *testing.T) {
	boxes := (Object{Category: CategoryBlock}).EffectiveCollisionBoxes()
	if len(boxes) != 1 || boxes[0].Width != DefaultBlockSize || boxes[0].Height != DefaultBlockSize {
		t.Fatalf("boxes = %+v, want %vx%v", boxes, DefaultBlockSize, DefaultBlockSize)
	}
}

func TestSolidRectsKeepAxisAlignedBoxesExact(t *testing.T) {
	for _, rotation := range []float64{0, 90, 180, -90} {
		obj := Object{Category: CategoryBlock, X: 100, Y: 200, Width: 50, Height: 20, Rotation: rotation}
		rects := obj.SolidRects()
		if len(rects) != 1 {
			t.Fatalf("rotation %v: rects = %+v, want a single exact rect", rotation, rects)
		}
		wantW, wantH := 50.0, 20.0
		if rotation == 90 || rotation == -90 {
			wantW, wantH = 20, 50
		}
		rect := rects[0]
		if math.Abs(rect.W-wantW) > 1e-9 || math.Abs(rect.H-wantH) > 1e-9 {
			t.Fatalf("rotation %v: size = %vx%v, want %vx%v", rotation, rect.W, rect.H, wantW, wantH)
		}
		if math.Abs(rect.X+rect.W/2-obj.X) > 1e-9 || math.Abs(rect.Y+rect.H/2-obj.Y) > 1e-9 {
			t.Fatalf("rotation %v: center = (%v, %v), want the object center", rotation, rect.X+rect.W/2, rect.Y+rect.H/2)
		}
	}
}

func TestSolidRectsSplitRotatedBoxIntoThinStrips(t *testing.T) {
	const size = 50.0
	obj := Object{Category: CategoryBlock, X: 100, Y: 200, Width: size, Height: size, Rotation: 45}

	rects := obj.SolidRects()
	if len(rects) < 2 {
		t.Fatalf("rects = %+v, want the rotated box split into strips", rects)
	}

	halfDiag := size * math.Sqrt2 / 2
	wantMinX, wantMaxX := obj.X-halfDiag, obj.X+halfDiag
	wantMinY := obj.Y - halfDiag

	minX, maxX, minY, maxY := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
	for _, rect := range rects {
		if rect.W > SolidStripWidth+1e-9 {
			t.Fatalf("strip width = %v, want at most %v", rect.W, SolidStripWidth)
		}
		minX = math.Min(minX, rect.X)
		maxX = math.Max(maxX, rect.X+rect.W)
		minY = math.Min(minY, rect.Y)
		maxY = math.Max(maxY, rect.Y+rect.H)
	}
	if math.Abs(minX-wantMinX) > 1e-9 || math.Abs(maxX-wantMaxX) > 1e-9 {
		t.Fatalf("strips span X %v..%v, want %v..%v", minX, maxX, wantMinX, wantMaxX)
	}
	if math.Abs(minY-wantMinY) > 1e-9 {
		t.Fatalf("topmost strip top = %v, want the rotated box top %v", minY, wantMinY)
	}
	if math.Abs(maxY-(obj.Y+halfDiag)) > 1e-9 {
		t.Fatalf("bottommost strip bottom = %v, want the rotated box bottom %v", maxY, obj.Y+halfDiag)
	}
}

func TestSolidRectsDoNotFillRotatedBoundingBoxCorners(t *testing.T) {
	const size = 50.0
	obj := Object{Category: CategoryBlock, X: 100, Y: 200, Width: size, Height: size, Rotation: 45}
	halfDiag := size * math.Sqrt2 / 2
	cornerX := obj.X - halfDiag + 1
	cornerY := obj.Y - halfDiag + 1

	for _, rect := range obj.SolidRects() {
		inX := cornerX >= rect.X && cornerX <= rect.X+rect.W
		inY := cornerY >= rect.Y && cornerY <= rect.Y+rect.H
		if inX && inY {
			t.Fatalf("strip %+v covers the empty bounding-box corner (%v, %v)", rect, cornerX, cornerY)
		}
	}
}

func TestSolidRectsSkipObjectsWithoutCollision(t *testing.T) {
	for _, obj := range []Object{
		{Category: CategoryBG, Width: 20, Height: 20},
		{Category: CategoryStart},
		{Category: CategoryBlock, Width: 20, Height: 20, CollisionDisabled: true},
		{Category: CategoryProp, Width: 20, Height: 20},
	} {
		if rects := obj.SolidRects(); len(rects) != 0 {
			t.Errorf("%q rects = %+v, want none", obj.Category, rects)
		}
	}
}

func TestEffectiveHitboxes(t *testing.T) {
	obj := Object{Category: CategoryProp, Hitboxes: []Box{{X: 4, Width: 10, Height: 8}}}
	hitboxes := obj.EffectiveHitboxes()
	if len(hitboxes) != 1 || hitboxes[0].X != 4 {
		t.Fatalf("hitboxes = %+v, want the authored box", hitboxes)
	}

	obj.FlipX = true
	hitboxes = obj.EffectiveHitboxes()
	if len(hitboxes) != 1 || hitboxes[0].X != -4 {
		t.Fatalf("flipped hitboxes = %+v, want mirrored X -4", hitboxes)
	}
	if obj.Hitboxes[0].X != 4 {
		t.Fatalf("authored hitboxes changed: %+v", obj.Hitboxes)
	}

	for _, disabled := range []Object{
		{Category: CategoryBG, Hitboxes: obj.Hitboxes},
		{Category: CategoryStart, Hitboxes: obj.Hitboxes},
		{Category: CategoryBlock, Hitboxes: obj.Hitboxes, CollisionDisabled: true},
	} {
		if hitboxes := disabled.EffectiveHitboxes(); len(hitboxes) != 0 {
			t.Errorf("%q hitboxes = %+v, want none", disabled.Category, hitboxes)
		}
	}
}

func TestDrawOrderLayersZAndCategory(t *testing.T) {
	s := Stage{
		Layers: []Layer{{Name: "Back"}, {Name: "Front"}},
		Objects: []Object{
			{Layer: "Front", Category: CategoryBlock, Z: 5},
			{Layer: "Front", Category: CategoryBG},
			{Layer: "Back", Category: CategoryBlock, Z: -10},
			{Layer: "Front", Category: CategoryProp, Z: 5},
		},
	}

	order := DrawOrder(s)

	want := []int{2, 1, 3, 0}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

func TestDrawOrderAtEqualZKeepsCategoryOrder(t *testing.T) {
	s := Stage{
		Layers: []Layer{{Name: "Layer 1"}},
		Objects: []Object{
			{Layer: "Layer 1", Category: CategoryAnimBlock},
			{Layer: "Layer 1", Category: CategoryBlock},
			{Layer: "Layer 1", Category: CategoryAnimProp},
			{Layer: "Layer 1", Category: CategoryProp},
			{Layer: "Layer 1", Category: CategoryBG},
		},
	}

	order := DrawOrder(s)

	want := []int{4, 3, 2, 1, 0}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

func TestDrawOrderSkipsObjectsOutsideListedLayers(t *testing.T) {
	s := Stage{
		Layers: []Layer{{Name: "Layer 1"}},
		Objects: []Object{
			{Layer: "Layer 1", Category: CategoryBG},
			{Layer: "Missing", Category: CategoryBG},
		},
	}

	order := DrawOrder(s)

	if len(order) != 1 || order[0] != 0 {
		t.Fatalf("order = %v, want only the object whose layer is listed", order)
	}
}
