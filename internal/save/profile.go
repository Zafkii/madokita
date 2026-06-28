package save

type Profile struct {
	Username     string `json:"username"`
	FavoriteChar string `json:"favoriteChar"`
	LastSlot     string `json:"lastSlot"`
}

func DefaultProfile() Profile {
	return Profile{
		Username:     "Player",
		FavoriteChar: "madoka",
		LastSlot:     "Avance 1",
	}
}
