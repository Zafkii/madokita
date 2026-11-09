package editor

import (
	"animprite/internal/project"
)

func (a *EditorApp) openMovementFile(path string) error {
	proj, err := ImportMovement(path)
	if err != nil {
		return err
	}
	session := a.proj.Sprites
	a.proj = *proj
	a.proj.Sprites = mergeProjectSprites(proj.Sprites, session)
	a.rebuildFromProj()
	return nil
}

func (a *EditorApp) openAttackFile(path string) error {
	proj, err := ImportAttack(path)
	if err != nil {
		return err
	}
	session := a.proj.Sprites
	a.proj = *proj
	a.proj.Sprites = mergeProjectSprites(proj.Sprites, session)
	a.rebuildFromProj()
	return nil
}

// mergeProjectSprites keeps the imported sprite rows when the file carries
// them; session sprites only survive as a fallback for legacy files without
// a Sprites section.
func mergeProjectSprites(imported, session []project.SpriteRow) []project.SpriteRow {
	if len(imported) > 0 {
		return imported
	}
	return session
}
