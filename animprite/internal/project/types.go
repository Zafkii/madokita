package project

type FramePhase int

const (
	PhaseWindup  FramePhase = iota
	PhaseActive
	PhaseRecover
	PhaseArmed
)

type FrameSpriteEntry struct {
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

type AnimationFrame struct {
	Sprites   []FrameSpriteEntry
	Phase     FramePhase
	Hurtboxes []HurtboxRow
}

type AnimationRow struct {
	Name       string
	CurrentIdx int
	Loop       bool
	Frames     []AnimationFrame

	Windup   float64
	Active   float64
	Recover  float64
	Armed    float64
	ArmedFPS float64

	FPS float64
}

type SpriteRow struct {
	Name       string
	File       string
	Width      int
	Height     int
	FrameCount int
	CurrentIdx int

	OffsetX  float64
	OffsetY  float64
	ScaleX   float64
	ScaleY   float64
	Rotation float64
	OriginX  float64
	OriginY  float64
}

type HurtboxRow struct {
	X          float64
	Y          float64
	Width      float64
	Height     float64
	Rotation   float64
	DmgMult    float64
}

type HitboxRow struct {
	Width  float64
	Height float64
}

type ProjectData struct {
	AssetName      string
	AssetKey       string
	DefaultOriginX float64
	DefaultOriginY float64
	Animations     []AnimationRow
	Sprites        []SpriteRow
	HitDefs        []HitboxRow
}
