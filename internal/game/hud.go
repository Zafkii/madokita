package game

import (
	"image"
	"math"

	math2 "madokita-gems/internal/math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Dark Souls style stats cluster: the soul gem icon at the top left, with the
// health, soul and stamina bars stacked to its right.
//
// The bar PNGs are hollow frames (one white pixel plus one black pixel of edge,
// transparent interior). The matching *_refill and <character>-soul-bar images
// are the flat fill colour, cropped inside the frame up to the current value.
const (
	hudMargin   = 20.0
	hudIconSize = 60.0
	hudIconGap  = 8.0

	// The bar assets are 400x20 frames; they are painted at half size, which
	// keeps the two-pixel edge as a crisp one-pixel frame.
	hudBarScale    = 0.5
	hudBarAssetW   = 400.0
	hudBarAssetH   = 20.0
	hudBarAssetInX = 2.0
	hudBarAssetInY = 2.0
	hudBarAssetInW = 396.0
	hudBarAssetInH = 16.0
	hudBarW        = hudBarAssetW * hudBarScale
	hudBarH        = hudBarAssetH * hudBarScale
	hudBarGap      = 2.0
	hudBarInnerX   = hudBarAssetInX * hudBarScale
	hudBarInnerY   = hudBarAssetInY * hudBarScale
	hudBarInnerW   = hudBarAssetInW * hudBarScale
	hudBarInnerH   = hudBarAssetInH * hudBarScale
)

var (
	hudIconPath  = "stats/icons/gem-soul-icon.png"
	hudBarFrames = [...]string{
		"stats/bar/health-bar.png",
		"stats/bar/soul-bar.png",
		"stats/bar/stam-bar.png",
	}
	// Sayaka is the only playable character so far, and each character ships its
	// own soul colour.
	hudBarFills = [...]string{
		"stats/bar/health-bar_refill.png",
		"stats/bar/sayaka-soul-bar.png",
		"stats/bar/stam-bar_refill.png",
	}
)

type statsHUD struct {
	icon   *ebiten.Image
	frames [3]*ebiten.Image
	fills  [3]*ebiten.Image
}

// hudLayout places the cluster in the top left corner: the icon sits at the
// margin and the bars start to its right, vertically centred on the icon.
func hudLayout() (icon math2.Rect, bars [3]math2.Rect) {
	icon = math2.NewRect(hudMargin, hudMargin, hudIconSize, hudIconSize)
	stackH := 3*hudBarH + 2*hudBarGap
	barX := icon.X + hudIconSize + hudIconGap
	top := icon.Y + (hudIconSize-stackH)/2
	for i := range bars {
		bars[i] = math2.NewRect(barX, top+float64(i)*(hudBarH+hudBarGap), hudBarW, hudBarH)
	}
	return icon, bars
}

// loadStatsHUD loads the cluster images once. Missing images are kept as nil so
// a missing fill degrades to an empty bar instead of breaking the scene.
func (s *GameScene) loadStatsHUD() {
	if s.statsHUD != nil || s.assetM == nil {
		return
	}
	hud := &statsHUD{}
	if img, err := s.assetM.LoadImage(hudIconPath); err == nil {
		hud.icon = img
	}
	for i, path := range hudBarFrames {
		if img, err := s.assetM.LoadImage(path); err == nil {
			hud.frames[i] = img
		}
	}
	for i, path := range hudBarFills {
		if img, err := s.assetM.LoadImage(path); err == nil {
			hud.fills[i] = img
		}
	}
	s.statsHUD = hud
}

// drawStatsHUD paints the cluster in canvas space. Values come from the
// player's combat stats, so damage and stamina spending move the bars.
func (s *GameScene) drawStatsHUD(screen *ebiten.Image) {
	if s.statsHUD == nil || s.Player == nil || s.Player.Actor == nil {
		return
	}
	stats := s.Player.Actor.Stats
	if stats == nil {
		return
	}

	ratios := [3]float64{
		hudRatio(stats.Health, stats.MaxHealth),
		hudRatio(stats.Soul, stats.MaxSoul),
		hudRatio(stats.Stamina, stats.MaxStamina),
	}
	icon, bars := hudLayout()

	if s.statsHUD.icon != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(icon.X, icon.Y)
		screen.DrawImage(s.statsHUD.icon, op)
	}

	for i, bar := range bars {
		if frame := s.statsHUD.frames[i]; frame != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(hudBarScale, hudBarScale)
			op.GeoM.Translate(bar.X, bar.Y)
			screen.DrawImage(frame, op)
		}
		fill := s.statsHUD.fills[i]
		if fill == nil {
			continue
		}
		fillW := hudFillWidth(ratios[i])
		if fillW <= 0 {
			continue
		}
		src := fill.SubImage(image.Rect(
			0, int(hudBarAssetInY),
			fillW, int(hudBarAssetInY+hudBarAssetInH),
		)).(*ebiten.Image)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(hudBarScale, hudBarScale)
		op.GeoM.Translate(bar.X+hudBarInnerX, bar.Y+hudBarInnerY)
		screen.DrawImage(src, op)
	}
}

// hudRatio turns a value/max pair into a 0..1 fraction, clamped so a stat that
// briefly overshoots its maximum never paints outside the bar.
func hudRatio(value, max float64) float64 {
	if max <= 0 || value <= 0 {
		return 0
	}
	ratio := value / max
	if ratio > 1 {
		return 1
	}
	return ratio
}

// hudFillWidth is the filled width for a 0..1 ratio, in bar *asset* pixels: the
// draw call scales it down, so this matches the width of the cropped source.
func hudFillWidth(ratio float64) int {
	if ratio <= 0 {
		return 0
	}
	if ratio > 1 {
		ratio = 1
	}
	return int(math.Round(hudBarAssetInW * ratio))
}
