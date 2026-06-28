package editor

func (a *EditorApp) saveMovementFile(path string) error {
	return ExportMovement(path, &a.proj)
}
