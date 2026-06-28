package assets

type LayerGroup int

const (
	LayerSky LayerGroup = iota
	LayerClouds
	LayerMountainsFar
	LayerMountainsNear
	LayerStructFar
	LayerStructNear
	LayerFloor
	LayerCharBack
	LayerCharMain
	LayerCharFront
	LayerOverlay
	LayerEffect
)

var layerGroupNames = map[LayerGroup]string{
	LayerSky:           "sky",
	LayerClouds:        "clouds",
	LayerMountainsFar:  "mountains_far",
	LayerMountainsNear: "mountains_near",
	LayerStructFar:     "struct_far",
	LayerStructNear:    "struct_near",
	LayerFloor:         "floor",
	LayerCharBack:      "char_back",
	LayerCharMain:      "char_main",
	LayerCharFront:     "char_front",
	LayerOverlay:       "overlay",
	LayerEffect:        "effect",
}

func (g LayerGroup) String() string {
	if n, ok := layerGroupNames[g]; ok {
		return n
	}
	return "unknown"
}

func (g LayerGroup) Valid() bool {
	return g >= LayerSky && g <= LayerEffect
}
