package editor

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"animprite/internal/filedialog"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func (a *EditorApp) browseSprite(idx int) {
	path, err := filedialog.OpenFile("Select PNG", "PNG Images\000*.png\000All Files\000*.*")
	if err != nil {
		a.setStatus("Browse cancelled")
		return
	}

	f, err := os.Open(path)
	if err != nil {
		a.setStatus("Error opening file")
		return
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		a.setStatus("Error decoding PNG")
		return
	}

	ebImg := ebiten.NewImageFromImage(img)

	if old, ok := a.loadedSprites[idx]; ok {
		old.Deallocate()
	}
	a.loadedSprites[idx] = ebImg

	row := &a.proj.Sprites[idx]
	row.File = path

	if row.Width <= 0 {
		row.Width = 256
	}
	if row.Height <= 0 {
		row.Height = 256
	}
	fw := row.Width
	fh := row.Height
	cols := img.Bounds().Dx() / fw
	rows2 := img.Bounds().Dy() / fh
	if cols < 1 {
		cols = 1
	}
	if rows2 < 1 {
		rows2 = 1
	}
	row.FrameCount = cols * rows2
	row.CurrentIdx = 0

	if idx < len(a.spriteWidthInputs) {
		a.spriteWidthInputs[idx].SetNumeric(float64(row.Width))
		a.spriteHeightInputs[idx].SetNumeric(float64(row.Height))
	}

	ext := filepath.Ext(path)
	if ext == "" {
		ext = "img"
	}
	ext = strings.TrimPrefix(ext, ".")
	a.spriteBrowseBtns[idx].Text = ext
	a.spriteBrowseBtns[idx].BtnType = ui.BtnGreen

	a.setStatus("Loaded: " + path)
}
