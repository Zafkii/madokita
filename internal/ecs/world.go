package ecs

type System interface {
	Update(dt float64) error
}

type World struct {
	entities         []*Entity
	nextID           EntityID
	componentManager *ComponentManager
	systems          []System
}

func NewWorld() *World {
	return &World{
		componentManager: NewComponentManager(),
	}
}

func (w *World) NewEntity() *Entity {
	e := &Entity{
		ID:     w.nextID,
		World:  w,
		active: true,
	}
	w.nextID++
	w.entities = append(w.entities, e)
	return e
}

func (w *World) AddSystem(s System) {
	w.systems = append(w.systems, s)
}

func (w *World) Update(dt float64) error {
	for _, s := range w.systems {
		if err := s.Update(dt); err != nil {
			return err
		}
	}
	return nil
}

func (w *World) Components() *ComponentManager {
	return w.componentManager
}

func (w *World) ActiveEntities() []*Entity {
	var active []*Entity
	for _, e := range w.entities {
		if e.active {
			active = append(active, e)
		}
	}
	return active
}

func (w *World) DestroyEntity(id EntityID) {
	for _, e := range w.entities {
		if e.ID == id {
			e.Destroy()
			return
		}
	}
}
