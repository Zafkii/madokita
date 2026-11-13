package stage

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
		return o.CollisionBoxes
	}
	if CategoryIsBlock(o.Category) {
		return []Box{o.DefaultCollisionBox()}
	}
	return nil
}
