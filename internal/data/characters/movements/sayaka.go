package movements

import "madokita/internal/animation"

var SayakaMovement = animation.Movement{
	AssetKey:       "sayaka_movement",
	DefaultOriginX: 0.506,
	DefaultOriginY: 0.586,
	Animations: map[string]animation.MovementAnimDef{
		"idle": {
			FPS:  3,
			Loop: true,
			Frames: []animation.Frame{
				{
					SpriteFrames: []int{0},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 95, H: 61, OffsetX: 1.5, OffsetY: -32.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 54, H: 130, OffsetX: 4, OffsetY: 62, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{1},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 95, H: 61, OffsetX: 1.5, OffsetY: -32.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 54, H: 130, OffsetX: 4, OffsetY: 62, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{2},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 95, H: 61, OffsetX: 1.5, OffsetY: -32.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 54, H: 130, OffsetX: 4, OffsetY: 62, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{3},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 95, H: 61, OffsetX: 1.5, OffsetY: -32.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 54, H: 130, OffsetX: 4, OffsetY: 62, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
			},
		},
		"walk": {
			FPS:  10,
			Loop: true,
			Frames: []animation.Frame{
				{
					SpriteFrames: []int{4},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 100, H: 57, OffsetX: 1, OffsetY: -32.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 130, OffsetX: 1, OffsetY: 61, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{5},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 100, H: 57, OffsetX: 1, OffsetY: -32.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 130, OffsetX: 1, OffsetY: 61, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{6},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 100, H: 57, OffsetX: 1, OffsetY: -32.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 130, OffsetX: 1, OffsetY: 61, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{7},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 100, H: 57, OffsetX: 1, OffsetY: -32.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 130, OffsetX: 1, OffsetY: 61, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
			},
		},
		"jump": {
			FPS:  7,
			Loop: false,
			Frames: []animation.Frame{
				{
					SpriteFrames: []int{17},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 92, H: 63, OffsetX: 0, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 103, OffsetX: -2, OffsetY: 49.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{18},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 92, H: 63, OffsetX: 0, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 128, OffsetX: -2, OffsetY: 62, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{19},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 92, H: 63, OffsetX: 0, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 128, OffsetX: -2, OffsetY: 62, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{20},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 92, H: 63, OffsetX: 0, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 128, OffsetX: -2, OffsetY: 62, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{21},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 92, H: 63, OffsetX: 0, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 128, OffsetX: -2, OffsetY: 62, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{22},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 92, H: 63, OffsetX: 0, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 128, OffsetX: -2, OffsetY: 62, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{23},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 92, H: 63, OffsetX: 0, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 107, OffsetX: -2, OffsetY: 51.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
			},
		},
		"dodge": {
			FPS:  1,
			Loop: false,
			Frames: []animation.Frame{
				{
					SpriteFrames: []int{24},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 89, H: 74, OffsetX: 5, OffsetY: -11, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 40, H: 102, OffsetX: -35.1, OffsetY: 69.65, ScaleX: 1, ScaleY: 1, Rotation: 35, DamageMultiplier: 1},
					},
				},
			},
		},
		"skill": {
			FPS:  8,
			Loop: false,
			Frames: []animation.Frame{
				{
					SpriteFrames: []int{8},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 96, H: 56, OffsetX: 1, OffsetY: -34, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 52, H: 133, OffsetX: 8, OffsetY: 60.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
			},
		},
		"burst": {
			FPS:  7,
			Loop: false,
			Frames: []animation.Frame{
				{
					SpriteFrames: []int{9},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 93, H: 61, OffsetX: -3.5, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 35, H: 131, OffsetX: -1.5, OffsetY: 59.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{10},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 93, H: 61, OffsetX: -3.5, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 35, H: 131, OffsetX: -1.5, OffsetY: 59.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{11},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 93, H: 61, OffsetX: -3.5, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 35, H: 131, OffsetX: -1.5, OffsetY: 59.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{12},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 93, H: 61, OffsetX: -3.5, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 35, H: 131, OffsetX: -1.5, OffsetY: 59.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{13},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 93, H: 61, OffsetX: -3.5, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 35, H: 131, OffsetX: -1.5, OffsetY: 59.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{14},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 93, H: 61, OffsetX: -3.5, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 35, H: 131, OffsetX: -1.5, OffsetY: 59.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{15},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 93, H: 61, OffsetX: -3.5, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 35, H: 131, OffsetX: -1.5, OffsetY: 59.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
				{
					SpriteFrames: []int{16},
					OffsetX:      []float64{0},
					OffsetY:      []float64{0},
					Rotation:     []float64{0},
					Hurtboxes: []animation.FrameHurtbox{
						{W: 93, H: 61, OffsetX: -3.5, OffsetY: -34.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
						{W: 35, H: 131, OffsetX: -1.5, OffsetY: 59.5, ScaleX: 1, ScaleY: 1, Rotation: 0, DamageMultiplier: 1},
					},
				},
			},
		},
	},
}
