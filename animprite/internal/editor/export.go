package editor

func (a *EditorApp) saveMovementFile(path string) error {
	return ExportMovement(path, &a.proj)
}

func (a *EditorApp) saveAttackFile(path string) error {
	return ExportAttack(path, &a.proj)
}
