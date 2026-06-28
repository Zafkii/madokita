package editor

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

type windowPrefs struct {
	WindowX int `json:"windowX"`
	WindowY int `json:"windowY"`
	WindowW int `json:"windowW"`
	WindowH int `json:"windowH"`
}

func prefsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "madokita", "animprite.json"), nil
}

func LoadWindowPrefs() windowPrefs {
	path, err := prefsPath()
	if err != nil {
		return windowPrefs{}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return windowPrefs{}
	}
	var p windowPrefs
	json.Unmarshal(raw, &p)
	return p
}

func saveWindowPrefs(p windowPrefs) {
	path, err := prefsPath()
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(path), 0755)
	raw, _ := json.MarshalIndent(p, "", "  ")
	os.WriteFile(path, raw, 0644)
}

func (a *EditorApp) saveCurrentPrefs() {
	wx, wy := ebiten.WindowPosition()
	ww, wh := ebiten.WindowSize()
	saveWindowPrefs(windowPrefs{
		WindowX: wx,
		WindowY: wy,
		WindowW: ww,
		WindowH: wh,
	})
}
