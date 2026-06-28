package localization

import "strings"

type Manager struct {
	translations map[string]string
}

var global *Manager

func Initialize(lang string) error {
	global = &Manager{
		translations: map[string]string{
			"MENU.TOUCH_TO_START":              "Touch to Start",
			"MENU.NEW_GAME":                    "New Game",
			"MENU.CONTINUE":                    "Continue",
			"MENU.SETTINGS":                    "Settings",
			"MENU.EXIT":                        "Exit",
			"MENU.SETTINGS.DISPLAY.RESOLUTION": "Resolution",
			"MENU.SETTINGS.DISPLAY.FULLSCREEN": "Fullscreen",
			"MENU.SETTINGS.LANGUAGE":          "Language",
			"MENU.SETTINGS.FPS_LIMIT":         "FPS Limit",
			"MENU.BACK":                       "Back",
			"MENU.RESET":                      "Reset",
		},
	}
	return nil
}

func Get(path string) string {
	if global == nil {
		return path
	}
	if val, ok := global.translations[path]; ok {
		return val
	}
	parts := strings.Split(path, ".")
	return parts[len(parts)-1]
}
