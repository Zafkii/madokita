package editor

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"animprite/internal/project"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

// spriteAssignment describes one PNG to load into the project. RowIdx is the
// existing sprite row it maps to, or -1 when the PNG must be appended as a new
// row.
type spriteAssignment struct {
	RowIdx int
	Name   string
	Path   string
	File   string
}

func modeFromFolderName(name string) (editorMode, bool) {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, "_atk"):
		return modeAttack, true
	case strings.HasSuffix(lower, "_moves"), strings.HasSuffix(lower, "_move"):
		return modeMovement, true
	case strings.HasSuffix(lower, "stage"):
		return modeStage, true
	default:
		return modeMovement, false
	}
}

func findProjectGo(dir string, mode editorMode) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".go") {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(files)

	if len(files) == 0 {
		return "", nil
	}
	if len(files) == 1 {
		return files[0], nil
	}

	preferred := ""
	switch mode {
	case modeAttack:
		preferred = "atk.go"
	case modeMovement:
		preferred = "moves.go"
	case modeStage:
		preferred = "stage.go"
	}
	if preferred != "" {
		for _, file := range files {
			if strings.EqualFold(filepath.Base(file), preferred) {
				return file, nil
			}
		}
	}

	markers := []string{"= Attack{", "= Movement{"}
	switch mode {
	case modeAttack:
		markers = []string{"= Attack{"}
	case modeMovement:
		markers = []string{"= Movement{"}
	case modeStage:
		markers = []string{"= Stage{", "= StageDef{", "= stage.Stage{", "= stage.StageDef{"}
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		for _, marker := range markers {
			if bytes.Contains(data, []byte(marker)) {
				return file, nil
			}
		}
	}

	return "", fmt.Errorf("multiple .go files in %s and none matches the project mode", dir)
}

// planFolderSprites resolves which PNG in dir belongs to each existing sprite
// row and optionally collects the remaining PNGs as new rows. It does not touch
// Ebiten, so it is safe to use from unit tests.
func planFolderSprites(dir string, sprites []project.SpriteRow, appendExtras bool) []spriteAssignment {
	pngs := folderPNGs(dir)
	used := make(map[string]bool)
	var assignments []spriteAssignment

	assign := func(rowIdx int, path, name string) {
		key := filepath.Clean(path)
		if used[key] {
			return
		}
		used[key] = true
		assignments = append(assignments, spriteAssignment{
			RowIdx: rowIdx,
			Name:   name,
			Path:   path,
			File:   relativeSpritePath(dir, path),
		})
	}

	for i, row := range sprites {
		for _, candidate := range spriteCandidates(dir, row, i) {
			key := filepath.Clean(candidate)
			if isRegularFile(candidate) && !used[key] {
				assign(i, candidate, "")
				break
			}
		}
	}

	if !appendExtras {
		return assignments
	}

	extra := 0
	for _, path := range pngs {
		if used[filepath.Clean(path)] {
			continue
		}
		assign(-1, path, spriteDefaultName(len(sprites)+extra))
		extra++
	}

	return assignments
}

func conventionSpriteName(idx int) string {
	if idx == 0 {
		return "base.png"
	}
	return fmt.Sprintf("sprite%d.png", idx+1)
}

// spriteCandidates lists where a sprite row's PNG may live, in priority order:
// the path recorded in the project, its basename inside the folder, the naming
// convention, and finally the original absolute path.
func spriteCandidates(dir string, row project.SpriteRow, idx int) []string {
	var out []string
	recorded := filepath.FromSlash(row.File)
	if row.File != "" {
		if !filepath.IsAbs(recorded) {
			out = append(out, filepath.Join(dir, recorded))
		}
		if base := filepath.Base(recorded); base != "" && base != "." {
			out = append(out, filepath.Join(dir, base))
		}
	}
	out = append(out, filepath.Join(dir, conventionSpriteName(idx)))
	if row.File != "" && filepath.IsAbs(recorded) {
		out = append(out, recorded)
	}
	return out
}

func folderPNGs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i]) < strings.ToLower(files[j])
	})
	return files
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// relativeSpritePath stores paths under dir as portable forward-slash paths.
// Absolute paths outside dir are left untouched.
func relativeSpritePath(dir, path string) string {
	if path == "" {
		return ""
	}
	if !filepath.IsAbs(path) {
		return filepath.ToSlash(filepath.Clean(path))
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}
	return filepath.ToSlash(rel)
}

func (a *EditorApp) openFolder(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}
	if stageConfig, configErr := findStageConfig(dir); configErr != nil {
		return configErr
	} else if stageConfig != "" {
		return a.openStageFolder(dir)
	}

	mode, known := modeFromFolderName(filepath.Base(filepath.Clean(dir)))
	if known && mode == modeStage {
		return a.openStageFolder(dir)
	}
	projectPath, err := findProjectGo(dir, mode)
	if err != nil {
		return err
	}
	if projectPath != "" {
		return a.openProjectFile(projectPath, true)
	}

	if !known {
		mode = modeMovement
	}
	a.clearLoadedSprites()
	a.proj = emptyProjectForMode(mode)
	a.currentFilePath = ""
	a.setEditorMode(mode)
	a.loadProjectSprites(dir, true)
	a.rebuildFromProj()
	return nil
}

func emptyProjectForMode(mode editorMode) project.ProjectData {
	assetName, assetKey := "MyMovement", "my_movement"
	if mode == modeAttack {
		assetName, assetKey = "MyAttack", "my_attack"
	}
	proj := project.ProjectData{
		AssetName:      assetName,
		AssetKey:       assetKey,
		DefaultOriginX: 0.5,
		DefaultOriginY: 0.5,
		Sprites:        []project.SpriteRow{},
		Animations:     []project.AnimationRow{},
	}
	if mode == modeStage {
		proj.Stage = project.NewStageData("Stage1")
	}
	return proj
}

func (a *EditorApp) openProjectFile(path string, appendExtras bool) error {
	if strings.EqualFold(filepath.Ext(path), ".json") {
		return a.openStageFile(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	switch {
	case bytes.Contains(data, []byte("= Stage{")) ||
		bytes.Contains(data, []byte("= StageDef{")) ||
		bytes.Contains(data, []byte("= stage.Stage{")) ||
		bytes.Contains(data, []byte("= stage.StageDef{")):
		a.clearLoadedSprites()
		return a.openStageFile(path)
	case bytes.Contains(data, []byte("= Attack{")):
		a.clearLoadedSprites()
		if err := a.openAttackFile(path); err != nil {
			return err
		}
		a.setEditorMode(modeAttack)
	case bytes.Contains(data, []byte("= Movement{")):
		a.clearLoadedSprites()
		if err := a.openMovementFile(path); err != nil {
			return err
		}
		a.setEditorMode(modeMovement)
	default:
		return fmt.Errorf("unknown file type: no Stage, Movement or Attack declaration found")
	}

	a.currentFilePath = path
	a.loadProjectSprites(filepath.Dir(path), appendExtras)
	a.rebuildFromProj()
	return nil
}

func (a *EditorApp) setEditorMode(mode editorMode) {
	prev := a.mode
	a.mode = mode
	if prev != mode {
		a.undoStack = nil
		a.redoStack = nil

		a.st.dragAsset = -1
		a.st.draggingFromPanel = false
		a.st.draggingObj = false
		a.st.dragMoved = false
		a.st.marqueeActive = false
		a.st.panning = false
		a.st.dragUndoSaved = false
		a.scaleHandle = -1
		a.clearStageInputFocus()
	}

	switch mode {
	case modeStage:
		a.rp.scroll = 0
		a.prev.previewPlaying = false
		a.ensureStageMode()
		a.selectStageObject(a.st.selectedObj)
	case modeAttack, modeMovement:
		if prev == modeStage {
			a.panelMode = panelModeAnimFrame
			a.resetAnimFramePropLabels()
			a.canvas.BoundWidth = 1280
			a.canvas.BoundHeight = 720
			a.rebuildFromProj()
		}
		a.syncLayout()
	}

	if a.modeDropdown == nil {
		return
	}
	switch mode {
	case modeAttack:
		a.modeDropdown.Selected = 1
		a.modeDropdown.DisplayText = "Attack Editor"
	case modeStage:
		a.modeDropdown.Selected = 2
		a.modeDropdown.DisplayText = "Stage Editor"
	default:
		a.modeDropdown.Selected = 0
		a.modeDropdown.DisplayText = "Movement Editor"
	}
}

func (a *EditorApp) clearLoadedSprites() {
	for _, img := range a.loadedSprites {
		if img != nil {
			img.Deallocate()
		}
	}
	a.loadedSprites = make(map[int]*ebiten.Image)
}

// needsFrameCountRecalc reports whether loading the assigned PNG should derive
// the row's frame count from the image dimensions. Imported movement/attack
// files without a Sprites section produce placeholder rows with File == ""
// and FrameCount 1; without recalculating, every animation frame clamps to
// frame 0 and the canvas never appears to animate. Rows that declare their
// spritesheet keep the declared frame count.
func needsFrameCountRecalc(sprites []project.SpriteRow, assignment spriteAssignment) bool {
	if assignment.RowIdx < 0 || assignment.RowIdx >= len(sprites) {
		return true
	}
	return sprites[assignment.RowIdx].File == ""
}

func (a *EditorApp) loadProjectSprites(dir string, appendExtras bool) {
	assignments := planFolderSprites(dir, a.proj.Sprites, appendExtras)
	for _, assignment := range assignments {
		idx := assignment.RowIdx
		if idx < 0 {
			idx = len(a.proj.Sprites)
			a.proj.Sprites = append(a.proj.Sprites, project.SpriteRow{
				Name:       assignment.Name,
				File:       assignment.File,
				Width:      256,
				Height:     256,
				FrameCount: 1,
				CurrentIdx: 0,
				ScaleX:     1,
				ScaleY:     1,
				OriginX:    0.5,
				OriginY:    0.5,
			})
		}
		if idx < 0 || idx >= len(a.proj.Sprites) {
			continue
		}
		if err := a.loadSpriteImage(idx, assignment.Path, assignment.File, needsFrameCountRecalc(a.proj.Sprites, assignment)); err != nil {
			if assignment.RowIdx < 0 {
				a.proj.Sprites = a.proj.Sprites[:idx]
			}
			a.setStatus("Sprite load error: " + err.Error())
		}
	}
	if len(assignments) > 0 {
		a.syncSpriteBtns()
	}
}

func (a *EditorApp) loadSpriteImage(idx int, path, storePath string, recalcFrameCount bool) error {
	if idx < 0 || idx >= len(a.proj.Sprites) {
		return fmt.Errorf("sprite index %d out of range", idx)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return fmt.Errorf("decode PNG: %w", err)
	}

	ebImg := ebiten.NewImageFromImage(img)
	if old, ok := a.loadedSprites[idx]; ok && old != nil {
		old.Deallocate()
	}
	a.loadedSprites[idx] = ebImg

	row := &a.proj.Sprites[idx]
	row.File = storePath
	if row.Width <= 0 {
		row.Width = 256
	}
	if row.Height <= 0 {
		row.Height = 256
	}

	if recalcFrameCount {
		cols := img.Bounds().Dx() / row.Width
		rows := img.Bounds().Dy() / row.Height
		if cols < 1 {
			cols = 1
		}
		if rows < 1 {
			rows = 1
		}
		row.FrameCount = cols * rows
		row.CurrentIdx = 0
	}

	if idx < len(a.spriteWidthInputs) {
		a.spriteWidthInputs[idx].SetNumeric(float64(row.Width))
	}
	if idx < len(a.spriteHeightInputs) {
		a.spriteHeightInputs[idx].SetNumeric(float64(row.Height))
	}

	if idx < len(a.spriteBrowseBtns) {
		ext := filepath.Ext(storePath)
		if ext == "" {
			ext = filepath.Ext(path)
		}
		ext = strings.TrimPrefix(ext, ".")
		if ext == "" {
			ext = "img"
		}
		a.spriteBrowseBtns[idx].Text = ext
		a.spriteBrowseBtns[idx].BtnType = ui.BtnGreen
	}

	return nil
}
