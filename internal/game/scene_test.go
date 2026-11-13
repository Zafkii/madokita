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

func TestGameSceneSolidRectsFromStageBlocks(t *testing.T) {
	s := &GameScene{
		stage: &assets.StageDef{
			Layout: &stagepkg.Stage{
				Objects: []stagepkg.Object{
					{Category: stagepkg.CategoryBG, X: 10, Y: 10, Width: 20, Height: 20},
					{Category: stagepkg.CategoryStart, X: 30, Y: 30},
					{Category: stagepkg.CategoryBlock, X: 100, Y: 200, Width: 50, Height: 50},
					{Category: stagepkg.CategoryAnimBlock, X: 300, Y: 100, Width: 20, Height: 40},
					{Category: stagepkg.CategoryBlock, X: 500, Y: 100, Width: 20, Height: 20, CollisionDisabled: true},
					{Category: stagepkg.CategoryProp, X: 700, Y: 100, Width: 10, Height: 10},
					{Category: stagepkg.CategoryProp, X: 600, Y: 100, Width: 10, Height: 10, CollisionBoxes: []stagepkg.Box{{X: 5, Y: -5, Width: 20, Height: 10}}},
				},
			},
		},
	}

	rects := s.solidRects()
	if len(rects) != 3 {
		t.Fatalf("solidRects() returned %d rects, want 3: %+v", len(rects), rects)
	}
	want := []struct{ x, y, w, h float64 }{
		{75, 175, 50, 50},
		{290, 80, 20, 40},
		{595, 90, 20, 10},
	}
	for i, w := range want {
		got := rects[i]
		if got.X != w.x || got.Y != w.y || got.W != w.w || got.H != w.h {
			t.Errorf("rect[%d] = %+v, want (%v,%v,%v,%v)", i, got, w.x, w.y, w.w, w.h)
		}
	}
}

func TestGameScenePlayerLandsOnStageBlocks(t *testing.T) {
	s := &GameScene{
		stage: &assets.StageDef{
			Layout: &stagepkg.Stage{
				Width:  1280,
				Height: 720,
				Objects: []stagepkg.Object{
					{Category: stagepkg.CategoryBlock, X: 320, Y: 700, Width: 50, Height: 50},
				},
			},
			SpawnX:  320,
			SpawnY:  500,
			GroundY: 1000,
		},
	}
	p := player.New(0, 0, nil, nil, nil)
	s.SetPlayer(p)
	p.State.IsControlled = true

	for i := 0; i < 300; i++ {
		if err := s.Update(0.01); err != nil {
			t.Fatal(err)
		}
	}

	wantY := 675 - (32.0+190.0/2)*0.5
	if diff := p.Y - wantY; diff < -0.01 || diff > 0.01 {
		t.Fatalf("player Y = %v, want %v on the block top", p.Y, wantY)
	}
	if !p.Movement.IsGrounded {
		t.Fatal("player must be grounded on the stage block")
	}
}

func TestGameScenePlayerLandsOnTestStageFloor(t *testing.T) {
	s := &GameScene{stage: TestStageDef}
	p := player.New(0, 0, nil, nil, nil)
	s.SetPlayer(p)
	p.State.IsControlled = true

	for i := 0; i < 300; i++ {
		if err := s.Update(0.01); err != nil {
			t.Fatal(err)
		}
	}

	if !p.Movement.IsGrounded {
		t.Fatal("player must land on the imported test stage blocks")
	}
	if p.Y >= p.StageGroundY {
		t.Fatalf("player Y = %v, want a stage block above the fallback ground line %v", p.Y, p.StageGroundY)
	}
}
