package engine

type Phase int

const (
	PhasePreInit Phase = iota
	PhaseInit
	PhasePostInit
	PhaseReady
	PhaseRunning
	PhaseShutdown
)

type Runtime struct {
	phase     Phase
	systems   *SystemManager
	container *Container
}

func NewRuntime() *Runtime {
	return &Runtime{
		phase:     PhasePreInit,
		systems:   NewSystemManager(),
		container: NewContainer(),
	}
}

func (r *Runtime) Initialize() error {
	r.phase = PhaseInit
	if err := r.systems.InitializeAll(); err != nil {
		return err
	}
	r.phase = PhasePostInit
	if err := r.systems.StartAll(); err != nil {
		return err
	}
	r.phase = PhaseReady
	if err := r.systems.ReadyAll(); err != nil {
		return err
	}
	r.phase = PhaseRunning
	return nil
}

func (r *Runtime) Shutdown() error {
	r.phase = PhaseShutdown
	return r.systems.ShutdownAll()
}

func (r *Runtime) Phase() Phase {
	return r.phase
}

func (r *Runtime) Container() *Container {
	return r.container
}

func (r *Runtime) Systems() *SystemManager {
	return r.systems
}
