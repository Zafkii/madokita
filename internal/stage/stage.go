package stage

// CategoryStart marks the single player start-position marker in a stage
// layout. The marker has no image; its X/Y define where the player spawns.
const CategoryStart = "start"

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
