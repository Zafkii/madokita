package editor

import (
	"animprite/internal/filedialog"
)

func (a *EditorApp) browseSprite(idx int) {
	path, err := filedialog.OpenFile("Select PNG", "PNG Images\000*.png\000All Files\000*.*")
	if err != nil {
		a.setStatus("Browse cancelled")
		return
	}

	if err := a.loadSpriteImage(idx, path, path, true); err != nil {
		a.setStatus("Error loading PNG: " + err.Error())
		return
	}

	a.dirty = true
	a.setStatus("Loaded: " + path)
}
