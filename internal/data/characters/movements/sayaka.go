package movements

import "madokita/internal/animation"

var (
	idleHB = []animation.FrameHurtbox{
		animation.HB(95, 61, 1.5, -32.5),
		animation.HB(54, 130, 4, 62),
	}
	walkHB = []animation.FrameHurtbox{
		animation.HB(100, 57, 1, -32.5),
		animation.HB(52, 130, 1, 61),
	}
	jumpHB1 = []animation.FrameHurtbox{
		animation.HB(92, 63, 0, -34.5),
		animation.HB(52, 103, -2, 49.5),
	}
	jumpHB2 = []animation.FrameHurtbox{
		animation.HB(92, 63, 0, -34.5),
		animation.HB(52, 128, -2, 62),
	}
	jumpHB3 = []animation.FrameHurtbox{
		animation.HB(92, 63, 0, -34.5),
		animation.HB(52, 107, -2, 51.5),
	}
	burstHB = []animation.FrameHurtbox{
		animation.HB(93, 61, -3.5, -34.5),
		animation.HB(35, 131, -1.5, 59.5),
	}
)

var SayakaMovement = animation.Movement{
	AssetKey:       "sayaka_movement",
	DefaultOriginX: 0.506,
	DefaultOriginY: 0.586,
	Animations: map[string]animation.MovementAnimDef{
		"idle": {
			FPS:  3,
			Loop: true,
			Frames: []animation.Frame{
				animation.F(0, idleHB...),
				animation.F(1, idleHB...),
				animation.F(2, idleHB...),
				animation.F(3, idleHB...),
			},
		},
		"walk": {
			FPS:  10,
			Loop: true,
			Frames: []animation.Frame{
				animation.F(4, walkHB...),
				animation.F(5, walkHB...),
				animation.F(6, walkHB...),
				animation.F(7, walkHB...),
			},
		},
		"jump": {
			FPS:  7,
			Loop: false,
			Frames: []animation.Frame{
				animation.F(17, jumpHB1...),
				animation.F(18, jumpHB2...),
				animation.F(19, jumpHB2...),
				animation.F(20, jumpHB2...),
				animation.F(21, jumpHB2...),
				animation.F(22, jumpHB2...),
				animation.F(23, jumpHB3...),
			},
		},
		"dodge": {
			FPS:  1,
			Loop: false,
			Frames: []animation.Frame{
				animation.F(24,
					animation.HBR(89, 74, 5, -11, 0),
					animation.HBR(40, 102, -35.1, 69.65, 35),
				),
			},
		},
		"skill": {
			FPS:  8,
			Loop: false,
			Frames: []animation.Frame{
				animation.F(8,
					animation.HB(96, 56, 1, -34),
					animation.HB(52, 133, 8, 60.5),
				),
			},
		},
		"burst": {
			FPS:  7,
			Loop: false,
			Frames: []animation.Frame{
				animation.F(9, burstHB...),
				animation.F(10, burstHB...),
				animation.F(11, burstHB...),
				animation.F(12, burstHB...),
				animation.F(13, burstHB...),
				animation.F(14, burstHB...),
				animation.F(15, burstHB...),
				animation.F(16, burstHB...),
			},
		},
	},
}
