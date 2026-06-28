package effect

type Interface interface {
	Play()
	Stop()
	Update(delta float64)
	Destroy()
}

type Manager struct {
	effects  map[string]Interface
	registry map[string]func() Interface
}

func NewManager() *Manager {
	return &Manager{
		effects:  make(map[string]Interface),
		registry: make(map[string]func() Interface),
	}
}

func (m *Manager) Register(key string, factory func() Interface) {
	m.registry[key] = factory
}

func (m *Manager) PlayBurst(characterKey string) {
	if factory, ok := m.registry[characterKey]; ok {
		eff := factory()
		eff.Play()
		m.effects[characterKey] = eff
	}
}

func (m *Manager) StopAll() {
	for id, eff := range m.effects {
		eff.Stop()
		delete(m.effects, id)
	}
}

func (m *Manager) Update(delta float64) {
	for _, eff := range m.effects {
		eff.Update(delta)
	}
}

func (m *Manager) Destroy() {
	m.StopAll()
}
