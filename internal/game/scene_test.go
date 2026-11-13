package game

import (
	"testing"

	"madokita-gems/internal/assets"
	"madokita-gems/internal/entity/player"
	stagepkg "madokita-gems/internal/stage"
)

func TestGameSceneStageBounds(t *testing.T) {
	s := &GameScene{
		stage: &assets.StageDef{
			Layout: &stagepkg.Stage{Width: 4000, Height: 900},
		},
	}

	if got := s.stageWidth(); got != 4000 {
		t.Fatalf("stageWidth() = %v, want 4000", got)
	}
	if got := s.stageHeight(); got != 900 {
		t.Fatalf("stageHeight() = %v, want 900", got)
	}
	floor := s.FloorRect()
	if floor.W != 4000 || floor.H != 200 {
		t.Fatalf("FloorRect() = %vx%v, want 4000x200", floor.W, floor.H)
	}
}

func TestGameSceneStageBoundsFallback(t *testing.T) {
	s := &GameScene{}

	if got := s.stageWidth(); got != StageWidth {
		t.Fatalf("stageWidth() = %v, want %v", got, StageWidth)
	}
	if got := s.stageHeight(); got != StageHeight {
		t.Fatalf("stageHeight() = %v, want %v", got, StageHeight)
	}
}

func TestGameSceneAppliesStartMarkerSpawnY(t *testing.T) {
	s := &GameScene{
		stage: &assets.StageDef{
			Layout:  &stagepkg.Stage{Width: 4000, Height: 900},
			SpawnX:  320,
			SpawnY:  480,
			GroundY: 700,
		},
	}
	p := &player.Player{}
	s.SetPlayer(p)

	if p.X != 320 || p.Y != 480 {
		t.Fatalf("player spawn = (%v, %v), want (320, 480)", p.X, p.Y)
	}
	if p.StageGroundY != 700 {
		t.Fatalf("player ground line = %v, want 700", p.StageGroundY)
	}
	if p.Movement.IsGrounded {
		t.Fatal("player spawned above the ground line must start airborne")
	}
}

func TestGameSceneSpawnYFallsBackToGround(t *testing.T) {
	s := &GameScene{
		stage: &assets.StageDef{
			Layout:  &stagepkg.Stage{Width: 4000, Height: 900},
			SpawnX:  100,
			GroundY: 700,
		},
	}
	p := &player.Player{}
	s.SetPlayer(p)

	if p.X != 100 || p.Y != 700 {
		t.Fatalf("player spawn = (%v, %v), want (100, 700)", p.X, p.Y)
	}
	if !p.Movement.IsGrounded {
		t.Fatal("player spawned on the ground line must start grounded")
	}
}

func TestGameSceneAppliesStageBoundsToPlayer(t *testing.T) {
	s := &GameScene{
		stage: &assets.StageDef{
			Layout:  &stagepkg.Stage{Width: 4000, Height: 900},
			SpawnX:  100,
			GroundY: 700,
		},
	}
	p := &player.Player{}
	s.SetPlayer(p)

	if p.WorldWidth != 4000 || p.WorldHeight != 900 {
		t.Fatalf("player world bounds = %vx%v, want 4000x900", p.WorldWidth, p.WorldHeight)
	}
	if p.X != 100 || p.Y != 700 {
		t.Fatalf("player spawn = (%v, %v), want (100, 700)", p.X, p.Y)
	}
}
