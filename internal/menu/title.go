package menu

import (
	"image/color"
	"madokita-gems/internal/audio"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type BorderColor struct {
	R, G, B, A float32
}

func titleColor(r, g, b uint8, a float32) BorderColor {
	return BorderColor{
		R: float32(r) / 255,
		G: float32(g) / 255,
		B: float32(b) / 255,
		A: a,
	}
}

// titleInitialColor is the color shown before the cycle starts.
var titleInitialColor = titleColor(255, 255, 255, 0.7)

// titleColorOrder is the title color cycle, in playback order.
// Edit this list to reorder, add, or remove colors: the cycle starts at the
// initial color and then walks this slice in order, looping forever.
var titleColorOrder = []BorderColor{
	titleColor(253, 128, 255, 0.7), // pink
	titleColor(255, 238, 0, 0.8),   // yellow
	titleColor(198, 83, 255, 0.7),  // purple
	titleColor(255, 8, 8, 0.7),     // red
	titleColor(0, 217, 255, 0.7),   // light blue
}

type CosmicMotion int

const (
	CosmicFull CosmicMotion = iota
	CosmicLeft
)

type TitleConfig struct {
	X, Y   float64
	Scale  float64
	Motion CosmicMotion
}

type AnimatedTitle struct {
	base    *ebiten.Image
	overlay *ebiten.Image
	cosmic  *ebiten.Image
	pixel   *ebiten.Image

	nw, nh      int
	screenScale float64
	screenX     float64
	screenY     float64

	offscreen *ebiten.Image
	maskBuf   *ebiten.Image
	tintBuf   *ebiten.Image

	cosmicX      float64
	cosmicY      float64
	cosmicMotion CosmicMotion

	targetR, targetG, targetB, targetA float32
	cycleActive                        bool

	audioMgr *audio.AudioManager
	entered  bool
}

func NewAnimatedTitle(base, overlay, cosmic *ebiten.Image, audioMgr *audio.AudioManager, cfg TitleConfig) *AnimatedTitle {
	b := base.Bounds()
	nw, nh := b.Dx(), b.Dy()

	pix := ebiten.NewImage(1, 1)
	pix.Fill(color.White)

	t := &AnimatedTitle{
		base:         base,
		overlay:      overlay,
		cosmic:       cosmic,
		pixel:        pix,
		nw:           nw,
		nh:           nh,
		screenScale:  cfg.Scale,
		screenX:      cfg.X,
		screenY:      cfg.Y,
		cosmicMotion: cfg.Motion,
		audioMgr:     audioMgr,
	}
	t.setColor(titleInitialColor)
	return t
}

func (a *AnimatedTitle) setColor(c BorderColor) {
	a.targetR, a.targetG, a.targetB, a.targetA = c.R, c.G, c.B, c.A
}

func (a *AnimatedTitle) Reset() {
	a.entered = true
	// Evaluate the timeline now so the first frame drawn after a scene
	// switch already has the right state instead of a white flash or a
	// backward jump of the cosmic texture.
	pos := a.musicPos()
	a.cycleActive = pos >= TitleFillStart
	a.applyCosmicAt(pos)
	a.applyColorAt(pos)
}

func (a *AnimatedTitle) musicPos() float64 {
	if a.audioMgr != nil && a.entered {
		pos, ok := a.audioMgr.GetPosition("menu-theme")
		if ok {
			dur := TrackDuration
			if d, ok := a.audioMgr.GetDuration("menu-theme"); ok && d.Seconds() > 0 {
				dur = d.Seconds()
			}
			return math.Mod(pos.Seconds(), dur)
		}
	}
	return 0
}

func (a *AnimatedTitle) Update(_ float64) {
	if !a.entered {
		return
	}

	pos := a.musicPos()
	a.cycleActive = pos >= TitleFillStart
	a.applyCosmicAt(pos)
	a.applyColorAt(pos)
}

// applyCosmicAt derives the cosmic drift from the music position. The motion
// used to be accumulated frame by frame, which made the texture jump backwards
// whenever a scene rebuilt the title. A time-based offset keeps it continuous
// across scene switches.
func (a *AnimatedTitle) applyCosmicAt(pos float64) {
	if pos < TitleFillStart {
		a.cosmicX = 0
		a.cosmicY = 0
		return
	}

	active := pos - TitleFillStart

	if a.cosmicMotion == CosmicLeft {
		a.cosmicX = -GemsCosmicSpeed * active
		a.cosmicY = 0
		return
	}

	// The original motion added the wobble/jitter terms once per frame.
	// Reference them to 60 FPS so the drift stays frame-rate independent
	// while keeping the calibrated speed.
	const refFPS = 60.0
	t := pos * 0.00018
	a.cosmicX = active * (CosmicTravelX + refFPS*(math.Sin(t*0.002)*CosmicWobbleX+math.Cos(t*0.02)*CosmicJitterX))
	a.cosmicY = active * (CosmicTravelY + refFPS*(math.Sin(t*0.002)*CosmicWobbleY+math.Cos(t*0.06)*CosmicJitterY))
}

// applyColorAt derives the current color from the music position instead of
// accumulating per-scene state. A title created mid-song (for example when
// entering Settings and coming back) therefore resumes the exact color on the
// timeline instead of resetting to the initial white.
func (a *AnimatedTitle) applyColorAt(pos float64) {
	if len(titleColorOrder) == 0 || CycleInterval <= 0 || pos < TitleFillStart {
		a.setColor(titleInitialColor)
		return
	}

	n := len(titleColorOrder)
	elapsed := pos - TitleFillStart
	step := int(elapsed / CycleInterval)
	segT := elapsed - float64(step)*CycleInterval
	to := titleColorOrder[step%n]

	if segT >= ColorTransitionDur {
		a.setColor(to)
		return
	}

	from := titleInitialColor
	if step > 0 {
		from = titleColorOrder[(step-1)%n]
	}
	t := float32(segT / ColorTransitionDur)
	a.setColor(BorderColor{
		R: from.R + (to.R-from.R)*t,
		G: from.G + (to.G-from.G)*t,
		B: from.B + (to.B-from.B)*t,
		A: from.A + (to.A-from.A)*t,
	})
}

func (a *AnimatedTitle) ensureBufs() {
	if a.nw <= 0 || a.nh <= 0 {
		return
	}
	if a.offscreen != nil && a.offscreen.Bounds().Dx() == a.nw && a.offscreen.Bounds().Dy() == a.nh {
		return
	}
	a.offscreen = ebiten.NewImage(a.nw, a.nh)
	a.maskBuf = ebiten.NewImage(a.nw, a.nh)
	a.tintBuf = ebiten.NewImage(a.nw, a.nh)
}

func (a *AnimatedTitle) Draw(screen *ebiten.Image) {
	if !a.entered || a.base == nil {
		return
	}
	a.ensureBufs()

	nwF, nhF := float64(a.nw), float64(a.nh)

	a.offscreen.Clear()
	a.offscreen.DrawImage(a.base, nil)

	if a.cosmic != nil && a.cycleActive {
		a.maskBuf.Clear()
		a.maskBuf.DrawImage(a.base, nil)

		a.tintBuf.Clear()

		cb := a.cosmic.Bounds()
		cw, ch := float64(cb.Dx()), float64(cb.Dy())

		scale := math.Max(nwF/cw, nhF/ch)

		sw, sh := cw*scale, ch*scale

		ox := math.Mod(a.cosmicX, sw)
		oy := math.Mod(a.cosmicY, sh)
		if ox < 0 {
			ox += sw
		}
		if oy < 0 {
			oy += sh
		}

		tilesX := int(math.Ceil(nwF/sw)) + 1
		tilesY := int(math.Ceil(nhF/sh)) + 1

		startX := ox - sw
		startY := oy - sh

		op := &ebiten.DrawImageOptions{}
		for ty := 0; ty < tilesY; ty++ {
			for tx := 0; tx < tilesX; tx++ {
				op.GeoM.Reset()
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(startX+float64(tx)*sw, startY+float64(ty)*sh)
				a.tintBuf.DrawImage(a.cosmic, op)
			}
		}

		a.tintBuf.DrawImage(a.maskBuf, &ebiten.DrawImageOptions{
			CompositeMode: ebiten.CompositeModeDestinationIn,
		})

		a.offscreen.DrawImage(a.tintBuf, nil)
	}

	if a.overlay != nil && a.targetA > 0.001 {
		a.maskBuf.Clear()
		a.maskBuf.DrawImage(a.overlay, nil)

		a.tintBuf.Clear()
		fillOp := &ebiten.DrawImageOptions{}
		fillOp.GeoM.Scale(nwF, nhF)
		fillOp.ColorScale.Scale(a.targetR, a.targetG, a.targetB, a.targetA)
		a.tintBuf.DrawImage(a.pixel, fillOp)

		a.tintBuf.DrawImage(a.maskBuf, &ebiten.DrawImageOptions{
			CompositeMode: ebiten.CompositeModeDestinationIn,
		})

		a.offscreen.DrawImage(a.tintBuf, nil)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(a.screenScale, a.screenScale)
	op.GeoM.Translate(a.screenX, a.screenY)
	screen.DrawImage(a.offscreen, op)
}
