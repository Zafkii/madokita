package animation

type FrameHurtbox struct {
	W, H, OffsetX, OffsetY     float64
	ScaleX, ScaleY, Rotation   float64
	DamageMultiplier           float64
}

type Frame struct {
	SpriteFrames []int
	OffsetX      []float64
	OffsetY      []float64
	Rotation     []float64
	ScaleX       []float64
	ScaleY       []float64
	Hurtboxes    []FrameHurtbox
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
	SpriteFrames []int
	OffsetX      []float64
	OffsetY      []float64
	Rotation     []float64
	ScaleX       []float64
	ScaleY       []float64
	Phase        *AttackPhase
}

type AttackAnimDef struct {
	Frames        []AttackFrame
	FPS           float64
	Loop          bool
	Windup        float64
	ActiveTime    float64
	Recover       float64
	WindupFrames  int
	ActiveFrames  int
	RecoverFrames int
}

type Attack struct {
	AssetKey       string
	DefaultOriginX float64
	DefaultOriginY float64
	Animations     map[string]AttackAnimDef
}

func F(spriteFrame int, hurtboxes ...FrameHurtbox) Frame {
	return Frame{
		SpriteFrames: []int{spriteFrame},
		OffsetX:      []float64{0},
		OffsetY:      []float64{0},
		Rotation:     []float64{0},
		Hurtboxes:    hurtboxes,
	}
}

func HB(w, h, ox, oy float64) FrameHurtbox {
	return FrameHurtbox{
		W: w, H: h, OffsetX: ox, OffsetY: oy,
		ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1,
	}
}

func HBR(w, h, ox, oy, rot float64) FrameHurtbox {
	return FrameHurtbox{
		W: w, H: h, OffsetX: ox, OffsetY: oy,
		ScaleX: 1, ScaleY: 1, Rotation: rot, DamageMultiplier: 1,
	}
}
