package menu

import (
	"image/color"
	"madokita/internal/audio"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

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

	cosmicX float64
	cosmicY float64

	targetR, targetG, targetB, targetA float32
	lerpFrom, lerpTo                   BorderColor
	lerpT                              float64
	cycleIdx                           int
	cycleTimer                         float64
	cycleActive                        bool

	audioMgr *audio.AudioManager
	entered  bool
	lastPos  float64
}

func NewAnimatedTitle(base, overlay, cosmic *ebiten.Image, audioMgr *audio.AudioManager) *AnimatedTitle {
	b := base.Bounds()
	nw, nh := b.Dx(), b.Dy()

	sx := (TitleDesignWidth - float64(nw)*TitleBaseScale) / 2

	pix := ebiten.NewImage(1, 1)
	pix.Fill(color.White)

	t := &AnimatedTitle{
		base:    base,
		overlay: overlay,
		cosmic:  cosmic,
		pixel:   pix,
		nw:      nw, nh: nh,
		screenScale: TitleBaseScale,
		screenX:     sx,
		screenY:     TitleY - float64(nh)*TitleBaseScale/2,
		audioMgr:    audioMgr,
	}
	t.setColor(BorderCycleColors[len(BorderCycleColors)-1])
	t.lastPos = -1
	return t
}

func (a *AnimatedTitle) setColor(c BorderColor) {
	a.targetR, a.targetG, a.targetB, a.targetA = c.R, c.G, c.B, c.A
	a.lerpFrom, a.lerpTo = c, c
	a.lerpT = 1
}

func (a *AnimatedTitle) Reset() {
	a.cycleActive = false
	a.cycleTimer = 0
	a.cycleIdx = 0
	a.setColor(BorderCycleColors[len(BorderCycleColors)-1])
	a.cosmicX = 0
	a.cosmicY = 0
	a.entered = true
	a.lastPos = -1
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

func (a *AnimatedTitle) Update(dt float64) {
	if !a.entered {
		return
	}

	pos := a.musicPos()

	if a.lastPos > 0 && pos < a.lastPos-TrackDuration*0.1 {
		a.cycleActive = false
		a.cycleTimer = 0
		a.cycleIdx = 0
		a.setColor(BorderCycleColors[len(BorderCycleColors)-1])
		a.cosmicX = 0
		a.cosmicY = 0
	}
	a.lastPos = pos

	if a.cycleActive {
		t := pos * 0.00018
		a.cosmicX += CosmicTravelX*dt + math.Sin(t*0.002)*CosmicWobbleX + math.Cos(t*0.02)*CosmicJitterX
		a.cosmicY += CosmicTravelY*dt + math.Sin(t*0.002)*CosmicWobbleY + math.Cos(t*0.06)*CosmicJitterY
	}

	if pos >= TitleFillStart && !a.cycleActive {
		a.cycleActive = true
		a.cycleTimer = 0
		a.cycleIdx = 0
		last := len(BorderCycleColors) - 1
		a.lerpFrom = BorderCycleColors[last]
		a.lerpTo = BorderCycleColors[0]
		a.lerpT = 0
	}

	if a.cycleActive {
		a.cycleTimer += dt
		if a.cycleTimer >= CycleInterval && a.lerpT >= 1 {
			a.cycleTimer -= CycleInterval
			a.cycleIdx++
			if a.cycleIdx >= len(BorderCycleColors)-1 {
				a.cycleIdx = 0
			}
			a.startCycle()
		}
		if a.lerpT < 1 {
			a.lerpT += dt / ColorTransitionDur
			if a.lerpT > 1 {
				a.lerpT = 1
			}
			from, to := a.lerpFrom, a.lerpTo
			a.targetR = from.R + (to.R-from.R)*float32(a.lerpT)
			a.targetG = from.G + (to.G-from.G)*float32(a.lerpT)
			a.targetB = from.B + (to.B-from.B)*float32(a.lerpT)
			a.targetA = from.A + (to.A-from.A)*float32(a.lerpT)
		}
	}
}

func (a *AnimatedTitle) startCycle() {
	a.lerpFrom = BorderCycleColors[a.cycleIdx]
	next := a.cycleIdx + 1
	if next >= len(BorderCycleColors)-1 {
		next = 0
	}
	a.lerpTo = BorderCycleColors[next]
	a.lerpT = 0
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
