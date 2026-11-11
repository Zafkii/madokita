package project

import (
	"path/filepath"
	"strings"
)

const (
	StageVersion       = 1
	StageCategoryBG    = "bg"
	StageCategoryBlock = "block"
	DefaultStageLayer  = "Layer 1"
	DefaultBlockSize   = 50.0
	DefaultStageWidth  = 2560.0
	DefaultStageHeight = 720.0
)

// StageAsset describes one image available in the stage asset library.
type StageAsset struct {
	Category   string `json:"category"`
	Name       string `json:"name"`
	File       string `json:"file"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	FrameCount int    `json:"frameCount,omitempty"`
}

// StageLayer is one named stack of stage objects. Layers render bottom to top
// in the order they appear in StageData.Layers.
type StageLayer struct {
	Name string `json:"name"`
}

// StageObject is one image placed in the stage. X/Y are the object center.
// Layer is the name of the StageLayer that owns the object.
type StageObject struct {
	Layer    string  `json:"layer,omitempty"`
	Category string  `json:"category"`
	File     string  `json:"file"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Rotation float64 `json:"rotation"`
	Opacity  float64 `json:"opacity"`
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
	out.Objects = append([]StageObject(nil), l.Objects...)
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

// StageCategoryForFile classifies an image by its filename. Backgrounds are
// detected by "bg", "background" or "sky"; everything else is a block asset.
func StageCategoryForFile(name string) string {
	base := filepath.Base(name)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.ToLower(base)
	for _, marker := range []string{"bg", "background", "sky"} {
		if strings.Contains(base, marker) {
			return StageCategoryBG
		}
	}
	return StageCategoryBlock
}
