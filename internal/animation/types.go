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
