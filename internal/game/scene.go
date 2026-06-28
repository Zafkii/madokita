package game

import (
	"image/color"
	"time"

	"madokita/internal/assets"
	"madokita/internal/audio"
	"madokita/internal/debug"
	"madokita/internal/entity/player"
	"madokita/internal/input"
	math2 "madokita/internal/math"
	"madokita/internal/scene"

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
	mgr       *scene.Manager
	assetM    *assets.AssetManager
	stage     *assets.StageDef
	audioMgr  *audio.AudioManager
	CameraX   float64
	Player    *player.Player
	ShowDebug bool
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
	if s.Player != nil {
		if s.Player.Input != nil && s.Player.Input.IsJustPressed(input.ActionToggleDebug) {
			s.ShowDebug = !s.ShowDebug
		}

		s.Player.Update(timeDurationFromFloat(dt))

		s.CameraX = s.Player.X - CanvasW/2
		if s.CameraX < 0 {
			s.CameraX = 0
		}
		if maxCX := float64(StageWidth - CanvasW); s.CameraX > maxCX {
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
	return math2.NewRect(0, GroundY, StageWidth, CanvasH-GroundY)
}

func (s *GameScene) Enter() error {
	s.audioMgr.Stop("menu-theme")
	return nil
}

func (s *GameScene) Exit() error {
	return nil
}

func (s *GameScene) Pause()         {}
func (s *GameScene) Resume()        {}
func (s *GameScene) IsActive() bool { return true }
