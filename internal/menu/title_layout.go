package menu

const (
	TrackDuration    = 173.0
	TitleFillStart   = 40.0
	TitleFillFadeOut = 8.0

	CycleInterval      = 10.0
	ColorTransitionDur = 2.0

	BorderInitialR = 1.0
	BorderInitialG = 1.0
	BorderInitialB = 1.0
	BorderInitialA = 0.7
)

type BorderColor struct {
	R, G, B, A float32
}

var BorderCycleColors = []BorderColor{
	{R: 253.0 / 255, G: 128.0 / 255, B: 1, A: 0.7},
	{R: 1, G: 8.0 / 255, B: 8.0 / 255, A: 0.7},
	{R: 1, G: 238.0 / 255, B: 0, A: 0.8},
	{R: 0, G: 217.0 / 255, B: 1, A: 0.7},
	{R: 198.0 / 255, G: 83.0 / 255, B: 1, A: 0.7},
	{R: 1, G: 1, B: 1, A: 0.7},
}

const (
	TitleWidth       = 912
	TitleHeight      = 208
	TitleBaseScale   = 1.2
	TitleY           = 220.0
	TitleDesignWidth = 1280

	CosmicTravelX = 0.012
	CosmicTravelY = 0.04
	CosmicWobbleX = 0.2
	CosmicWobbleY = 0.1
	CosmicJitterX = 0.25
	CosmicJitterY = 0.25
)
