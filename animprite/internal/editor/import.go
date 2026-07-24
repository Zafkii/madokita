package editor

import (
	"bytes"
	"fmt"
	"os"
)

func (a *EditorApp) openFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if bytes.Contains(data, []byte("= Attack{")) {
		if err := a.openAttackFile(path); err != nil {
			return err
		}
		a.currentFilePath = path
		a.mode = modeAttack
		a.modeDropdown.Selected = 1
		a.modeDropdown.DisplayText = "Attack Editor"
		return nil
	}

	if bytes.Contains(data, []byte("= Movement{")) {
		if err := a.openMovementFile(path); err != nil {
			return err
		}
		a.currentFilePath = path
		a.mode = modeMovement
		a.modeDropdown.Selected = 0
		a.modeDropdown.DisplayText = "Movement Editor"
		return nil
	}

	return fmt.Errorf("unknown file type: neither Movement nor Attack declaration found")
}

func (a *EditorApp) openMovementFile(path string) error {
	proj, err := ImportMovement(path)
	if err != nil {
		return err
	}
	savedSprites := a.proj.Sprites
	a.proj = *proj
	a.proj.Sprites = savedSprites
	a.rebuildFromProj()
	return nil
}

func (a *EditorApp) openAttackFile(path string) error {
	proj, err := ImportAttack(path)
	if err != nil {
		return err
	}
	savedSprites := a.proj.Sprites
	a.proj = *proj
	a.proj.Sprites = savedSprites
	a.rebuildFromProj()
	return nil
}
