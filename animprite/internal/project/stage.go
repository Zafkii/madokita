package project

import (
	"path/filepath"
	"strings"
)

const (
	StageVersion           = 1
	StageCategoryBG        = "bg"
	StageCategoryProp      = "prop"
	StageCategoryBlock     = "block"
	StageCategoryAnimProp  = "anim_prop"
	StageCategoryAnimBlock = "anim_block"
	StageCategoryStart     = "start"
	DefaultStageLayer      = "Layer 1"
	DefaultBlockSize       = 50.0
	DefaultStartMarkerSize = 32.0
	DefaultStageWidth      = 2560.0
	DefaultStageHeight     = 720.0
)

// StageAsset describes one image available in the stage asset library.
// Animated assets keep the primary spritesheet in File and reference the
// movement definition that describes how to slice and compose it.
type StageAsset struct {
	Category   string `json:"category"`
	Name       string `json:"name"`
	File       string `json:"file"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	FrameCount int    `json:"frameCount,omitempty"`
	AnimFile   string `json:"animFile,omitempty"`
	AnimName   string `json:"animName,omitempty"`
}

// StageLayer is one named stack of stage objects. Layers render bottom to top
// in the order they appear in StageData.Layers.
type StageLayer struct {
	Name string `json:"name"`
}

// StageBox is one collision or hit box owned by an object. X/Y are offsets
// from the object center; Rotation is in degrees and adds to the object angle.
type StageBox struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Rotation float64 `json:"rotation,omitempty"`
}

// StageObject is one image placed in the stage. X/Y are the object center.
// Layer is the name of the StageLayer that owns the object.
//
// Blocks have collision on by default: when CollisionBoxes is empty, a single
// box matching Width/Height is used. Props keep their image size and have no
// collision until an authored box is added. CollisionDisabled turns off both
// the collision and every authored hitbox for the object.
type StageObject struct {
	Layer             string     `json:"layer,omitempty"`
	Category          string     `json:"category"`
	File              string     `json:"file"`
	AnimFile          string     `json:"animFile,omitempty"`
	AnimName          string     `json:"animName,omitempty"`
	X                 float64    `json:"x"`
	Y                 float64    `json:"y"`
	Width             float64    `json:"width"`
	Height            float64    `json:"height"`
	Rotation          float64    `json:"rotation"`
	Opacity           float64    `json:"opacity"`
	FlipX             bool       `json:"flipX,omitempty"`
	CollisionDisabled bool       `json:"collisionDisabled,omitempty"`
	CollisionBoxes    []StageBox `json:"collisionBoxes,omitempty"`
	Hitboxes          []StageBox `json:"hitboxes,omitempty"`
}

// DeepCopy copies the object and its box slices so undo snapshots stay
// independent from later edits.
func (o StageObject) DeepCopy() StageObject {
	o.CollisionBoxes = append([]StageBox(nil), o.CollisionBoxes...)
	o.Hitboxes = append([]StageBox(nil), o.Hitboxes...)
	return o
}

// DefaultCollisionBox returns the implicit collision box of a block: the box
// matches the rendered object size and sits at its center.
func (o StageObject) DefaultCollisionBox() StageBox {
	w := o.Width
	h := o.Height
	if w <= 0 {
		w = DefaultBlockSize
	}
	if h <= 0 {
		h = DefaultBlockSize
	}
	return StageBox{Width: w, Height: h}
}

// EffectiveCollisionBoxes returns the boxes rendered and exported for this
// object. Backgrounds and disabled objects have none. Blocks (including
// animated blocks) with no authored boxes fall back to the default box; props
// only collide when a box was authored explicitly.
func (o StageObject) EffectiveCollisionBoxes() []StageBox {
	if o.Category == StageCategoryBG || StageCategoryIsStart(o.Category) || o.CollisionDisabled {
		return nil
	}
	if len(o.CollisionBoxes) > 0 {
		if o.FlipX {
			return mirrorStageBoxes(o.CollisionBoxes)
		}
		return o.CollisionBoxes
	}
	if StageCategoryIsBlock(o.Category) {
		return []StageBox{o.DefaultCollisionBox()}
	}
	return nil
}

// EffectiveHitboxes returns the authored damage boxes for this object.
func (o StageObject) EffectiveHitboxes() []StageBox {
	if o.Category == StageCategoryBG || StageCategoryIsStart(o.Category) || o.CollisionDisabled {
		return nil
	}
	if o.FlipX {
		return mirrorStageBoxes(o.Hitboxes)
	}
	return o.Hitboxes
}

// mirrorStageBoxes mirrors box offsets and rotations across the object's local
// vertical axis, matching the horizontal flip of the rendered image.
func mirrorStageBoxes(boxes []StageBox) []StageBox {
	if len(boxes) == 0 {
		return nil
	}
	out := make([]StageBox, len(boxes))
	for i, box := range boxes {
		box.X = -box.X
		box.Rotation = -box.Rotation
		out[i] = box
	}
	return out
}

// ApplySizeFromFirstBackground sets Width/Height from the first background
// asset that has a usable image size. It returns false and leaves the stage
// size unchanged when no such background exists.
func (l *StageData) ApplySizeFromFirstBackground() bool {
	for _, asset := range l.Assets {
		category := asset.Category
		if category == "" {
			category = StageCategoryForFile(asset.File)
		}
		if !strings.EqualFold(category, StageCategoryBG) || asset.Width <= 0 || asset.Height <= 0 {
			continue
		}
		l.Width = float64(asset.Width)
		l.Height = float64(asset.Height)
		return true
	}
	return false
}

// StageData is the serialized stage document.
type StageData struct {
	Version int           `json:"version"`
	Name    string        `json:"name"`
	Width   float64       `json:"width"`
	Height  float64       `json:"height"`
	Layers  []StageLayer  `json:"layers"`
	Assets  []StageAsset  `json:"assets"`
	Objects []StageObject `json:"objects"`
}

func NewStageData(name string) StageData {
	if name == "" {
		name = "Stage1"
	}
	return StageData{
		Version: StageVersion,
		Name:    name,
		Width:   DefaultStageWidth,
		Height:  DefaultStageHeight,
		Layers:  []StageLayer{{Name: DefaultStageLayer}},
		Assets:  []StageAsset{},
		Objects: []StageObject{},
	}
}

func (l StageData) DeepCopy() StageData {
	out := l
	out.Layers = append([]StageLayer(nil), l.Layers...)
	out.Assets = append([]StageAsset(nil), l.Assets...)
	out.Objects = make([]StageObject, len(l.Objects))
	for i := range l.Objects {
		out.Objects[i] = l.Objects[i].DeepCopy()
	}
	return out
}

// NormalizeLayers repairs legacy and hand-written stage documents. It makes
// sure at least one layer exists, removes duplicate or empty layer names, adds
// layers referenced by objects, and assigns every object to a known layer.
func (l *StageData) NormalizeLayers() {
	layers := make([]StageLayer, 0, len(l.Layers)+1)
	seen := make(map[string]bool, len(l.Layers)+1)
	for _, layer := range l.Layers {
		name := strings.TrimSpace(layer.Name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		layers = append(layers, StageLayer{Name: name})
	}
	if len(layers) == 0 {
		layers = append(layers, StageLayer{Name: DefaultStageLayer})
		seen[DefaultStageLayer] = true
	}

	for i := range l.Objects {
		name := strings.TrimSpace(l.Objects[i].Layer)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		layers = append(layers, StageLayer{Name: name})
	}

	for i := range l.Objects {
		name := strings.TrimSpace(l.Objects[i].Layer)
		if name == "" || !seen[name] {
			name = layers[0].Name
		}
		l.Objects[i].Layer = name
	}

	l.Layers = layers
}

func (l StageData) LayerIndex(name string) int {
	for i := range l.Layers {
		if l.Layers[i].Name == name {
			return i
		}
	}
	return -1
}

// LayerNameFor returns the resolved layer name for an object. It falls back
// to the first layer so legacy documents and tests stay usable before
// NormalizeLayers runs.
func (l StageData) LayerNameFor(obj StageObject) string {
	name := strings.TrimSpace(obj.Layer)
	if name != "" && l.LayerIndex(name) >= 0 {
		return name
	}
	if len(l.Layers) > 0 {
		return l.Layers[0].Name
	}
	return DefaultStageLayer
}

// StageCategoryForFile classifies an image by its filename or containing
// folder. Backgrounds are detected by "bg", "background" or "sky"; props by
// "prop", "decor" or "scenery"; everything else is a block asset.
func StageCategoryForFile(name string) string {
	base := stageBaseName(name)
	for _, marker := range []string{"bg", "background", "sky"} {
		if strings.Contains(base, marker) {
			return StageCategoryBG
		}
	}
	for _, marker := range []string{"prop", "decor", "scenery"} {
		if strings.Contains(base, marker) {
			return StageCategoryProp
		}
	}
	return StageCategoryBlock
}

// StageCategoryForAnimatedFile classifies an animated asset from the names
// that surround it (folder, movement file, spritesheet, animation). Block-like
// markers win over prop markers; everything else becomes an animated prop.
func StageCategoryForAnimatedFile(name string) string {
	haystack := strings.ToLower(name)
	for _, marker := range []string{"block", "tile", "platform", "floor"} {
		if strings.Contains(haystack, marker) {
			return StageCategoryAnimBlock
		}
	}
	return StageCategoryAnimProp
}

func stageBaseName(name string) string {
	base := filepath.Base(name)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return strings.ToLower(base)
}

// StageCategoryIsAnimated reports whether category refers to an animated
// stage object.
func StageCategoryIsAnimated(category string) bool {
	return category == StageCategoryAnimProp || category == StageCategoryAnimBlock
}

// StageCategoryIsStart reports whether category marks the player start marker.
func StageCategoryIsStart(category string) bool {
	return category == StageCategoryStart
}

// StageCategoryIsBlock reports whether category uses the block collision
// model (implicit collision box from the object size).
func StageCategoryIsBlock(category string) bool {
	return category == StageCategoryBlock || category == StageCategoryAnimBlock
}

// StageCategoryIsProp reports whether category uses the prop collision model
// (collision only when authored explicitly).
func StageCategoryIsProp(category string) bool {
	return category == StageCategoryProp || category == StageCategoryAnimProp
}

// StageCategoryHasCollision reports whether category can own collision and
// hit boxes.
func StageCategoryHasCollision(category string) bool {
	return StageCategoryIsBlock(category) || StageCategoryIsProp(category)
}
