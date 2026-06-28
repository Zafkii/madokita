package player

type ControlManager struct {
	characters []*Player
	activeIdx  int
	isOnline   bool
	onSwitch   func(old, new_ *Player)
}

func NewControlManager(characters []*Player) *ControlManager {
	cm := &ControlManager{
		characters: characters,
		activeIdx:  0,
	}
	if len(characters) > 0 {
		characters[0].State.IsControlled = true
	}
	return cm
}

func (cm *ControlManager) Active() *Player {
	if cm.activeIdx >= 0 && cm.activeIdx < len(cm.characters) {
		return cm.characters[cm.activeIdx]
	}
	return nil
}

func (cm *ControlManager) SwitchTo(idx int) bool {
	if cm.isOnline || idx < 0 || idx >= len(cm.characters) || idx == cm.activeIdx {
		return false
	}
	old := cm.Active()
	if old != nil {
		old.State.IsControlled = false
	}
	cm.activeIdx = idx
	newChar := cm.Active()
	if newChar != nil {
		newChar.State.IsControlled = true
	}
	if cm.onSwitch != nil {
		cm.onSwitch(old, newChar)
	}
	return true
}

func (cm *ControlManager) Next() bool {
	if len(cm.characters) == 0 {
		return false
	}
	return cm.SwitchTo((cm.activeIdx + 1) % len(cm.characters))
}

func (cm *ControlManager) SetOnline(online bool) {
	cm.isOnline = online
}

func (cm *ControlManager) IsOnline() bool {
	return cm.isOnline
}

func (cm *ControlManager) SetOnSwitch(fn func(old, new_ *Player)) {
	cm.onSwitch = fn
}

func (cm *ControlManager) Party() []*Player {
	return cm.characters
}
