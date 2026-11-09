package editor

import (
	"image/color"
	"strings"

	"animprite/internal/theme"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type confirmState struct {
	visible   bool
	title     string
	message   string
	onConfirm func()
}

func (a *EditorApp) markClean() {
	a.dirty = false
}

func (a *EditorApp) hasUnsavedChanges() bool {
	if a.mode == modeStage {
		// TODO: enable for stage once its save flow is settled.
		return false
	}
	return a.dirty
}

func (a *EditorApp) requestOpenFolder(dir string) {
	if a.hasUnsavedChanges() {
		a.showConfirm(
			"Unsaved Changes",
			"The current project has unsaved changes.\nOpening another folder will discard them.",
			func() { a.openFolderFromDialog(dir) },
		)
		return
	}
	a.openFolderFromDialog(dir)
}

func (a *EditorApp) requestOpenFile(path string) {
	if a.hasUnsavedChanges() {
		a.showConfirm(
			"Unsaved Changes",
			"The current project has unsaved changes.\nOpening another file will discard them.",
			func() { a.openFileFromDialog(path) },
		)
		return
	}
	a.openFileFromDialog(path)
}

func (a *EditorApp) openFolderFromDialog(dir string) {
	if err := a.openFolder(dir); err != nil {
		a.setStatus("Open directory error: " + err.Error())
		return
	}
	a.markClean()
	a.setStatus("Opened folder: " + dir)
}

func (a *EditorApp) openFileFromDialog(path string) {
	if err := a.openProjectFile(path, false); err != nil {
		a.setStatus("Open error: " + err.Error())
		return
	}
	a.markClean()
	a.setStatus("Opened: " + path)
}

func (a *EditorApp) showConfirm(title, message string, onConfirm func()) {
	a.confirm = confirmState{
		visible:   true,
		title:     title,
		message:   message,
		onConfirm: onConfirm,
	}
	a.layoutConfirmDialog()
}

func (a *EditorApp) acceptConfirm() {
	if !a.confirm.visible {
		return
	}
	action := a.confirm.onConfirm
	a.confirm = confirmState{}
	if action != nil {
		action()
	}
}

func (a *EditorApp) cancelConfirm() {
	a.confirm = confirmState{}
}

func (a *EditorApp) confirmDialogRect() (x, y, w, h int) {
	w, h = 460, 158
	if w > a.win.outsideWidth-40 {
		w = a.win.outsideWidth - 40
	}
	if h > a.win.outsideHeight-40 {
		h = a.win.outsideHeight - 40
	}
	x = (a.win.outsideWidth - w) / 2
	y = (a.win.outsideHeight - h) / 2
	return
}

func (a *EditorApp) layoutConfirmDialog() {
	if a.confirmYesBtn == nil || a.confirmNoBtn == nil {
		return
	}
	x, y, w, h := a.confirmDialogRect()
	btnW, btnH := 130, 26
	gap := 10
	total := btnW*2 + gap
	bx := x + (w-total)/2
	by := y + h - btnH - 22
	a.confirmYesBtn.X, a.confirmYesBtn.Y = bx, by
	a.confirmYesBtn.W, a.confirmYesBtn.H = btnW, btnH
	a.confirmNoBtn.X, a.confirmNoBtn.Y = bx+btnW+gap, by
	a.confirmNoBtn.W, a.confirmNoBtn.H = btnW, btnH
}

func (a *EditorApp) updateConfirmDialog(mx, my int, justL bool) {
	if !a.confirm.visible {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.cancelConfirm()
		return
	}
	a.layoutConfirmDialog()
	a.confirmNoBtn.HandleMouse(mx, my, justL)
	if !a.confirm.visible {
		return
	}
	a.confirmYesBtn.HandleMouse(mx, my, justL)
}

func (a *EditorApp) drawConfirmDialog(screen *ebiten.Image, p theme.Palette) {
	if !a.confirm.visible {
		return
	}
	ui.FillRect(screen, 0, 0, a.win.outsideWidth, a.win.outsideHeight, color.RGBA{0, 0, 0, 160})

	x, y, w, h := a.confirmDialogRect()
	ui.FillRect(screen, x, y, w, h, p.Border)
	ui.FillRect(screen, x+1, y+1, w-2, h-2, p.RightPanelBG)

	ui.DrawText(screen, a.confirm.title, x+20, y+18, 1, p.TextPrimary)
	lines := strings.Split(a.confirm.message, "\n")
	for i, line := range lines {
		ui.DrawText(screen, line, x+20, y+48+i*18, 0.9, p.TextMuted)
	}

	a.confirmYesBtn.Draw(screen)
	a.confirmNoBtn.Draw(screen)
}
