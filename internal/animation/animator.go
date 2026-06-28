package animation

type AnimPhase int

const (
	AnimNone AnimPhase = iota
	AnimWindup
	AnimActive
	AnimRecover
	AnimArmed
	AnimGuard
	AnimIdle
)

type Animator struct {
	def         Movement
	sprites     int
	frame       int
	timer       float64
	playing     bool
	loop        bool
	currentAnim string
}

func NewAnimator(def Movement, spriteCount int) *Animator {
	return &Animator{
		def:     def,
		sprites: spriteCount,
	}
}

func (a *Animator) PlayAnimation(name string) bool {
	_, ok := a.def.Animations[name]
	if !ok {
		return false
	}
	a.currentAnim = name
	a.frame = 0
	a.timer = 0
	a.playing = true
	a.loop = a.def.Animations[name].Loop
	return true
}

func (a *Animator) Update(dt float64) {
	if !a.playing {
		return
	}
	anim, ok := a.def.Animations[a.currentAnim]
	if !ok || anim.FPS <= 0 {
		return
	}
	frameDuration := 1.0 / anim.FPS
	a.timer += dt
	for a.timer >= frameDuration {
		a.timer -= frameDuration
		a.frame++
		if a.frame >= len(anim.Frames) {
			if a.loop {
				a.frame = 0
			} else {
				a.frame = len(anim.Frames) - 1
				a.playing = false
			}
		}
	}
}

func (a *Animator) Frame() int             { return a.frame }
func (a *Animator) IsPlaying() bool        { return a.playing }
func (a *Animator) Stop()                  { a.playing = false; a.frame = 0 }
func (a *Animator) CurrentAnimation() string { return a.currentAnim }

func (a *Animator) CurrentFrame() *Frame {
	anim, ok := a.def.Animations[a.currentAnim]
	if !ok {
		return nil
	}
	if a.frame < 0 || a.frame >= len(anim.Frames) {
		return nil
	}
	return &anim.Frames[a.frame]
}

func (a *Animator) currentAnimName() string {
	return a.currentAnim
}
