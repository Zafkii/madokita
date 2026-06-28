package ecs

type ComponentID uint64

type Component interface {
	ComponentID() ComponentID
}

type ComponentManager struct {
	components map[ComponentID]map[EntityID]Component
}

func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		components: make(map[ComponentID]map[EntityID]Component),
	}
}

func (cm *ComponentManager) Add(entityID EntityID, c Component) {
	id := c.ComponentID()
	if _, ok := cm.components[id]; !ok {
		cm.components[id] = make(map[EntityID]Component)
	}
	cm.components[id][entityID] = c
}

func (cm *ComponentManager) Remove(entityID EntityID, cid ComponentID) {
	if m, ok := cm.components[cid]; ok {
		delete(m, entityID)
	}
}

func (cm *ComponentManager) Get(entityID EntityID, cid ComponentID) Component {
	if m, ok := cm.components[cid]; ok {
		return m[entityID]
	}
	return nil
}

func (cm *ComponentManager) GetAll(cid ComponentID) map[EntityID]Component {
	return cm.components[cid]
}
