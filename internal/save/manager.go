package save

import "log"

type Manager struct {
	repo        Repository
	profile     Profile
	currentSlot string
	data        Data
}

var global *Manager

func Initialize(repo Repository) error {
	p, err := repo.LoadProfile()
	if err != nil {
		p = DefaultProfile()
	}

	slots, err := repo.ListSlots()
	if err != nil || len(slots) == 0 {
		if err := repo.SaveSlot(p.LastSlot, DefaultData()); err != nil {
			return err
		}
		slots = []SlotInfo{{Name: p.LastSlot, Version: 1}}
	}

	current := p.LastSlot
	found := false
	for _, s := range slots {
		if s.Name == current {
			found = true
			break
		}
	}
	if !found {
		current = slots[0].Name
	}

	data, err := repo.LoadSlot(current)
	if err != nil {
		data = DefaultData()
	}

	global = &Manager{
		repo:        repo,
		profile:     p,
		currentSlot: current,
		data:        data,
	}

	if p.Username == "Player" && p.FavoriteChar == "madoka" {
		log.Printf("[save] profile: %s | fav: %s | slot: %s", p.Username, p.FavoriteChar, current)
	}
	return nil
}

func GetProfile() Profile {
	if global == nil {
		return DefaultProfile()
	}
	return global.profile
}

func SaveProfile(p Profile) error {
	if global == nil {
		return nil
	}
	global.profile = p
	return global.repo.SaveProfile(p)
}

func GetData() Data {
	if global == nil {
		return DefaultData()
	}
	return global.data
}

func Save() error {
	if global == nil {
		return nil
	}
	if err := global.repo.SaveSlot(global.currentSlot, global.data); err != nil {
		return err
	}
	return global.repo.SaveProfile(global.profile)
}

func Update(partial Data) {
	if global == nil {
		return
	}
	if partial.Progress.StagesUnlocked != nil {
		global.data.Progress.StagesUnlocked = partial.Progress.StagesUnlocked
	}
	if partial.Progress.CharactersUnlocked != nil {
		global.data.Progress.CharactersUnlocked = partial.Progress.CharactersUnlocked
	}
}

func ListSlots() ([]SlotInfo, error) {
	if global == nil {
		return nil, nil
	}
	return global.repo.ListSlots()
}

func SwitchSlot(slot string) error {
	if global == nil {
		return nil
	}
	data, err := global.repo.LoadSlot(slot)
	if err != nil {
		return err
	}
	global.currentSlot = slot
	global.data = data
	global.profile.LastSlot = slot
	return global.repo.SaveProfile(global.profile)
}

func DeleteSlot(slot string) error {
	if global == nil {
		return nil
	}
	return global.repo.DeleteSlot(slot)
}

func CurrentSlot() string {
	if global == nil {
		return ""
	}
	return global.currentSlot
}
