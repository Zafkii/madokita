package game

import (
	"image/color"
	"math"
	"time"

	"madokita-gems/internal/assets"
	"madokita-gems/internal/audio"
	"madokita-gems/internal/debug"
	"madokita-gems/internal/entity/player"
	"madokita-gems/internal/input"
	math2 "madokita-gems/internal/math"
	"madokita-gems/internal/scene"
	"madokita-gems/internal/stage"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

const (
	GroundY     = 700
	CanvasW     = 1280
	CanvasH     = 720
	StageWidth  = 2560
	StageHeight = 720
)

type GameScene struct {
	mgr            *scene.Manager
	assetM         *assets.AssetManager
	stage          *assets.StageDef
	audioMgr       *audio.AudioManager
	CameraX        float64
	Player         *player.Player
	ShowDebug      bool
	stageAnimClock float64
}

func NewGameScene(mgr *scene.Manager, am *assets.AssetManager, stage *assets.StageDef, audioMgr *audio.AudioManager) *GameScene {
	return &GameScene{
		mgr:      mgr,
		assetM:   am,
		stage:    stage,
		audioMgr: audioMgr,
	}
}

func (s *GameScene) SetPlayer(p *player.Player) {
	s.Player = p
	if s.stage != nil {
		p.Spawn(s.stage.SpawnX, s.stage.GroundY)
	}
}

func (s *GameScene) Update(dt float64) error {
	s.stageAnimClock += dt
	if s.Player != nil {
		if s.Player.Input != nil && s.Player.Input.IsJustPressed(input.ActionToggleDebug) {
			s.ShowDebug = !s.ShowDebug
		}

		s.Player.Update(timeDurationFromFloat(dt))

		s.CameraX = s.Player.X - CanvasW/2
		if s.CameraX < 0 {
			s.CameraX = 0
		}
		maxCX := s.stageWidth() - CanvasW
		if maxCX < 0 {
			maxCX = 0
		}
		if s.CameraX > maxCX {
			s.CameraX = maxCX
		}
	}
	return nil
}

func timeDurationFromFloat(dt float64) time.Duration {
	return time.Duration(dt * float64(time.Second))
}

func (s *GameScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	layers := s.assetM.StageLayers(s.stage.ID)
	for _, item := range layers {
		switch item.Entry.Group {
		case assets.LayerSky:
			op := &colorm.DrawImageOptions{}
			cm := s.assetM.Tint().ForGroup(item.Entry.Group)
			colorm.DrawImage(screen, item.Image, cm, op)

		case assets.LayerFloor:
			cm := s.assetM.Tint().ForGroup(item.Entry.Group)
			op := &colorm.DrawImageOptions{}
			op.GeoM.Translate(-s.CameraX, GroundY)
			colorm.DrawImage(screen, item.Image, cm, op)
		}
	}

	if s.stage != nil && s.stage.Layout != nil {
		s.drawStageLayout(screen)
	}

	if s.Player != nil {
		s.Player.Draw(screen, s.CameraX)

		if s.ShowDebug {
			debug.DrawGround(screen, s.CameraX)
			debug.DrawOrigin(screen, s.Player, s.CameraX)
			debug.DrawSpriteBBox(screen, s.Player, s.CameraX)
			debug.DrawHurtboxes(screen, s.Player, s.CameraX)
		}
	}
}

func (s *GameScene) DrawOverlay(screen *ebiten.Image) {
	if s.ShowDebug {
		debug.DrawInfo(screen, s.Player)
	}
}

func (s *GameScene) FloorRect() math2.Rect {
	return math2.NewRect(0, GroundY, s.stageWidth(), CanvasH-GroundY)
}

func (s *GameScene) stageWidth() float64 {
	if s.stage != nil && s.stage.Layout != nil && s.stage.Layout.Width > 0 {
		return s.stage.Layout.Width
	}
	return StageWidth
}

func (s *GameScene) drawStageLayout(screen *ebiten.Image) {
	layout := s.stage.Layout
	for _, layer := range layout.Layers {
		for _, category := range []string{"bg", "prop", "anim_prop", "block", "anim_block"} {
			for i := range layout.Objects {
				obj := &layout.Objects[i]
				if obj.Layer != layer.Name || normalizedStageCategory(obj.Category) != category {
					continue
				}
				img := s.assetM.StageObjectImage(s.stage.ID, obj.File)
				if anim := s.assetM.StageObjectAnimation(s.stage.ID, obj); anim != nil {
					if frame := anim.Frame(s.stageAnimClock); frame != nil {
						img = frame
					}
				}
				if img == nil {
					continue
				}
				s.drawStageObject(screen, img, obj)
			}
		}
	}
}

func (s *GameScene) drawStageObject(screen, img *ebiten.Image, obj *stage.Object) {
	bounds := img.Bounds()
	iw := float64(bounds.Dx())
	ih := float64(bounds.Dy())
	if iw <= 0 || ih <= 0 {
		return
	}
	w := obj.Width
	if w <= 0 {
		w = iw
	}
	h := obj.Height
	if h <= 0 {
		h = ih
	}
	opacity := obj.Opacity
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}

	op := &colorm.DrawImageOptions{}
	op.GeoM.Translate(-iw/2, -ih/2)
	op.GeoM.Scale(w/iw, h/ih)
	op.GeoM.Rotate(obj.Rotation * math.Pi / 180)
	op.GeoM.Translate(obj.X-s.CameraX, obj.Y)
	op.ColorScale.ScaleAlpha(float32(opacity))
	colorm.DrawImage(screen, img, s.assetM.Tint().ForGroup(stageObjectGroup(obj.Category)), op)
}

func normalizedStageCategory(category string) string {
	switch category {
	case "bg", "prop", "anim_prop", "block", "anim_block":
		return category
	default:
		return "block"
	}
}

func stageObjectGroup(category string) assets.LayerGroup {
	switch normalizedStageCategory(category) {
	case "bg":
		return assets.LayerSky
	case "prop", "anim_prop":
		return assets.LayerStructNear
	default:
		return assets.LayerFloor
	}
}

func (s *GameScene) Enter() error {
	s.stageAnimClock = 0
	s.audioMgr.Stop("menu-theme")
	return nil
}

func (s *GameScene) Exit() error {
	if s.Player != nil {
		if s.Player.AnimDef != nil {
			s.assetM.UnloadCharacter(s.Player.AnimDef.AssetKey)
		}
		if s.Player.AttackDef != nil {
			s.assetM.UnloadCharacter(s.Player.AttackDef.AssetKey)
		}
	}
	if s.stage != nil {
		s.assetM.UnloadStage(s.stage.ID)
	}
	return nil
}

func (s *GameScene) Pause()         {}
func (s *GameScene) Resume()        {}
func (s *GameScene) IsActive() bool { return true }
