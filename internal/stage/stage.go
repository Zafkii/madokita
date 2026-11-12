package stage

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

// Asset is one image available in the stage editor library.
type Asset struct {
	Name       string
	File       string
	Category   string
	Width      int
	Height     int
	FrameCount int
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
