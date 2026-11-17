package animation

import "testing"

func TestAttackFCollectsEveryPartKind(t *testing.T) {
	frame := AttackF(PhaseActive,
		S(0, 1),
		HB(95, 61, 1.5, -32.5),
		HIT(40, 30, 60, -20),
		HBR(20, 20, 0, 0, 30),
	)

	if frame.Phase == nil || *frame.Phase != PhaseActive {
		t.Fatalf("phase = %v, want active", frame.Phase)
	}
	if len(frame.Sprites) != 1 {
		t.Fatalf("sprites = %d, want 1", len(frame.Sprites))
	}
	if len(frame.Hurtboxes) != 2 {
		t.Fatalf("hurtboxes = %d, want 2", len(frame.Hurtboxes))
	}
	if got, want := frame.Hitboxes, []FrameHitbox{{W: 40, H: 30, OffsetX: 60, OffsetY: -20}}; len(got) != 1 || got[0] != want[0] {
		t.Fatalf("hitboxes = %+v, want %+v", got, want)
	}
	if frame.Hurtboxes[0].DamageMultiplier != 1 {
		t.Fatalf("plain HB damage multiplier = %v, want 1", frame.Hurtboxes[0].DamageMultiplier)
	}
	if frame.Hurtboxes[1].Rotation != 30 {
		t.Fatalf("HBR rotation = %v, want 30", frame.Hurtboxes[1].Rotation)
	}
}

func TestHBOptionalTail(t *testing.T) {
	rotated := HB(10, 10, 0, 0, 45)
	if rotated.Rotation != 45 || rotated.DamageMultiplier != 1 {
		t.Fatalf("HB with rotation = %+v", rotated)
	}
	weak := HB(10, 10, 0, 0, 0, 0.5)
	if weak.Rotation != 0 || weak.DamageMultiplier != 0.5 {
		t.Fatalf("HB with damage multiplier = %+v", weak)
	}
}
