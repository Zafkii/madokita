package engine

type System interface {
	ID() string
	Initialize() error
	Start() error
	Ready() error
	Shutdown() error
}

type SystemManager struct {
	systems []System
}

func NewSystemManager() *SystemManager {
	return &SystemManager{}
}

func (sm *SystemManager) Add(s System) {
	sm.systems = append(sm.systems, s)
}

func (sm *SystemManager) InitializeAll() error {
	for _, s := range sm.systems {
		if err := s.Initialize(); err != nil {
			return err
		}
	}
	return nil
}

func (sm *SystemManager) StartAll() error {
	for _, s := range sm.systems {
		if err := s.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (sm *SystemManager) ReadyAll() error {
	for _, s := range sm.systems {
		if err := s.Ready(); err != nil {
			return err
		}
	}
	return nil
}

func (sm *SystemManager) ShutdownAll() error {
	for i := len(sm.systems) - 1; i >= 0; i-- {
		if err := sm.systems[i].Shutdown(); err != nil {
			return err
		}
	}
	return nil
}
