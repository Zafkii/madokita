package editor

func (a *EditorApp) openMovementFile(path string) error {
	proj, err := ImportMovement(path)
	if err != nil {
		return err
	}
	a.proj = *proj
	a.rebuildFromProj()
	return nil
}
