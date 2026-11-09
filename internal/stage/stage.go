package stage

// Stage is a serializable stage layout: the asset catalog plus placed objects.
type Stage struct {
	Name    string
	Width   float64
	Height  float64
	Assets  []Asset
	Objects []Object
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

// Object is one image placed in the stage. X/Y are the object center.
type Object struct {
	Category string
	File     string
	X        float64
	Y        float64
	Width    float64
	Height   float64
	Rotation float64
	Opacity  float64
}
