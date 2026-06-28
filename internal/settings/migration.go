package settings

import "os"

func TryMigrateFromFile() {
	if _, err := os.Stat("settings.json"); os.IsNotExist(err) {
		return
	}
	if HasSettings() {
		return
	}
	if err := MigrateFromFile("settings.json"); err != nil {
		return
	}
	os.Remove("settings.json")
}
