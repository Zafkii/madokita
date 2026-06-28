package game

import "madokita/internal/assets"

var TestStageDef = &assets.StageDef{
	ID:      "teststage",
	Name:    "Stage de Prueba",
	GroundY: 700,
	SpawnX:  50,
	Images: []assets.AssetEntry{
		{Key: "ts_sky", Path: "images/stages/testStage/sky.png", Group: assets.LayerSky},
		{Key: "ts_floor", Path: "images/stages/testStage/floor.png", Group: assets.LayerFloor},
	},
}
