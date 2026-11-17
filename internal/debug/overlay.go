package debug

import (
	"fmt"
	"image/color"
	"math"

	"madokita-gems/internal/animation"
	"madokita-gems/internal/entity/player"
	"madokita-gems/internal/input"
	"madokita-gems/internal/stage"
	"madokita-gems/internal/ui"

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

// drawBox renders a world-space box centered at (cx, cy). Rotation is applied
// around the box center, matching the hurtbox/hitbox authoring model.
func drawBox(screen *ebiten.Image, cx, cy, w, h, rotation, cameraX float64, fillCM colorm.ColorM, stroke color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}

	geoM := ebiten.GeoM{}
	geoM.Scale(w, h)
	geoM.Translate(-w/2, -h/2)
	if rotation != 0 {
		geoM.Rotate(rotation)
	}
	geoM.Translate(cx, cy)
	geoM.Translate(-cameraX, 0)

	drawColoredRect(screen, geoM, fillCM)

	x0, y0 := geoM.Apply(0, 0)
	x1, y1 := geoM.Apply(1, 0)
	x2, y2 := geoM.Apply(1, 1)
	x3, y3 := geoM.Apply(0, 1)
	ebitenutil.DrawLine(screen, x0, y0, x1, y1, stroke)
	ebitenutil.DrawLine(screen, x1, y1, x2, y2, stroke)
	ebitenutil.DrawLine(screen, x2, y2, x3, y3, stroke)
	ebitenutil.DrawLine(screen, x3, y3, x0, y0, stroke)
}

func DrawHurtboxes(screen *ebiten.Image, p *player.Player, cameraX float64) {
	// While an attack plays, its frames own the body geometry: the movement
	// animator is frozen on the last movement frame and would draw a stale
	// box. Frames without authored hurtboxes fall back to the movement ones.
	var boxes []animation.FrameHurtbox
	if p.State.IsAttacking {
		boxes = p.CurrentAttackHurtboxes()
	}
	if len(boxes) == 0 {
		if p.Animations.Animator == nil {
			return
		}
		frame := p.Animations.Animator.CurrentFrame()
		if frame == nil {
			return
		}
		boxes = frame.Hurtboxes
	}
	fillCM := colorm.ColorM{}
	fillCM.Scale(1, 1, 0, 0.20)
	stroke := color.RGBA{255, 255, 0, 230}

	for _, hb := range boxes {
		ox := hb.OffsetX * p.Scale
		oy := hb.OffsetY * p.Scale
		if p.FlipX {
			ox = -ox
		}

		w := hb.W * math.Abs(hb.ScaleX) * p.Scale
		hh := hb.H * math.Abs(hb.ScaleY) * p.Scale
		rotation := hb.Rotation * math.Pi / 180

		drawBox(screen, p.X+ox, p.Y+oy, w, hh, rotation, cameraX, fillCM, stroke)
	}
}

func DrawHitboxes(screen *ebiten.Image, p *player.Player, cameraX float64) {
	fillCM := colorm.ColorM{}
	fillCM.Scale(1, 0, 0, 0.30)
	stroke := color.RGBA{255, 0, 0, 230}

	for _, r := range p.ActiveHitboxes() {
		drawBox(screen, r.X+r.W/2, r.Y+r.H/2, r.W, r.H, 0, cameraX, fillCM, stroke)
	}
}

// drawFilledRect renders an axis-aligned world-space rectangle as a single
// uniform fill, used for the collision strips the runtime resolves against.
func drawFilledRect(screen *ebiten.Image, x, y, w, h, cameraX float64, fillCM colorm.ColorM) {
	if w <= 0 || h <= 0 {
		return
	}
	geoM := ebiten.GeoM{}
	geoM.Scale(w, h)
	geoM.Translate(x, y)
	geoM.Translate(-cameraX, 0)
	drawColoredRect(screen, geoM, fillCM)
}

// DrawStageBoxes renders the authored geometry of every stage object: collision
// boxes in blue and damage hitboxes in red. Collision boxes are outlined on
// their authored (rotated) shape and filled with the axis-aligned strips the
// runtime actually resolves against, so tuned slopes can be compared with what
// the player walks on.
func DrawStageBoxes(screen *ebiten.Image, layout *stage.Stage, cameraX float64) {
	if layout == nil {
		return
	}

	collisionFill := colorm.ColorM{}
	collisionFill.Scale(0.31, 0.75, 1, 0.18)
	collisionStroke := color.RGBA{80, 190, 255, 230}

	hitboxFill := colorm.ColorM{}
	hitboxFill.Scale(1, 0.36, 0.36, 0.22)
	hitboxStroke := color.RGBA{255, 92, 92, 230}

	noFill := colorm.ColorM{}
	noFill.Scale(0, 0, 0, 0)

	screenW := float64(screen.Bounds().Dx())
	for i := range layout.Objects {
		obj := &layout.Objects[i]
		objRot := obj.Rotation * math.Pi / 180
		cos, sin := math.Cos(objRot), math.Sin(objRot)
		worldBox := func(box stage.Box) (float64, float64, float64) {
			cx := obj.X + box.X*cos - box.Y*sin
			cy := obj.Y + box.X*sin + box.Y*cos
			return cx, cy, objRot + box.Rotation*math.Pi/180
		}

		for _, box := range obj.EffectiveCollisionBoxes() {
			cx, cy, rot := worldBox(box)
			if !boxOnScreen(cx, cy, box.Width, box.Height, rot, cameraX, screenW) {
				continue
			}
			drawBox(screen, cx, cy, box.Width, box.Height, rot, cameraX, noFill, collisionStroke)
		}
		for _, solid := range obj.SolidRects() {
			if !boxOnScreen(solid.X+solid.W/2, solid.Y+solid.H/2, solid.W, solid.H, 0, cameraX, screenW) {
				continue
			}
			drawFilledRect(screen, solid.X, solid.Y, solid.W, solid.H, cameraX, collisionFill)
		}
		for _, box := range obj.EffectiveHitboxes() {
			cx, cy, rot := worldBox(box)
			if !boxOnScreen(cx, cy, box.Width, box.Height, rot, cameraX, screenW) {
				continue
			}
			drawBox(screen, cx, cy, box.Width, box.Height, rot, cameraX, hitboxFill, hitboxStroke)
		}
	}
}

// boxOnScreen reports whether a rotated box can touch the visible camera span.
func boxOnScreen(cx, cy, w, h, rotation, cameraX, screenW float64) bool {
	if w <= 0 || h <= 0 {
		return false
	}
	hw := (math.Abs(math.Cos(rotation))*w + math.Abs(math.Sin(rotation))*h) / 2
	return cx+hw >= cameraX && cx-hw <= cameraX+screenW
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
	for i, l := range lines {
		ui.DrawTextOutlined(screen, l, 8, 8+i*18, 1, color.White, color.Black)
	}
}
