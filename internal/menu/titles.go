package menu

import (
	"madokita-gems/internal/audio"
	"madokita-gems/internal/ui"
)

// newMenuTitles builds the animated Madokita and Gems titles from the image cache.
func newMenuTitles(cache *ui.ImageCache, audioMgr *audio.AudioManager) (*AnimatedTitle, *AnimatedTitle) {
	base := cache.Get("menu/madokita-gems-title.png")
	if base == nil {
		return nil, nil
	}

	b := base.Bounds()
	madokita := NewAnimatedTitle(
		base,
		cache.Get("menu/madokita-gems-title-top.png"),
		cache.Get("menu/cosmic-effect.png"),
		audioMgr,
		TitleConfig{
			X:      (TitleDesignWidth - float64(b.Dx())*TitleBaseScale) / 2,
			Y:      TitleY - float64(b.Dy())*TitleBaseScale/2,
			Scale:  TitleBaseScale,
			Motion: CosmicFull,
		},
	)
	madokita.Reset()

	gemsBase := cache.Get("menu/gemstitle.png")
	if gemsBase == nil {
		return madokita, nil
	}

	gems := NewAnimatedTitle(
		gemsBase,
		cache.Get("menu/gemstop.png"),
		cache.Get("menu/gemseffect.jpg"),
		audioMgr,
		TitleConfig{
			X:      GemsTitleX,
			Y:      GemsTitleY,
			Scale:  GemsTitleScale,
			Motion: CosmicLeft,
		},
	)
	gems.Reset()

	return madokita, gems
}
