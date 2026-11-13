package assets

import "madokita-gems/internal/stage"

type AssetEntry struct {
	Key        string
	Path       string
	Group      LayerGroup
	FrameW     int
	FrameH     int
	FrameCount int
}

type CharacterSheets struct {
	Key    string
	Sheets []AssetEntry
}

type StageDef struct {
	ID         string
	Name       string
	Images     []AssetEntry
	Layout     *stage.Stage
	Characters []string
	Enemies    []string
	BGM        string
	GroundY    float64
	SpawnX     float64
}
