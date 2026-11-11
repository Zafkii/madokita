package editor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"animprite/internal/canvas"
	"animprite/internal/project"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

type stageAssetRect struct {
	AssetIdx   int
	X, Y, W, H int
}

type stageCategoryRect struct {
	Category   string
	X, Y, W, H int
}

type stageToolRect struct {
	Key        string
	X, Y, W, H int
}

type stageObjectDragStart struct {
	Index int
	X, Y  float64
}

type stageLayerRect struct {
	Index      int
	X, Y, W, H int
}

type stageState struct {
	images   map[string]*ebiten.Image
	expanded map[string]bool

	assetRects    []stageAssetRect
	categoryRects []stageCategoryRect
	layerRects    []stageLayerRect

	toolTab          string
	toolInfoAuto     bool
	toolInfoHold     int
	toolToggles      map[string]bool
	toolTabRects     []stageToolRect
	toolToggleRects  []stageToolRect
	toolPage         int
	toolPagePrevRect stageToolRect
	toolPageNextRect stageToolRect

	snap     bool
	snapRect struct{ X, Y, W, H int }

	selectedObj  int
	selectedObjs map[int]bool
	activeLayer  string

	dragAsset         int
	draggingFromPanel bool
	dragStartX        int
	dragStartY        int
	ghostWX           float64
	ghostWY           float64

	draggingObj          bool
	dragMoved            bool
	dragUndoSaved        bool
	dragObjStartX        float64
	dragObjStartY        float64
	dragWorldStartX      float64
	dragWorldStartY      float64
	dragObjStarts        []stageObjectDragStart
	marqueeActive        bool
	marqueeStartWX       float64
	marqueeStartWY       float64
	marqueeCurWX         float64
	marqueeCurWY         float64
	panning              bool
	stagePointerOnObject bool

	opacityInput   *ui.TextInput
	deleteBtn      *ui.Button
	duplicateBtn   *ui.Button
	layerAddBtn    *ui.Button
	layerDeleteBtn *ui.Button
	layerUpBtn     *ui.Button
	layerDownBtn   *ui.Button
}

func (a *EditorApp) ensureStageMode() {
	if a.proj.Stage.Version == 0 && a.proj.Stage.Name == "" {
		a.proj.Stage = project.NewStageData("Stage1")
	}
	if a.proj.Stage.Version == 0 {
		a.proj.Stage.Version = project.StageVersion
	}
	if a.proj.Stage.Name == "" {
		a.proj.Stage.Name = "Stage1"
	}
	if a.proj.Stage.Assets == nil {
		a.proj.Stage.Assets = []project.StageAsset{}
	}
	if a.proj.Stage.Objects == nil {
		a.proj.Stage.Objects = []project.StageObject{}
	}
	a.proj.Stage.NormalizeLayers()
	if a.proj.Stage.LayerIndex(a.st.activeLayer) < 0 {
		a.st.activeLayer = a.proj.Stage.Layers[len(a.proj.Stage.Layers)-1].Name
	}
	if a.st.images == nil {
		a.st.images = make(map[string]*ebiten.Image)
	}
	if a.st.expanded == nil {
		a.st.expanded = map[string]bool{
			project.StageCategoryBG:    true,
			project.StageCategoryBlock: true,
		}
	}
	if a.st.toolTab == "" {
		a.st.toolTab = "build"
	}
	if a.st.toolToggles == nil {
		a.st.toolToggles = make(map[string]bool)
	}
	if a.st.toolPage < 0 || a.st.toolPage >= stageToolPageCount() {
		a.st.toolPage = 0
	}
	if a.st.selectedObjs == nil {
		a.st.selectedObjs = make(map[int]bool)
	}
	if a.st.selectedObj < -1 || a.st.selectedObj >= len(a.proj.Stage.Objects) {
		a.st.selectedObj = -1
	}
	for idx := range a.st.selectedObjs {
		if idx < 0 || idx >= len(a.proj.Stage.Objects) {
			delete(a.st.selectedObjs, idx)
		}
	}
	if a.st.selectedObj >= 0 {
		a.st.selectedObjs[a.st.selectedObj] = true
	} else {
		a.st.selectedObjs = make(map[int]bool)
	}
	if a.st.dragAsset >= len(a.proj.Stage.Assets) {
		a.st.dragAsset = -1
	}
	a.setStageCanvasBounds()
	a.syncLayout()
}

func (a *EditorApp) setStageCanvasBounds() {
	w := a.proj.Stage.Width
	h := a.proj.Stage.Height
	if w <= 0 {
		w = project.DefaultStageWidth
	}
	if h <= 0 {
		h = project.DefaultStageHeight
	}
	a.canvas.BoundWidth = w
	a.canvas.BoundHeight = h
}

func (a *EditorApp) stageImageKey(file string) string {
	if file == "" {
		return ""
	}
	abs, err := filepath.Abs(file)
	if err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(file)
}

func (a *EditorApp) stageImage(file string) *ebiten.Image {
	key := a.stageImageKey(file)
	if key == "" {
		return nil
	}
	return a.st.images[key]
}

func (a *EditorApp) loadStageImage(file string) (*ebiten.Image, error) {
	key := a.stageImageKey(file)
	if key == "" {
		return nil, fmt.Errorf("empty image path")
	}
	if img := a.st.images[key]; img != nil {
		return img, nil
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open image: %w", err)
	}
	defer f.Close()

	decoded, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	img := ebiten.NewImageFromImage(decoded)
	a.st.images[key] = img
	return img, nil
}

func (a *EditorApp) clearStageImages() {
	for key, img := range a.st.images {
		if img != nil {
			img.Deallocate()
		}
		delete(a.st.images, key)
	}
	a.st.images = make(map[string]*ebiten.Image)
}

// scanStageAssets lists the top-stage PNGs in dir, sorted case-insensitively.
func scanStageAssets(dir string) ([]project.StageAsset, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var assets []project.StageAsset
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if abs, absErr := filepath.Abs(path); absErr == nil {
			path = filepath.Clean(abs)
		}

		f, openErr := os.Open(path)
		if openErr != nil {
			continue
		}
		cfg, decodeErr := png.DecodeConfig(f)
		f.Close()
		if decodeErr != nil {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		assets = append(assets, project.StageAsset{
			Category:   project.StageCategoryForFile(entry.Name()),
			Name:       name,
			File:       path,
			Width:      cfg.Width,
			Height:     cfg.Height,
			FrameCount: 1,
		})
	}

	sort.SliceStable(assets, func(i, j int) bool {
		return strings.ToLower(assets[i].Name) < strings.ToLower(assets[j].Name)
	})
	return assets, nil
}

func resolveStagePath(dir, file string) string {
	if file == "" {
		return ""
	}
	file = filepath.FromSlash(file)
	if filepath.IsAbs(file) {
		return filepath.Clean(file)
	}
	return filepath.Clean(filepath.Join(dir, file))
}

// findStageConfig returns the stage data file inside dir, preferring stage.go,
// then any .go file with a Stage declaration, then stage.json. It returns an
// empty string when the folder only contains images.
func findStageConfig(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var candidates []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if strings.EqualFold(entry.Name(), "stage.go") {
			return path, nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		for _, marker := range []string{
			"= Stage{",
			"= StageDef{",
			"= stage.Stage{",
			"= stage.StageDef{",
		} {
			if bytes.Contains(data, []byte(marker)) {
				candidates = append(candidates, path)
				break
			}
		}
	}
	if len(candidates) > 1 {
		return "", fmt.Errorf("multiple stage configs in %s: %s", dir, strings.Join(candidates, ", "))
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}

	jsonPath := filepath.Join(dir, "stage.json")
	if info, statErr := os.Stat(jsonPath); statErr == nil && !info.IsDir() {
		return jsonPath, nil
	}
	return "", nil
}

func (a *EditorApp) openStageFolder(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	configPath, err := findStageConfig(dir)
	if err != nil {
		return err
	}
	if configPath != "" {
		if err := a.openStageFile(configPath); err != nil {
			return err
		}
		// The folder may also contain images that are not in the config yet.
		if err := a.addStageAssetsFromFolder(dir); err != nil {
			a.setStatus("Asset scan error: " + err.Error())
		}
		return nil
	}

	assets, err := scanStageAssets(dir)
	if err != nil {
		return err
	}

	name := filepath.Base(filepath.Clean(dir))
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "Stage1"
	}
	stage := project.NewStageData(name)
	stage.Assets = assets
	a.applyLoadedStage(stage, "")
	return nil
}

func loadStageFile(path string) (project.StageData, error) {
	if !strings.EqualFold(filepath.Ext(path), ".json") {
		return ImportStage(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return project.StageData{}, err
	}

	var stage project.StageData
	if err := json.Unmarshal(data, &stage); err != nil {
		return project.StageData{}, fmt.Errorf("parse stage JSON: %w", err)
	}
	dir := filepath.Dir(path)
	for i := range stage.Assets {
		stage.Assets[i].File = resolveStagePath(dir, stage.Assets[i].File)
	}
	for i := range stage.Objects {
		stage.Objects[i].File = resolveStagePath(dir, stage.Objects[i].File)
	}
	return stage, nil
}

func (a *EditorApp) openStageFile(path string) error {
	stage, err := loadStageFile(path)
	if err != nil {
		return err
	}
	a.applyLoadedStage(stage, path)
	return nil
}

func (a *EditorApp) applyLoadedStage(stage project.StageData, path string) {
	if stage.Version == 0 {
		stage.Version = project.StageVersion
	}
	if stage.Name == "" && path != "" {
		stage.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if stage.Name == "" {
		stage.Name = "Stage1"
	}
	if stage.Width <= 0 {
		stage.Width = project.DefaultStageWidth
	}
	if stage.Height <= 0 {
		stage.Height = project.DefaultStageHeight
	}
	for i := range stage.Assets {
		if stage.Assets[i].Category == "" {
			stage.Assets[i].Category = project.StageCategoryForFile(stage.Assets[i].File)
		}
		if stage.Assets[i].Name == "" {
			stage.Assets[i].Name = strings.TrimSuffix(filepath.Base(stage.Assets[i].File), filepath.Ext(stage.Assets[i].File))
		}
	}
	for i := range stage.Objects {
		if stage.Objects[i].Category == "" {
			stage.Objects[i].Category = project.StageCategoryForFile(stage.Objects[i].File)
		}
	}
	stage.NormalizeLayers()

	a.clearStageImages()
	a.proj.Stage = stage
	a.st.activeLayer = stage.Layers[len(stage.Layers)-1].Name
	a.currentFilePath = path
	a.undoStack = nil
	a.redoStack = nil
	a.scaleHandle = -1
	a.st.selectedObj = -1
	a.st.selectedObjs = make(map[int]bool)
	a.st.dragObjStarts = a.st.dragObjStarts[:0]
	a.st.dragAsset = -1
	a.st.draggingFromPanel = false
	a.st.draggingObj = false
	a.st.dragMoved = false
	a.st.marqueeActive = false
	a.st.panning = false
	a.st.dragUndoSaved = false
	a.rp.scroll = 0
	a.clearStageInputFocus()

	for _, asset := range stage.Assets {
		if _, loadErr := a.loadStageImage(asset.File); loadErr != nil {
			a.setStatus("Image load error: " + loadErr.Error())
		}
	}
	for _, obj := range stage.Objects {
		if a.stageImage(obj.File) != nil {
			continue
		}
		if _, loadErr := a.loadStageImage(obj.File); loadErr != nil {
			a.setStatus("Image load error: " + loadErr.Error())
		}
	}

	a.setEditorMode(modeStage)
	a.selectStageObject(-1)
}

func (a *EditorApp) saveStageFile(path string) error {
	stage := a.proj.Stage.DeepCopy()
	if stage.Version == 0 {
		stage.Version = project.StageVersion
	}
	if stage.Name == "" {
		stage.Name = "Stage1"
	}
	if stage.Width <= 0 {
		stage.Width = project.DefaultStageWidth
	}
	if stage.Height <= 0 {
		stage.Height = project.DefaultStageHeight
	}
	stage.NormalizeLayers()
	if !strings.EqualFold(filepath.Ext(path), ".json") {
		return ExportStage(path, stage)
	}

	dir := filepath.Dir(path)
	for i := range stage.Assets {
		stage.Assets[i].File = relativeSpritePath(dir, stage.Assets[i].File)
	}
	for i := range stage.Objects {
		stage.Objects[i].File = relativeSpritePath(dir, stage.Objects[i].File)
	}

	data, err := json.MarshalIndent(stage, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (a *EditorApp) addStageAssetsFromFolder(dir string) error {
	assets, err := scanStageAssets(dir)
	if err != nil {
		return err
	}

	existing := make(map[string]bool, len(a.proj.Stage.Assets))
	for _, asset := range a.proj.Stage.Assets {
		existing[a.stageImageKey(asset.File)] = true
	}
	if a.st.expanded == nil {
		a.st.expanded = map[string]bool{}
	}
	for _, asset := range assets {
		key := a.stageImageKey(asset.File)
		if existing[key] {
			continue
		}
		existing[key] = true
		a.proj.Stage.Assets = append(a.proj.Stage.Assets, asset)
		a.st.expanded[asset.Category] = true
		if _, loadErr := a.loadStageImage(asset.File); loadErr != nil {
			a.setStatus("Image load error: " + loadErr.Error())
		}
	}
	return nil
}

func (a *EditorApp) stageDefaultSize(asset project.StageAsset) (float64, float64) {
	if asset.Category == project.StageCategoryBlock {
		return project.DefaultBlockSize, project.DefaultBlockSize
	}
	w := float64(asset.Width)
	h := float64(asset.Height)
	if w <= 0 {
		w = project.DefaultBlockSize
	}
	if h <= 0 {
		h = project.DefaultBlockSize
	}
	return w, h
}

func (a *EditorApp) addStageObjectFromAsset(assetIdx int, wx, wy float64) {
	if assetIdx < 0 || assetIdx >= len(a.proj.Stage.Assets) {
		return
	}
	asset := a.proj.Stage.Assets[assetIdx]
	w, h := a.stageDefaultSize(asset)
	wx, wy = snapStagePosition(wx, wy, w, h, a.st.snap)

	a.saveSnapshot()
	a.ensureActiveStageLayer()
	a.proj.Stage.Objects = append(a.proj.Stage.Objects, project.StageObject{
		Layer:    a.st.activeLayer,
		Category: asset.Category,
		File:     asset.File,
		X:        wx,
		Y:        wy,
		Width:    w,
		Height:   h,
		Rotation: 0,
		Opacity:  1,
	})
	a.selectStageObject(len(a.proj.Stage.Objects) - 1)
	a.setStatus("Added " + asset.Name)
}

// snapStagePosition snaps an object's top-left corner to the 50px world grid.
// wx/wy are the desired center; the returned values are the snapped center.
func snapStagePosition(wx, wy, w, h float64, enabled bool) (float64, float64) {
	if !enabled {
		return wx, wy
	}
	tlx := math.Round((wx-w/2)/project.DefaultBlockSize) * project.DefaultBlockSize
	tly := math.Round((wy-h/2)/project.DefaultBlockSize) * project.DefaultBlockSize
	return tlx + w/2, tly + h/2
}

// dragStagePosition applies the world delta accumulated since the drag started
// instead of the last frame's delta. Snapping the per-frame delta would reset
// slow drags back to the same grid cell and the object would never move.
// Snapping the accumulated delta preserves the object's original off-grid
// position when the cursor returns to where the drag started.
func dragStagePosition(startX, startY, startWX, startWY, wx, wy, w, h float64, snap bool) (float64, float64) {
	dx := wx - startWX
	dy := wy - startWY
	if snap {
		dx = math.Round(dx/project.DefaultBlockSize) * project.DefaultBlockSize
		dy = math.Round(dy/project.DefaultBlockSize) * project.DefaultBlockSize
	}
	return startX + dx, startY + dy
}

// markStageObjectLiveEdit moves the stage toolbar to Info while the selected
// object is being changed. tickStageInfoAuto returns to Edit once the edit has
// been idle for a moment, keeping the write actions close at hand again.
func (a *EditorApp) markStageObjectLiveEdit() {
	if a.selectedStageObject() == nil {
		return
	}
	a.st.toolTab = "info"
	a.st.toolInfoAuto = true
	a.st.toolInfoHold = stageInfoHoldFrames
}

func (a *EditorApp) tickStageInfoAuto() {
	if !a.isStageMode() || !a.st.toolInfoAuto {
		return
	}
	if a.focusedInput() != nil {
		return
	}
	if a.st.toolInfoHold > 0 {
		a.st.toolInfoHold--
	}
	if a.st.toolInfoHold == 0 {
		a.st.toolTab = "edit"
		a.st.toolInfoAuto = false
	}
}

func (a *EditorApp) selectStageObject(idx int) {
	if idx < 0 || idx >= len(a.proj.Stage.Objects) {
		idx = -1
	}
	a.st.selectedObj = idx
	a.st.selectedObjs = make(map[int]bool)
	if idx >= 0 {
		a.st.selectedObjs[idx] = true
		a.followStageObjectLayer(idx)
	}
	a.syncStageSelectionUI()
}

func (a *EditorApp) followStageObjectLayer(idx int) {
	if idx < 0 || idx >= len(a.proj.Stage.Objects) {
		return
	}
	a.ensureActiveStageLayer()
	a.st.activeLayer = a.proj.Stage.LayerNameFor(a.proj.Stage.Objects[idx])
}

func (a *EditorApp) syncStageSelectionUI() {
	a.panelMode = panelModeStage
	if a.st.selectedObj >= 0 || a.st.toolTab == "info" {
		a.st.toolTab = "edit"
	}
	a.st.toolInfoAuto = false
	a.st.toolInfoHold = 0

	obj := a.selectedStageObject()
	if obj == nil {
		a.canvas.SetSelectionRect(canvas.SelectionRect{Visible: false})
		return
	}

	a.props[0].SetLabel("X")
	a.props[0].SetNumeric(obj.X)
	a.props[0].Min = -99999
	a.props[0].Step = 1
	a.props[1].SetLabel("Y")
	a.props[1].SetNumeric(obj.Y)
	a.props[1].Min = -99999
	a.props[1].Step = 1
	a.props[2].SetLabel("Width")
	a.props[2].SetNumeric(obj.Width)
	a.props[2].Min = 1
	a.props[2].Step = 1
	a.props[3].SetLabel("Height")
	a.props[3].SetNumeric(obj.Height)
	a.props[3].Min = 1
	a.props[3].Step = 1
	a.props[4].SetLabel("Rotation (°)")
	a.props[4].SetNumeric(obj.Rotation)
	a.props[4].Min = -360
	a.props[4].Step = 0.5
	if a.st.opacityInput != nil {
		a.st.opacityInput.SetNumeric(obj.Opacity)
	}

	a.canvas.SetSelectionRect(canvas.SelectionRect{
		X:        obj.X - obj.Width/2,
		Y:        obj.Y - obj.Height/2,
		W:        obj.Width,
		H:        obj.Height,
		Rotation: obj.Rotation * math.Pi / 180,
		Visible:  true,
	})
}

func (a *EditorApp) stageSelectedIndices() []int {
	if a.st.selectedObjs == nil {
		a.st.selectedObjs = make(map[int]bool)
	}
	if a.st.selectedObj >= 0 && a.st.selectedObj < len(a.proj.Stage.Objects) {
		a.st.selectedObjs[a.st.selectedObj] = true
	}
	indices := make([]int, 0, len(a.st.selectedObjs))
	for idx := range a.st.selectedObjs {
		if idx >= 0 && idx < len(a.proj.Stage.Objects) {
			indices = append(indices, idx)
		}
	}
	sort.Ints(indices)
	return indices
}

func (a *EditorApp) stageObjectSelected(idx int) bool {
	if idx < 0 || idx >= len(a.proj.Stage.Objects) {
		return false
	}
	if idx == a.st.selectedObj {
		return true
	}
	return a.st.selectedObjs != nil && a.st.selectedObjs[idx]
}

func (a *EditorApp) stageSelectionCount() int {
	return len(a.stageSelectedIndices())
}

func (a *EditorApp) selectAllStageObjects() {
	if len(a.proj.Stage.Objects) == 0 {
		return
	}
	a.st.selectedObjs = make(map[int]bool, len(a.proj.Stage.Objects))
	for idx := range a.proj.Stage.Objects {
		a.st.selectedObjs[idx] = true
	}
	a.st.selectedObj = len(a.proj.Stage.Objects) - 1
	a.syncStageSelectionUI()
	a.setStatus(fmt.Sprintf("Selected %d objects", len(a.proj.Stage.Objects)))
}

func (a *EditorApp) addStageObjectSelection(idx int) {
	if idx < 0 || idx >= len(a.proj.Stage.Objects) {
		return
	}
	if a.st.selectedObjs == nil {
		a.st.selectedObjs = make(map[int]bool)
	}
	a.st.selectedObjs[idx] = true
	a.st.selectedObj = idx
	a.followStageObjectLayer(idx)
	a.syncStageSelectionUI()
	a.showStageSelectionCount()
}

func (a *EditorApp) toggleStageObjectSelection(idx int) {
	if idx < 0 || idx >= len(a.proj.Stage.Objects) {
		return
	}
	if a.st.selectedObjs == nil {
		a.st.selectedObjs = make(map[int]bool)
	}
	if a.stageObjectSelected(idx) {
		delete(a.st.selectedObjs, idx)
		if a.st.selectedObj == idx {
			a.st.selectedObj = -1
			if len(a.st.selectedObjs) > 0 {
				indices := a.stageSelectedIndices()
				if len(indices) > 0 {
					a.st.selectedObj = indices[0]
				}
			}
		}
	} else {
		a.st.selectedObjs[idx] = true
		a.st.selectedObj = idx
		a.followStageObjectLayer(idx)
	}
	a.syncStageSelectionUI()
	a.showStageSelectionCount()
}

func (a *EditorApp) showStageSelectionCount() {
	switch count := a.stageSelectionCount(); count {
	case 0:
		a.setStatus("Selection cleared")
	case 1:
		a.setStatus("Selected object")
	default:
		a.setStatus(fmt.Sprintf("Selected %d objects", count))
	}
}

func (a *EditorApp) setStageObjectPrimary(idx int) {
	if !a.stageObjectSelected(idx) {
		return
	}
	a.st.selectedObj = idx
	a.followStageObjectLayer(idx)
	a.syncStageSelectionUI()
}

func (a *EditorApp) beginStageObjectDrag(wx, wy float64) {
	obj := a.selectedStageObject()
	if obj == nil {
		return
	}
	a.st.dragObjStartX = obj.X
	a.st.dragObjStartY = obj.Y
	a.st.dragWorldStartX = wx
	a.st.dragWorldStartY = wy
	a.st.dragObjStarts = a.st.dragObjStarts[:0]
	for _, idx := range a.stageSelectedIndices() {
		start := &a.proj.Stage.Objects[idx]
		a.st.dragObjStarts = append(a.st.dragObjStarts, stageObjectDragStart{
			Index: idx,
			X:     start.X,
			Y:     start.Y,
		})
	}
	a.st.draggingObj = true
	a.st.dragMoved = false
	a.st.dragUndoSaved = false
}

func (a *EditorApp) applyStageObjectDrag(newX, newY float64) {
	dx := newX - a.st.dragObjStartX
	dy := newY - a.st.dragObjStartY
	if dx == 0 && dy == 0 && !a.st.dragMoved {
		return
	}
	if !a.st.dragUndoSaved {
		a.saveSnapshot()
		a.st.dragUndoSaved = true
	}
	for _, start := range a.st.dragObjStarts {
		if start.Index < 0 || start.Index >= len(a.proj.Stage.Objects) {
			continue
		}
		moved := &a.proj.Stage.Objects[start.Index]
		moved.X = start.X + dx
		moved.Y = start.Y + dy
	}
	a.st.dragMoved = true
	a.syncStageInputs()
	a.markStageObjectLiveEdit()
}

func (a *EditorApp) beginStageMarquee(wx, wy float64) {
	a.st.marqueeActive = true
	a.st.marqueeStartWX = wx
	a.st.marqueeStartWY = wy
	a.st.marqueeCurWX = wx
	a.st.marqueeCurWY = wy
	a.st.stagePointerOnObject = false
	a.st.draggingObj = false
	a.st.panning = false
	a.selectStageObject(-1)
}

func (a *EditorApp) updateStageMarquee(wx, wy float64) {
	if !a.st.marqueeActive {
		return
	}
	a.st.marqueeCurWX = wx
	a.st.marqueeCurWY = wy
	a.selectStageObjectsInMarquee()
}

func (a *EditorApp) finishStageMarquee() {
	if !a.st.marqueeActive {
		return
	}
	a.st.marqueeActive = false
	switch count := a.stageSelectionCount(); count {
	case 0:
		a.setStatus("No objects selected")
	case 1:
		a.setStatus("Selected object")
	default:
		a.setStatus(fmt.Sprintf("Selected %d objects", count))
	}
}

func (a *EditorApp) selectStageObjectsInMarquee() {
	if !a.st.marqueeActive {
		return
	}
	minX, maxX := a.st.marqueeStartWX, a.st.marqueeCurWX
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := a.st.marqueeStartWY, a.st.marqueeCurWY
	if minY > maxY {
		minY, maxY = maxY, minY
	}

	selected := make(map[int]bool)
	for idx := range a.proj.Stage.Objects {
		objMinX, objMinY, objMaxX, objMaxY := stageObjectWorldBounds(&a.proj.Stage.Objects[idx])
		if objMaxX < minX || objMinX > maxX || objMaxY < minY || objMinY > maxY {
			continue
		}
		selected[idx] = true
	}
	a.st.selectedObjs = selected
	a.st.selectedObj = -1
	if len(selected) > 0 {
		indices := a.stageSelectedIndices()
		if len(indices) > 0 {
			a.st.selectedObj = indices[0]
		}
	}
	a.syncStageSelectionUI()
}

func stageObjectWorldBounds(obj *project.StageObject) (minX, minY, maxX, maxY float64) {
	rot := obj.Rotation * math.Pi / 180
	cos := math.Cos(rot)
	sin := math.Sin(rot)
	hw := obj.Width / 2
	hh := obj.Height / 2
	corners := [4][2]float64{
		{-hw, -hh},
		{hw, -hh},
		{hw, hh},
		{-hw, hh},
	}
	minX, minY = math.Inf(1), math.Inf(1)
	maxX, maxY = math.Inf(-1), math.Inf(-1)
	for _, corner := range corners {
		x := obj.X + corner[0]*cos - corner[1]*sin
		y := obj.Y + corner[0]*sin + corner[1]*cos
		minX = math.Min(minX, x)
		minY = math.Min(minY, y)
		maxX = math.Max(maxX, x)
		maxY = math.Max(maxY, y)
	}
	return minX, minY, maxX, maxY
}

func (a *EditorApp) selectedStageObject() *project.StageObject {
	if a.st.selectedObj < 0 || a.st.selectedObj >= len(a.proj.Stage.Objects) {
		return nil
	}
	return &a.proj.Stage.Objects[a.st.selectedObj]
}

func (a *EditorApp) deleteSelectedStageObject() {
	indices := a.stageSelectedIndices()
	if len(indices) == 0 {
		return
	}
	a.saveSnapshot()
	selected := make(map[int]bool, len(indices))
	for _, idx := range indices {
		selected[idx] = true
	}
	objects := make([]project.StageObject, 0, len(a.proj.Stage.Objects)-len(indices))
	for idx, obj := range a.proj.Stage.Objects {
		if !selected[idx] {
			objects = append(objects, obj)
		}
	}
	a.proj.Stage.Objects = objects
	a.selectStageObject(-1)
	if len(indices) == 1 {
		a.setStatus("Deleted object")
	} else {
		a.setStatus(fmt.Sprintf("Deleted %d objects", len(indices)))
	}
}

func (a *EditorApp) duplicateSelectedStageObject() {
	indices := a.stageSelectedIndices()
	if len(indices) == 0 {
		return
	}
	a.saveSnapshot()
	newIndices := make([]int, 0, len(indices))
	for _, idx := range indices {
		dup := a.proj.Stage.Objects[idx]
		dup.X += project.DefaultBlockSize
		dup.Y += project.DefaultBlockSize
		dup.X, dup.Y = snapStagePosition(dup.X, dup.Y, dup.Width, dup.Height, a.st.snap)
		newIndices = append(newIndices, len(a.proj.Stage.Objects))
		a.proj.Stage.Objects = append(a.proj.Stage.Objects, dup)
	}

	a.st.selectedObjs = make(map[int]bool, len(newIndices))
	for _, idx := range newIndices {
		a.st.selectedObjs[idx] = true
	}
	a.st.selectedObj = newIndices[len(newIndices)-1]
	a.syncStageSelectionUI()
	if len(newIndices) == 1 {
		a.setStatus("Duplicated object")
	} else {
		a.setStatus(fmt.Sprintf("Duplicated %d objects", len(newIndices)))
	}
}

func (a *EditorApp) ensureActiveStageLayer() {
	if len(a.proj.Stage.Layers) == 0 {
		a.proj.Stage.NormalizeLayers()
	}
	if a.proj.Stage.LayerIndex(a.st.activeLayer) >= 0 {
		return
	}
	a.proj.Stage.NormalizeLayers()
	a.st.activeLayer = a.proj.Stage.Layers[len(a.proj.Stage.Layers)-1].Name
}

func (a *EditorApp) activeStageLayerIndex() int {
	a.ensureActiveStageLayer()
	return a.proj.Stage.LayerIndex(a.st.activeLayer)
}

func (a *EditorApp) setActiveStageLayer(idx int) {
	a.ensureActiveStageLayer()
	if idx < 0 || idx >= len(a.proj.Stage.Layers) {
		return
	}
	name := a.proj.Stage.Layers[idx].Name
	if a.st.activeLayer == name {
		a.setStatus("Active layer: " + name)
		return
	}
	a.st.activeLayer = name
	a.selectStageObject(-1)
	a.setStatus("Active layer: " + name)
}

func (a *EditorApp) nextStageLayerName() string {
	used := make(map[string]bool, len(a.proj.Stage.Layers))
	for _, layer := range a.proj.Stage.Layers {
		used[layer.Name] = true
	}
	for n := 1; ; n++ {
		name := fmt.Sprintf("Layer %d", n)
		if !used[name] {
			return name
		}
	}
}

func (a *EditorApp) addStageLayer() {
	a.saveSnapshot()
	name := a.nextStageLayerName()
	a.proj.Stage.Layers = append(a.proj.Stage.Layers, project.StageLayer{Name: name})
	a.st.activeLayer = name
	a.setStatus("Added " + name)
}

func (a *EditorApp) deleteActiveStageLayer() {
	a.ensureActiveStageLayer()
	idx := a.activeStageLayerIndex()
	if idx < 0 || len(a.proj.Stage.Layers) <= 1 {
		a.setStatus("Cannot delete the only layer")
		return
	}
	layer := a.proj.Stage.Layers[idx]
	for _, obj := range a.proj.Stage.Objects {
		if a.proj.Stage.LayerNameFor(obj) == layer.Name {
			a.setStatus("Layer is not empty")
			return
		}
	}
	a.saveSnapshot()
	a.proj.Stage.Layers = append(a.proj.Stage.Layers[:idx], a.proj.Stage.Layers[idx+1:]...)
	if idx >= len(a.proj.Stage.Layers) {
		idx = len(a.proj.Stage.Layers) - 1
	}
	a.st.activeLayer = a.proj.Stage.Layers[idx].Name
	a.setStatus("Deleted " + layer.Name)
}

func (a *EditorApp) moveActiveStageLayer(delta int) {
	idx := a.activeStageLayerIndex()
	target := idx + delta
	if idx < 0 || delta == 0 || target < 0 || target >= len(a.proj.Stage.Layers) {
		return
	}
	a.saveSnapshot()
	layers := a.proj.Stage.Layers
	layers[idx], layers[target] = layers[target], layers[idx]
	a.st.activeLayer = layers[target].Name
	if delta > 0 {
		a.setStatus("Layer moved forward")
	} else {
		a.setStatus("Layer moved backward")
	}
}

func (a *EditorApp) stageObjectAtWorld(wx, wy float64) int {
	if len(a.proj.Stage.Layers) == 0 {
		a.proj.Stage.NormalizeLayers()
	}
	objects := a.proj.Stage.Objects
	// Walk layers from front to back. Inside one layer, blocks stay ahead of
	// backgrounds so a full-stage backdrop does not swallow block clicks.
	for li := len(a.proj.Stage.Layers) - 1; li >= 0; li-- {
		layer := a.proj.Stage.Layers[li].Name
		for _, wantBlock := range []bool{true, false} {
			for i := len(objects) - 1; i >= 0; i-- {
				obj := &objects[i]
				if a.proj.Stage.LayerNameFor(*obj) != layer {
					continue
				}
				isBlock := obj.Category == project.StageCategoryBlock
				if isBlock != wantBlock {
					continue
				}
				dx := wx - obj.X
				dy := wy - obj.Y
				rot := -obj.Rotation * math.Pi / 180
				cos := math.Cos(rot)
				sin := math.Sin(rot)
				lx := dx*cos - dy*sin
				ly := dx*sin + dy*cos
				if math.Abs(lx) <= obj.Width/2 && math.Abs(ly) <= obj.Height/2 {
					return i
				}
			}
		}
	}
	return -1
}

func (a *EditorApp) updateStageAssetDrag(mx, my int, leftDown bool) {
	if !a.isStageMode() || a.st.dragAsset < 0 {
		return
	}
	assetIdx := a.st.dragAsset
	if assetIdx < 0 || assetIdx >= len(a.proj.Stage.Assets) {
		a.st.dragAsset = -1
		a.st.draggingFromPanel = false
		return
	}

	if leftDown {
		if a.pointInCanvas(mx, my) {
			asset := a.proj.Stage.Assets[assetIdx]
			w, h := a.stageDefaultSize(asset)
			wx, wy := a.canvasWorld(mx, my)
			wx, wy = snapStagePosition(wx, wy, w, h, a.st.snap)
			a.st.ghostWX = wx
			a.st.ghostWY = wy
			a.st.draggingFromPanel = true
		} else {
			a.st.draggingFromPanel = false
		}
		return
	}

	overCanvas := a.pointInCanvas(mx, my)
	moved := absInt(mx-a.st.dragStartX)+absInt(my-a.st.dragStartY) >= 4
	switch {
	case overCanvas:
		a.addStageObjectFromAsset(assetIdx, a.st.ghostWX, a.st.ghostWY)
	case !moved:
		cw := float64(a.canvas.Width)
		ch := float64(a.canvas.Height)
		wx, wy := a.canvas.Cam.CanvasToWorld(cw/2, ch/2)
		a.addStageObjectFromAsset(assetIdx, wx, wy)
	}

	a.st.dragAsset = -1
	a.st.draggingFromPanel = false
}

func (a *EditorApp) flushStageInputs() {
	if a.scaleHandle >= 0 || a.st.draggingObj || a.st.draggingFromPanel {
		return
	}
	primary := a.selectedStageObject()
	if primary == nil {
		return
	}
	indices := a.stageSelectedIndices()
	if len(indices) == 0 {
		return
	}
	focused := a.focusedInput()
	multi := len(indices) > 1
	if multi && focused == nil {
		return
	}

	x := a.props[0].NumericValue()
	y := a.props[1].NumericValue()
	w := a.props[2].NumericValue()
	if w < 1 {
		w = 1
	}
	h := a.props[3].NumericValue()
	if h < 1 {
		h = 1
	}
	rot := a.props[4].NumericValue()
	opacity := primary.Opacity
	if a.st.opacityInput != nil {
		opacity = clamp(a.st.opacityInput.NumericValue(), 0, 1)
	}

	origX := primary.X
	origY := primary.Y
	origW := primary.Width
	origH := primary.Height
	origRot := primary.Rotation
	origOpacity := primary.Opacity

	applyX := !multi || x != origX
	applyY := !multi || y != origY
	applyW := !multi || w != origW
	applyH := !multi || h != origH
	applyRot := !multi || rot != origRot
	applyOpacity := !multi || opacity != origOpacity
	if !applyX && !applyY && !applyW && !applyH && !applyRot && !applyOpacity {
		return
	}

	dx := x - origX
	dy := y - origY
	changed := false
	for _, idx := range indices {
		obj := &a.proj.Stage.Objects[idx]
		if applyX {
			next := x
			if idx != a.st.selectedObj {
				next = obj.X + dx
			}
			if obj.X != next {
				changed = true
				obj.X = next
			}
		}
		if applyY {
			next := y
			if idx != a.st.selectedObj {
				next = obj.Y + dy
			}
			if obj.Y != next {
				changed = true
				obj.Y = next
			}
		}
		if applyW && obj.Width != w {
			changed = true
			obj.Width = w
		}
		if applyH && obj.Height != h {
			changed = true
			obj.Height = h
		}
		if applyRot && obj.Rotation != rot {
			changed = true
			obj.Rotation = rot
		}
		if applyOpacity && obj.Opacity != opacity {
			changed = true
			obj.Opacity = opacity
		}
	}
	if changed && focused != nil {
		a.markStageObjectLiveEdit()
	}
}

func (a *EditorApp) syncStageInputs() {
	obj := a.selectedStageObject()
	if obj == nil {
		return
	}
	a.props[0].SetNumeric(obj.X)
	a.props[1].SetNumeric(obj.Y)
	a.props[2].SetNumeric(obj.Width)
	a.props[3].SetNumeric(obj.Height)
	a.props[4].SetNumeric(obj.Rotation)
	if a.st.opacityInput != nil {
		a.st.opacityInput.SetNumeric(obj.Opacity)
	}
}
