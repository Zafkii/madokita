package game

import (
	"path/filepath"
	"strings"

	"madokita-gems/internal/assets"
	generated "madokita-gems/internal/game/stages"
	stage "madokita-gems/internal/stage"
)

// TestStageDef is the sample stage generated from internal/game/stages. Save
// the Stage Editor output there and point this at the exported variable.
var TestStageDef = NewStageDef("teststage", "Stage de Prueba", &generated.TestStage, 700, 50)

// NewStageDef adapts a generated stage layout to the runtime stage definition.
// Generated files keep image paths relative to the .go file so the editor can
// reopen them; the runtime needs paths relative to the game asset root.
func NewStageDef(id, name string, layout *stage.Stage, groundY, spawnX float64) *assets.StageDef {
	if layout != nil {
		prepareStageLayout(layout)
		for i := range layout.Assets {
			asset := &layout.Assets[i]
			asset.File = runtimeStageAssetPath(asset.File)
			for j := range asset.Sprites {
				asset.Sprites[j].File = runtimeStageAssetPath(asset.Sprites[j].File)
			}
		}
		for i := range layout.Objects {
			layout.Objects[i].File = runtimeStageAssetPath(layout.Objects[i].File)
		}
	}
	return &assets.StageDef{
		ID:      id,
		Name:    name,
		Layout:  layout,
		GroundY: groundY,
		SpawnX:  spawnX,
	}
}

func prepareStageLayout(layout *stage.Stage) {
	if layout.Width <= 0 {
		layout.Width = StageWidth
	}
	if layout.Height <= 0 {
		layout.Height = StageHeight
	}

	layers := make([]stage.Layer, 0, len(layout.Layers))
	seen := make(map[string]bool, len(layout.Layers))
	for _, layer := range layout.Layers {
		if layer.Name == "" || seen[layer.Name] {
			continue
		}
		seen[layer.Name] = true
		layers = append(layers, layer)
	}
	if len(layers) == 0 {
		layer := stage.Layer{Name: "Layer 1"}
		layers = append(layers, layer)
		seen[layer.Name] = true
	}
	layout.Layers = layers

	defaultLayer := layers[len(layers)-1].Name
	for i := range layout.Objects {
		if !seen[layout.Objects[i].Layer] {
			layout.Objects[i].Layer = defaultLayer
		}
	}
}

// runtimeStageAssetPath converts a path relative to a generated stage file
// (or an absolute editor path) into a path relative to the game asset root.
func runtimeStageAssetPath(path string) string {
	clean := filepath.ToSlash(filepath.Clean(path))
	if clean == "." || clean == "" || clean == "assets" {
		return ""
	}
	if strings.HasPrefix(clean, "assets/") {
		return strings.TrimPrefix(clean, "assets/")
	}
	if i := strings.LastIndex(clean, "/assets/"); i >= 0 {
		return clean[i+len("/assets/"):]
	}
	return clean
}
