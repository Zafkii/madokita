package project

import (
	"path/filepath"
	"strings"
)

const (
	StageVersion       = 1
	StageCategoryBG    = "bg"
	StageCategoryBlock = "block"
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

// StageObject is one image placed in the stage. X/Y are the object center.
type StageObject struct {
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
		Assets:  []StageAsset{},
		Objects: []StageObject{},
	}
}

func (l StageData) DeepCopy() StageData {
	out := l
	out.Assets = append([]StageAsset(nil), l.Assets...)
	out.Objects = append([]StageObject(nil), l.Objects...)
	return out
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
