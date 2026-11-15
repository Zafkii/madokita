package stage

import (
	"math"

	math2 "madokita-gems/internal/math"
)

const (
	CategoryBG        = "bg"
	CategoryProp      = "prop"
	CategoryBlock     = "block"
	CategoryAnimProp  = "anim_prop"
	CategoryAnimBlock = "anim_block"
	// CategoryStart marks the single player start-position marker in a stage
	// layout. The marker has no image; its X/Y define where the player spawns.
	CategoryStart = "start"

	DefaultBlockSize = 50.0
)

// Stage is a serializable stage layout: the asset catalog plus placed objects.
type Stage struct {
	Name    string
	Width   float64
	Height  float64
	Layers  []Layer
	Assets  []Asset
	Objects []Object
}

// Layer is one named stack of objects. Layers render in slice order.
type Layer struct {
	Name string
}

// SpriteSheet describes one source spritesheet used by an animated stage
// asset. FrameW/FrameH are the size of one cell inside the sheet.
type SpriteSheet struct {
	File       string
	FrameW     int
	FrameH     int
	FrameCount int
}

// AnimSprite is one sprite drawn inside an animation frame. Offsets are in
// pixels from the frame origin; OriginX/OriginY are normalized 0..1 anchors
// inside the source cell.
type AnimSprite struct {
	SpriteIdx      int
	SpriteFrameIdx int
	OffsetX        float64
	OffsetY        float64
	Rotation       float64
	ScaleX         float64
	ScaleY         float64
	OriginX        float64
	OriginY        float64
}

// AnimFrame is one composited frame of a stage animation.
type AnimFrame struct {
	Sprites []AnimSprite
}

// Animation is the runtime description of an animated stage asset. The editor
// exports it so the game can compose frames without parsing movement files.
type Animation struct {
	Name   string
	FPS    float64
	Loop   bool
	FrameW int
	FrameH int
	Frames []AnimFrame
}

// Asset is one image available in the stage editor library.
type Asset struct {
	Name       string
	File       string
	Category   string
	Width      int
	Height     int
	FrameCount int
	AnimFile   string
	AnimName   string
	Sprites    []SpriteSheet
	Animation  *Animation
}

// Box is one collision or hit box owned by an object. X/Y are offsets from
// the object center; Rotation is in degrees and adds to the object angle.
type Box struct {
	X        float64
	Y        float64
	Width    float64
	Height   float64
	Rotation float64
}

// Object is one image placed in the stage. X/Y are the object center.
type Object struct {
	Layer             string
	Category          string
	File              string
	AnimFile          string
	AnimName          string
	X                 float64
	Y                 float64
	Width             float64
	Height            float64
	Rotation          float64
	Opacity           float64
	FlipX             bool
	CollisionDisabled bool
	CollisionBoxes    []Box
	Hitboxes          []Box
}

// CategoryIsBlock reports whether category uses the block collision model.
func CategoryIsBlock(category string) bool {
	return category == CategoryBlock || category == CategoryAnimBlock
}

// DefaultCollisionBox returns the implicit box of a block. It matches the
// rendered object and is centered on it.
func (o Object) DefaultCollisionBox() Box {
	w := o.Width
	h := o.Height
	if w <= 0 {
		w = DefaultBlockSize
	}
	if h <= 0 {
		h = DefaultBlockSize
	}
	return Box{Width: w, Height: h}
}

// EffectiveCollisionBoxes returns the solid boxes the runtime must resolve.
// Backgrounds, start markers, and disabled objects have none. Blocks without
// authored boxes use their rendered size; props only collide when a box was
// authored explicitly. Hitboxes are damage-only and never included here.
func (o Object) EffectiveCollisionBoxes() []Box {
	if o.Category == CategoryBG || o.Category == CategoryStart || o.CollisionDisabled {
		return nil
	}
	if len(o.CollisionBoxes) > 0 {
		if o.FlipX {
			return mirrorBoxes(o.CollisionBoxes)
		}
		return o.CollisionBoxes
	}
	if CategoryIsBlock(o.Category) {
		return []Box{o.DefaultCollisionBox()}
	}
	return nil
}

// SolidRects converts the object's effective collision boxes into world-space
// solid rectangles the runtime resolves against. An axis-aligned box (object and
// box rotation at a multiple of 90 degrees) yields one exact rectangle. A
// rotated box is decomposed into thin vertical strips whose top edge follows the
// rotated contour, so a sloped surface stays close to its authored height
// instead of blocking the player with the rotated box's full bounding box.
func (o Object) SolidRects() []math2.Rect {
	boxes := o.EffectiveCollisionBoxes()
	if len(boxes) == 0 {
		return nil
	}
	rects := make([]math2.Rect, 0, len(boxes))
	for _, box := range boxes {
		rects = append(rects, o.solidStrips(box)...)
	}
	return rects
}

// SolidStripWidth is the horizontal width of each axis-aligned strip used to
// approximate a rotated collision box. Thin strips keep the step height between
// neighbours small enough for the player's auto step-up to walk them smoothly.
const SolidStripWidth = 6.0

// solidStrips decomposes one collision box into axis-aligned strips. Each strip
// spans a slice of the box's world-space X extent and covers the full vertical
// extent of the rotated box inside that slice, which is conservative: the strip
// never sticks out sideways past the slice, only upwards to the highest point
// the surface reaches there.
func (o Object) solidStrips(box Box) []math2.Rect {
	if box.Width <= 0 || box.Height <= 0 {
		return nil
	}

	rot := (o.Rotation + box.Rotation) * math.Pi / 180
	cos, sin := math.Cos(rot), math.Sin(rot)
	cx := o.X + box.X*cos - box.Y*sin
	cy := o.Y + box.X*sin + box.Y*cos
	hw, hh := box.Width/2, box.Height/2

	corners := [4][2]float64{
		{cx - hw*cos + hh*sin, cy - hw*sin - hh*cos},
		{cx + hw*cos + hh*sin, cy + hw*sin - hh*cos},
		{cx + hw*cos - hh*sin, cy + hw*sin + hh*cos},
		{cx - hw*cos - hh*sin, cy - hw*sin + hh*cos},
	}

	minX, maxX := corners[0][0], corners[0][0]
	minY, maxY := corners[0][1], corners[0][1]
	for _, c := range corners[1:] {
		minX = math.Min(minX, c[0])
		maxX = math.Max(maxX, c[0])
		minY = math.Min(minY, c[1])
		maxY = math.Max(maxY, c[1])
	}

	// Axis-aligned rotations keep the exact rectangle; no decomposition needed.
	if math.Abs(sin) < 1e-6 || math.Abs(cos) < 1e-6 {
		if maxX-minX <= 0 || maxY-minY <= 0 {
			return nil
		}
		return []math2.Rect{math2.NewRect(minX, minY, maxX-minX, maxY-minY)}
	}
	if maxX-minX <= 0 {
		return nil
	}

	count := int(math.Ceil((maxX - minX) / SolidStripWidth))
	if count < 1 {
		count = 1
	}
	stripW := (maxX - minX) / float64(count)

	strips := make([]math2.Rect, 0, count)
	for i := 0; i < count; i++ {
		x0 := minX + float64(i)*stripW
		x1 := x0 + stripW

		top, bottom := math.Inf(1), math.Inf(-1)
		extend := func(y float64) {
			top = math.Min(top, y)
			bottom = math.Max(bottom, y)
		}
		// Corners inside the slice are extremes of the rotated contour.
		for _, c := range corners {
			if c[0] >= x0-1e-9 && c[0] <= x1+1e-9 {
				extend(c[1])
			}
		}
		// Edges crossing the slice boundaries bound the rest of the contour.
		for e := 0; e < 4; e++ {
			a, b := corners[e], corners[(e+1)%4]
			if math.Abs(b[0]-a[0]) < 1e-9 {
				continue
			}
			for _, x := range [2]float64{x0, x1} {
				if (x < a[0] && x < b[0]) || (x > a[0] && x > b[0]) {
					continue
				}
				t := (x - a[0]) / (b[0] - a[0])
				extend(a[1] + t*(b[1]-a[1]))
			}
		}
		if math.IsInf(top, 1) || bottom-top <= 0 {
			continue
		}
		strips = append(strips, math2.NewRect(x0, top, stripW, bottom-top))
	}
	return strips
}

// EffectiveHitboxes returns the authored damage boxes for this object.
// Backgrounds, start markers, and disabled objects have none. Hitboxes are
// damage-only and never take part in solid collision.
func (o Object) EffectiveHitboxes() []Box {
	if o.Category == CategoryBG || o.Category == CategoryStart || o.CollisionDisabled {
		return nil
	}
	if o.FlipX {
		return mirrorBoxes(o.Hitboxes)
	}
	return o.Hitboxes
}

// mirrorBoxes mirrors box offsets and rotations across the object's local
// vertical axis, matching the horizontal flip of its rendered image.
func mirrorBoxes(boxes []Box) []Box {
	if len(boxes) == 0 {
		return nil
	}
	out := make([]Box, len(boxes))
	for i, box := range boxes {
		box.X = -box.X
		box.Rotation = -box.Rotation
		out[i] = box
	}
	return out
}
