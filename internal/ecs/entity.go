package ecs

type EntityID uint64

type Entity struct {
	ID     EntityID
	World  *World
	active bool
}

func (e *Entity) IsActive() bool {
	return e.active
}

func (e *Entity) Destroy() {
	e.active = false
}
