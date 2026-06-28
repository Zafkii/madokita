package save

type SlotInfo struct {
	Name      string `json:"name"`
	Version   int    `json:"version"`
	UpdatedAt string `json:"updatedAt"`
}

type Data struct {
	Version    int                          `json:"version"`
	Progress   ProgressData                 `json:"progress"`
	Player     PlayerData                   `json:"player"`
	Characters map[string]CharacterSaveData `json:"characters"`
}

type ProgressData struct {
	StagesUnlocked     []string `json:"stagesUnlocked"`
	CharactersUnlocked []string `json:"charactersUnlocked"`
	EndingsUnlocked    []string `json:"endingsUnlocked"`
}

type PlayerData struct {
	Upgrades       map[string]int `json:"upgrades"`
	UnlockedSkills []string       `json:"unlockedSkills"`
}

type CharacterSaveData struct {
	UnlockedSkills []string `json:"unlockedSkills"`
}

type Repository interface {
	LoadProfile() (Profile, error)
	SaveProfile(p Profile) error
	ListSlots() ([]SlotInfo, error)
	LoadSlot(slot string) (Data, error)
	SaveSlot(slot string, data Data) error
	DeleteSlot(slot string) error
}

func DefaultData() Data {
	return Data{
		Version: 1,
		Progress: ProgressData{
			StagesUnlocked:     []string{"stage1"},
			CharactersUnlocked: []string{"madoka"},
		},
		Player: PlayerData{
			Upgrades: make(map[string]int),
		},
		Characters: map[string]CharacterSaveData{
			"madoka": {},
		},
	}
}
