package assets

import "github.com/hajimehoshi/ebiten/v2/colorm"

type TintChannel string

const (
	TintBG      TintChannel = "bg"
	TintEnv     TintChannel = "env"
	TintChar    TintChannel = "character"
	TintOverlay TintChannel = "overlay"
	TintEffect  TintChannel = "effect"
)

var DefaultLayerTint = map[LayerGroup]TintChannel{
	LayerSky:           TintBG,
	LayerClouds:        TintBG,
	LayerMountainsFar:  TintBG,
	LayerMountainsNear: TintBG,
	LayerStructFar:     TintBG,
	LayerStructNear:    TintEnv,
	LayerFloor:         TintEnv,
	LayerCharBack:      TintChar,
	LayerCharMain:      TintChar,
	LayerCharFront:     TintChar,
	LayerOverlay:       TintOverlay,
	LayerEffect:        TintEffect,
}

type TintController struct {
	channels map[TintChannel]colorm.ColorM
}

func NewTintController() *TintController {
	tc := &TintController{
		channels: make(map[TintChannel]colorm.ColorM),
	}
	for _, ch := range allChannels() {
		tc.channels[ch] = colorm.ColorM{}
	}
	return tc
}

func (tc *TintController) Set(ch TintChannel, m colorm.ColorM) {
	tc.channels[ch] = m
}

func (tc *TintController) ResetAll() {
	for ch := range tc.channels {
		tc.channels[ch] = colorm.ColorM{}
	}
}

func (tc *TintController) ForGroup(g LayerGroup) colorm.ColorM {
	ch, ok := DefaultLayerTint[g]
	if !ok {
		return colorm.ColorM{}
	}
	m, ok := tc.channels[ch]
	if !ok {
		return colorm.ColorM{}
	}
	return m
}

func (tc *TintController) Channel(g LayerGroup) TintChannel {
	return DefaultLayerTint[g]
}

func allChannels() []TintChannel {
	return []TintChannel{TintBG, TintEnv, TintChar, TintOverlay, TintEffect}
}

func IdentityMatrix() colorm.ColorM {
	return colorm.ColorM{}
}

func WhiteMatrix() colorm.ColorM {
	var m colorm.ColorM
	m.Scale(0, 0, 0, 1)
	m.Translate(1, 1, 1, 0)
	return m
}

func BlackMatrix() colorm.ColorM {
	var m colorm.ColorM
	m.Scale(0, 0, 0, 1)
	return m
}
