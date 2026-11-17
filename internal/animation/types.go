package animation

type FrameHurtbox struct {
	W, H, OffsetX, OffsetY   float64
	ScaleX, ScaleY, Rotation float64
	DamageMultiplier         float64
}

// FrameHitbox is a damage box authored per attack frame. It carries only the
// four values the editor exposes: size and centre offset from the frame's
// pivot, in the same space as FrameHurtbox.
type FrameHitbox struct {
	W, H, OffsetX, OffsetY float64
}

type SpriteSheetDef struct {
	File       string
	FrameW     int
	FrameH     int
	FrameCount int
}

type FrameSprite struct {
	SpriteIdx      int
	SpriteFrameIdx int
	OffsetX        float64
	OffsetY        float64
	Rotation       float64
	ScaleX         float64
	ScaleY         float64
	OriginX        float64
	OriginY        float64
}

type Frame struct {
	Sprites   []FrameSprite
	Hurtboxes []FrameHurtbox
}

type MovementAnimDef struct {
	Frames []Frame
	FPS    float64
	Loop   bool
}

type Movement struct {
	AssetKey       string
	DefaultOriginX float64
	DefaultOriginY float64
	Sprites        []SpriteSheetDef
	Animations     map[string]MovementAnimDef
}

type AttackPhase string

const (
	PhaseWindup  AttackPhase = "wu"
	PhaseActive  AttackPhase = "atk"
	PhaseRecover AttackPhase = "rc"
	PhaseArmed   AttackPhase = "armed"
	PhaseGuard   AttackPhase = "guard"
)

type AttackFrame struct {
	Sprites   []FrameSprite
	Hurtboxes []FrameHurtbox
	Hitboxes  []FrameHitbox
	Phase     *AttackPhase
}

type AttackAnimDef struct {
	Frames        []AttackFrame
	FPS           float64
	Loop          bool
	Windup        float64
	ActiveTime    float64
	Recover       float64
	Armed         float64
	ArmedFPS      float64
	WindupFrames  int
	ActiveFrames  int
	RecoverFrames int
}

type Attack struct {
	AssetKey       string
	DefaultOriginX float64
	DefaultOriginY float64
	Sprites        []SpriteSheetDef
	Animations     map[string]AttackAnimDef
}

func Anim(fps float64, loop bool, frames ...Frame) MovementAnimDef {
	return MovementAnimDef{FPS: fps, Loop: loop, Frames: frames}
}

func S(spriteIdx, spriteFrameIdx int, rest ...float64) FrameSprite {
	e := FrameSprite{
		SpriteIdx:      spriteIdx,
		SpriteFrameIdx: spriteFrameIdx,
		ScaleX:         1,
		ScaleY:         1,
		OriginX:        0.5,
		OriginY:        0.5,
	}
	if len(rest) > 0 {
		e.OffsetX = rest[0]
	}
	if len(rest) > 1 {
		e.OffsetY = rest[1]
	}
	if len(rest) > 2 {
		e.Rotation = rest[2]
	}
	if len(rest) > 3 {
		e.ScaleX = rest[3]
	}
	if len(rest) > 4 {
		e.ScaleY = rest[4]
	}
	if len(rest) > 5 {
		e.OriginX = rest[5]
	}
	if len(rest) > 6 {
		e.OriginY = rest[6]
	}
	return e
}

func F(parts ...any) Frame {
	f := Frame{}
	for _, p := range parts {
		switch v := p.(type) {
		case FrameSprite:
			f.Sprites = append(f.Sprites, v)
		case FrameHurtbox:
			f.Hurtboxes = append(f.Hurtboxes, v)
		}
	}
	return f
}

// AttackF builds an attack frame from S(...) sprites, HB(...)/HBR(...)
// hurtboxes and HIT(...) hitboxes, in any order and any combination.
func AttackF(phase AttackPhase, parts ...any) AttackFrame {
	p := phase
	f := AttackFrame{Phase: &p}
	for _, part := range parts {
		switch v := part.(type) {
		case FrameSprite:
			f.Sprites = append(f.Sprites, v)
		case FrameHurtbox:
			f.Hurtboxes = append(f.Hurtboxes, v)
		case FrameHitbox:
			f.Hitboxes = append(f.Hitboxes, v)
		}
	}
	return f
}

func AttackAnim(fps float64, loop bool, windup, active, recover, armed, armedFPS float64, windupFrames, activeFrames, recoverFrames int, frames ...AttackFrame) AttackAnimDef {
	return AttackAnimDef{
		FPS: fps, Loop: loop,
		Windup: windup, ActiveTime: active, Recover: recover,
		Armed: armed, ArmedFPS: armedFPS,
		WindupFrames: windupFrames, ActiveFrames: activeFrames, RecoverFrames: recoverFrames,
		Frames: frames,
	}
}

// HB builds a hurtbox. The optional trailing values are the rotation in
// degrees and the damage multiplier, both defaulting to 0 and 1.
func HB(w, h, ox, oy float64, rest ...float64) FrameHurtbox {
	hb := FrameHurtbox{
		W: w, H: h, OffsetX: ox, OffsetY: oy,
		ScaleX: 1, ScaleY: 1, DamageMultiplier: 1,
	}
	if len(rest) > 0 {
		hb.Rotation = rest[0]
	}
	if len(rest) > 1 {
		hb.DamageMultiplier = rest[1]
	}
	return hb
}

// HBR builds a rotated hurtbox. It is the two-argument tail of HB kept as its
// own name because the movement data reads better with it.
func HBR(w, h, ox, oy, rot float64) FrameHurtbox {
	return HB(w, h, ox, oy, rot)
}

// HIT builds the hitbox of an attack frame: width, height and the centre
// offset from the frame's pivot.
func HIT(w, h, ox, oy float64) FrameHitbox {
	return FrameHitbox{W: w, H: h, OffsetX: ox, OffsetY: oy}
}
