package editor

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
