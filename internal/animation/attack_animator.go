package animation

// AttackAnimator plays an AttackAnimDef frame by frame. Unlike Animator
// (which is time-driven per animation), attacks are frame-driven: frames
// advance sequentially at the def FPS and the current phase is derived from
// the phase tag on the current frame. The armed phase is special: its frames
// loop at ArmedFPS for the Armed duration, then the attack is done. The def
// Loop flag is ignored (only the armed phase loops).
type AttackAnimator struct {
	def      Attack
	animName string
	frames   []AttackFrame
	fps      float64

	frame   int
	timer   float64
	playing bool
	done    bool

	inArmed      bool
	armedFrames  []AttackFrame
	armedElapsed float64
	armedTotal   float64
	armedFPS     float64
}

func NewAttackAnimator(def Attack) *AttackAnimator {
	return &AttackAnimator{def: def}
}

func (a *AttackAnimator) PlayAnimation(name string) bool {
	anim, ok := a.def.Animations[name]
	if !ok {
		return false
	}
	a.animName = name
	a.frames = anim.Frames
	a.fps = anim.FPS
	a.frame = 0
	a.timer = 0
	a.playing = true
	a.done = false
	a.inArmed = false
	a.armedElapsed = 0
	a.armedTotal = anim.Armed / 1000.0
	a.armedFPS = anim.ArmedFPS
	a.armedFrames = nil
	for i := range anim.Frames {
		f := &anim.Frames[i]
		if f.Phase != nil && *f.Phase == PhaseArmed {
			a.armedFrames = append(a.armedFrames, *f)
		}
	}
	if len(a.frames) > 0 && a.isArmedFrame(0) {
		a.enterArmed()
	}
	return true
}

func (a *AttackAnimator) Update(dt float64) {
	if !a.playing || a.done {
		return
	}
	if a.inArmed {
		a.updateArmed(dt)
		return
	}
	if a.fps <= 0 || len(a.frames) == 0 {
		a.finish()
		return
	}
	frameDuration := 1.0 / a.fps
	a.timer += dt
	for a.timer >= frameDuration {
		a.timer -= frameDuration
		a.frame++
		if a.frame >= len(a.frames) {
			a.frame = len(a.frames) - 1
			a.finish()
			return
		}
		if a.isArmedFrame(a.frame) {
			a.enterArmed()
			return
		}
	}
}

func (a *AttackAnimator) updateArmed(dt float64) {
	a.armedElapsed += dt
	if a.armedTotal > 0 && a.armedElapsed >= a.armedTotal {
		a.finish()
		return
	}
	if len(a.armedFrames) > 0 && a.armedFPS > 0 {
		a.frame = int(a.armedElapsed*a.armedFPS) % len(a.armedFrames)
	}
}

func (a *AttackAnimator) enterArmed() {
	a.inArmed = true
	a.armedElapsed = 0
	a.frame = 0
}

func (a *AttackAnimator) isArmedFrame(idx int) bool {
	if idx < 0 || idx >= len(a.frames) {
		return false
	}
	f := &a.frames[idx]
	return f.Phase != nil && *f.Phase == PhaseArmed
}

func (a *AttackAnimator) finish() {
	a.done = true
	a.playing = false
}

// Phase returns the phase of the current frame. Windup is the default when
// no phase tag is present.
func (a *AttackAnimator) Phase() AttackPhase {
	if !a.playing || a.done {
		return PhaseWindup
	}
	if a.inArmed {
		return PhaseArmed
	}
	if a.frame < 0 || a.frame >= len(a.frames) {
		return PhaseWindup
	}
	f := &a.frames[a.frame]
	if f.Phase == nil {
		return PhaseWindup
	}
	return *f.Phase
}

func (a *AttackAnimator) CurrentFrame() *AttackFrame {
	if !a.playing || a.done {
		return nil
	}
	if a.inArmed {
		if len(a.armedFrames) == 0 {
			return nil
		}
		return &a.armedFrames[a.frame%len(a.armedFrames)]
	}
	if a.frame < 0 || a.frame >= len(a.frames) {
		return nil
	}
	return &a.frames[a.frame]
}

func (a *AttackAnimator) Frame() int                 { return a.frame }
func (a *AttackAnimator) IsPlaying() bool            { return a.playing }
func (a *AttackAnimator) Done() bool                 { return a.done }
func (a *AttackAnimator) Stop()                      { a.playing = false; a.done = true }
func (a *AttackAnimator) CurrentAnimation() string   { return a.animName }
