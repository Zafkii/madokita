package debug

import (
	"fmt"
	"image/color"
	"math"

	"madokita/internal/entity/player"
	"madokita/internal/input"
	"madokita/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var pixel *ebiten.Image

func init() {
	pixel = ebiten.NewImage(1, 1)
	pixel.Fill(color.White)
}

func drawColoredRect(screen *ebiten.Image, geoM ebiten.GeoM, c colorm.ColorM) {
	op := &colorm.DrawImageOptions{}
	op.GeoM = geoM
	colorm.DrawImage(screen, pixel, c, op)
}

func DrawHurtboxes(screen *ebiten.Image, p *player.Player, cameraX float64) {
	if p.Animations.Animator == nil {
		return
	}
	frame := p.Animations.Animator.CurrentFrame()
	if frame == nil {
		return
	}
	for _, hb := range frame.Hurtboxes {
		ox := hb.OffsetX * p.Scale
		oy := hb.OffsetY * p.Scale
		if p.FlipX {
			ox = -ox
		}

		w := hb.W * math.Abs(hb.ScaleX) * p.Scale
		hh := hb.H * math.Abs(hb.ScaleY) * p.Scale

		geoM := ebiten.GeoM{}
		geoM.Translate(-cameraX, 0)
		geoM.Translate(p.X, p.Y)
		geoM.Translate(ox, oy)
		if hb.Rotation != 0 {
			geoM.Rotate(hb.Rotation * math.Pi / 180)
		}
		geoM.Translate(-w/2, -hh/2)
		geoM.Scale(w, hh)

		fillCM := colorm.ColorM{}
		fillCM.Scale(1, 0, 0, 0.25)
		drawColoredRect(screen, geoM, fillCM)

		borders := [][4]float64{
			{0, 0, w, 1},
			{0, hh - 1, w, 1},
			{0, 0, 1, hh},
			{w - 1, 0, 1, hh},
		}
		outCM := colorm.ColorM{}
		outCM.Scale(1, 0, 0, 0.9)
		for _, b := range borders {
			bg := ebiten.GeoM{}
			bg = geoM
			bg.Translate(b[0], b[1])
			bg.Scale(b[2], b[3])
			drawColoredRect(screen, bg, outCM)
		}
	}
}

func DrawSpriteBBox(screen *ebiten.Image, p *player.Player, cameraX float64) {
	sw := float64(player.FrameSize) * p.Scale
	sh := float64(player.FrameSize) * p.Scale
	ox := -p.AnimDef.DefaultOriginX * sw
	oy := -p.AnimDef.DefaultOriginY * sh

	fillGM := ebiten.GeoM{}
	fillGM.Translate(-cameraX, 0)
	fillGM.Translate(p.X, p.Y)
	fillGM.Translate(ox, oy)
	fillGM.Scale(sw, sh)

	fillCM := colorm.ColorM{}
	fillCM.Scale(0, 1, 0, 0.15)
	drawColoredRect(screen, fillGM, fillCM)

	borders := [][4]float64{
		{0, 0, sw, 1},
		{0, sh - 1, sw, 1},
		{0, 0, 1, sh},
		{sw - 1, 0, 1, sh},
	}
	outCM := colorm.ColorM{}
	outCM.Scale(0, 1, 0, 0.8)
	for _, b := range borders {
		bg := ebiten.GeoM{}
		bg.Translate(-cameraX, 0)
		bg.Translate(p.X, p.Y)
		bg.Translate(ox, oy)
		bg.Translate(b[0], b[1])
		bg.Scale(b[2], b[3])
		drawColoredRect(screen, bg, outCM)
	}
}

func DrawOrigin(screen *ebiten.Image, p *player.Player, cameraX float64) {
	cx := p.X - cameraX
	cy := p.Y
	l := 8.0

	ebitenutil.DrawLine(screen, cx-l, cy, cx+l, cy, color.RGBA{0, 255, 255, 200})
	ebitenutil.DrawLine(screen, cx, cy-l, cx, cy+l, color.RGBA{0, 255, 255, 200})
}

func DrawGround(screen *ebiten.Image, cameraX float64) {
	ebitenutil.DrawLine(screen, -cameraX, player.GroundY, -cameraX+1280, player.GroundY, color.RGBA{255, 255, 0, 120})
}

func DrawInfo(screen *ebiten.Image, p *player.Player) {
	anim := "none"
	frameIdx := 0
	if p.Animations.Animator != nil {
		anim = p.Animations.Animator.CurrentAnimation()
		frameIdx = p.Animations.Animator.Frame()
	}
	grounded := "grounded"
	if !p.Movement.IsGrounded {
		grounded = "airborne"
	}
	moving := "idle"
	if p.Input != nil {
		if p.Input.IsPressed(input.ActionMoveLeft) || p.Input.IsPressed(input.ActionMoveRight) {
			moving = "moving"
		}
	}

	lines := []string{
		fmt.Sprintf("Pos:  %.1f, %.1f", p.X, p.Y),
		fmt.Sprintf("Anim: %s [%d]", anim, frameIdx),
		fmt.Sprintf("State: %s | %s", grounded, moving),
		fmt.Sprintf("FlipX: %v | Scale: %.1f", p.FlipX, p.Scale),
	}
	clr := color.RGBA{0, 255, 200, 255}
	for i, l := range lines {
		ui.DrawText(screen, l, 8, 8+i*18, 1, clr)
	}
}
