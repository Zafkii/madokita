package editor

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"animprite/internal/project"
)

func TestStageCategoryForFile(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"bg.png", project.StageCategoryBG},
		{"background_sky.png", project.StageCategoryBG},
		{"sky2.png", project.StageCategoryBG},
		{"block_grassfloor.png", project.StageCategoryBlock},
		{"floor.png", project.StageCategoryBlock},
		{"platform.png", project.StageCategoryBlock},
		{"random.png", project.StageCategoryBlock},
		{"prop_tree.png", project.StageCategoryProp},
		{"decor_rock.png", project.StageCategoryProp},
		{"scenery_hill.png", project.StageCategoryProp},
		{"background_prop.png", project.StageCategoryBG},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := project.StageCategoryForFile(tt.name); got != tt.want {
				t.Fatalf("StageCategoryForFile(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestSnapStagePosition(t *testing.T) {
	tests := []struct {
		name    string
		wx, wy  float64
		w, h    float64
		enabled bool
		wantX   float64
		wantY   float64
	}{
		{"block snaps top-left", 37, 112, 50, 50, true, 25, 125},
		{"block snaps to next cell", 60, 60, 50, 50, true, 75, 75},
		{"background snaps top-left", 1307, 372, 2560, 720, true, 1330, 360},
		{"snapping disabled", 37, 112, 100, 50, false, 37, 112},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := snapStagePosition(tt.wx, tt.wy, tt.w, tt.h, tt.enabled)
			if gotX != tt.wantX || gotY != tt.wantY {
				t.Fatalf("snapStagePosition() = (%v, %v), want (%v, %v)", gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestDragStagePositionAccumulatesAcrossFrames(t *testing.T) {
	// A block starts centered at (25, 25), top-left on the grid. Each frame
	// moves the cursor only 3 world units; snapping the per-frame delta would
	// reset the block to the same cell forever. The accumulated delta must win.
	startX, startY := 25.0, 25.0
	var x, y float64
	for step := 1; step <= 10; step++ {
		x, y = dragStagePosition(startX, startY, 0, 0, float64(step*3), 0, 50, 50, true)
	}
	if x != 75 || y != 25 {
		t.Fatalf("accumulated drag = (%v, %v), want (75, 25)", x, y)
	}
}

func TestDragStagePositionSnapsBackground(t *testing.T) {
	// A 2560x720 background centered on the origin fills the 2560x720 stage
	// bounds. Dragging it by (30, 40) must snap its top-left to the 50px grid.
	x, y := dragStagePosition(1280, 360, 0, 0, 30, 40, 2560, 720, true)
	if x != 1330 || y != 410 {
		t.Fatalf("background drag = (%v, %v), want (1330, 410)", x, y)
	}
}

func TestDragStagePositionBackgroundWithoutSnap(t *testing.T) {
	x, y := dragStagePosition(1280, 360, 0, 0, 10, 20, 2560, 720, false)
	if x != 1290 || y != 380 {
		t.Fatalf("background drag = (%v, %v), want (1290, 380)", x, y)
	}
}

func TestScanStageAssets(t *testing.T) {
	dir := t.TempDir()
	writeTestPNG(t, filepath.Join(dir, "block_grassfloor.png"), 10, 20)
	writeTestPNG(t, filepath.Join(dir, "wide_bg.png"), 30, 40)

	assets, err := scanStageAssets(dir)
	if err != nil {
		t.Fatalf("scanStageAssets() error: %v", err)
	}
	if len(assets) != 2 {
		t.Fatalf("scanStageAssets() returned %d assets, want 2", len(assets))
	}

	// Sorted case-insensitively: block_grassfloor before wide_bg.
	if assets[0].Name != "block_grassfloor" || assets[0].Category != project.StageCategoryBlock {
		t.Fatalf("first asset = %+v, want block_grassfloor block", assets[0])
	}
	if assets[0].Width != 10 || assets[0].Height != 20 {
		t.Fatalf("first asset size = %dx%d, want 10x20", assets[0].Width, assets[0].Height)
	}
	if !filepath.IsAbs(assets[0].File) {
		t.Fatalf("first asset file = %q, want absolute path", assets[0].File)
	}

	if assets[1].Name != "wide_bg" || assets[1].Category != project.StageCategoryBG {
		t.Fatalf("second asset = %+v, want wide_bg bg", assets[1])
	}
	if assets[1].Width != 30 || assets[1].Height != 40 {
		t.Fatalf("second asset size = %dx%d, want 30x40", assets[1].Width, assets[1].Height)
	}
	if !filepath.IsAbs(assets[1].File) {
		t.Fatalf("second asset file = %q, want absolute path", assets[1].File)
	}
}

func TestSaveStageFileUsesRelativePaths(t *testing.T) {
	dir := t.TempDir()
	assetDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		t.Fatalf("create assets dir: %v", err)
	}
	assetPath := filepath.Join(assetDir, "bg.png")
	objectPath := filepath.Join(assetDir, "block_grassfloor.png")
	writeTestPNG(t, assetPath, 64, 32)
	writeTestPNG(t, objectPath, 50, 50)

	app := &EditorApp{}
	app.proj.Stage = project.NewStageData("Test Stage")
	app.proj.Stage.Assets = []project.StageAsset{{
		Category: project.StageCategoryBG,
		Name:     "bg",
		File:     assetPath,
		Width:    64,
		Height:   32,
	}}
	app.proj.Stage.Objects = []project.StageObject{{
		Category: project.StageCategoryBlock,
		File:     objectPath,
		X:        25,
		Y:        25,
		Width:    50,
		Height:   50,
		Opacity:  1,
	}}

	outPath := filepath.Join(dir, "stage.json")
	if err := app.saveStageFile(outPath); err != nil {
		t.Fatalf("saveStageFile() error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read stage.json: %v", err)
	}
	var saved project.StageData
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal stage.json: %v", err)
	}
	if saved.Version != project.StageVersion {
		t.Fatalf("saved version = %d, want %d", saved.Version, project.StageVersion)
	}
	if saved.Name != "Test Stage" {
		t.Fatalf("saved name = %q, want Test Stage", saved.Name)
	}
	if saved.Assets[0].File != "assets/bg.png" {
		t.Fatalf("asset path = %q, want assets/bg.png", saved.Assets[0].File)
	}
	if saved.Objects[0].File != "assets/block_grassfloor.png" {
		t.Fatalf("object path = %q, want assets/block_grassfloor.png", saved.Objects[0].File)
	}
	if len(saved.Layers) != 1 || saved.Layers[0].Name != project.DefaultStageLayer {
		t.Fatalf("saved layers = %+v, want default layer", saved.Layers)
	}
	if saved.Objects[0].Layer != project.DefaultStageLayer {
		t.Fatalf("object layer = %q, want %q", saved.Objects[0].Layer, project.DefaultStageLayer)
	}

	// Saving must not mutate the in-memory project paths.
	if app.proj.Stage.Assets[0].File != assetPath {
		t.Fatalf("in-memory asset path mutated to %q", app.proj.Stage.Assets[0].File)
	}
	if app.proj.Stage.Objects[0].File != objectPath {
		t.Fatalf("in-memory object path mutated to %q", app.proj.Stage.Objects[0].File)
	}
}

func TestExportImportStageGo(t *testing.T) {
	dir := t.TempDir()
	assetPath := filepath.Join(dir, "bg.png")
	objectPath := filepath.Join(dir, "block_grassfloor.png")
	writeTestPNG(t, assetPath, 64, 32)
	writeTestPNG(t, objectPath, 50, 50)

	stage := project.NewStageData("Test Stage")
	stage.Assets = []project.StageAsset{{
		Category:   project.StageCategoryBG,
		Name:       "bg",
		File:       assetPath,
		Width:      64,
		Height:     32,
		FrameCount: 1,
	}}
	stage.Objects = []project.StageObject{{
		Category: project.StageCategoryBlock,
		File:     objectPath,
		X:        25,
		Y:        25,
		Width:    50,
		Height:   50,
		Opacity:  1,
	}}

	outPath := filepath.Join(dir, "stage.go")
	if err := ExportStage(outPath, stage); err != nil {
		t.Fatalf("ExportStage() error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read stage.go: %v", err)
	}
	if !bytes.Contains(data, []byte("package stages")) {
		t.Fatalf("stage.go missing package declaration:\n%s", data)
	}
	if !bytes.Contains(data, []byte("var TestStage = Stage{")) {
		t.Fatalf("stage.go missing Stage variable:\n%s", data)
	}

	saved, err := ImportStage(outPath)
	if err != nil {
		t.Fatalf("ImportStage() error: %v", err)
	}
	if saved.Version != project.StageVersion {
		t.Fatalf("saved version = %d, want %d", saved.Version, project.StageVersion)
	}
	if saved.Name != "Test Stage" {
		t.Fatalf("saved name = %q, want Test Stage", saved.Name)
	}
	if len(saved.Assets) != 1 || saved.Assets[0].File != assetPath {
		t.Fatalf("saved assets = %+v, want file %q", saved.Assets, assetPath)
	}
	if len(saved.Objects) != 1 || saved.Objects[0].File != objectPath {
		t.Fatalf("saved objects = %+v, want file %q", saved.Objects, objectPath)
	}
	if saved.Objects[0].Width != 50 || saved.Objects[0].Height != 50 {
		t.Fatalf("saved object size = %gx%g, want 50x50", saved.Objects[0].Width, saved.Objects[0].Height)
	}
	if len(saved.Layers) != 1 || saved.Layers[0].Name != project.DefaultStageLayer {
		t.Fatalf("saved layers = %+v, want default layer", saved.Layers)
	}
	if saved.Objects[0].Layer != project.DefaultStageLayer {
		t.Fatalf("saved object layer = %q, want %q", saved.Objects[0].Layer, project.DefaultStageLayer)
	}
}

func TestFindStageConfigPrefersStageGo(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "other.go"), []byte("package stages\n\nvar Other = Stage{}\n"), 0o644); err != nil {
		t.Fatalf("write other.go: %v", err)
	}
	stagePath := filepath.Join(dir, "stage.go")
	if err := os.WriteFile(stagePath, []byte("package stages\n\nvar FirstStage = Stage{}\n"), 0o644); err != nil {
		t.Fatalf("write stage.go: %v", err)
	}

	got, err := findStageConfig(dir)
	if err != nil {
		t.Fatalf("findStageConfig() error: %v", err)
	}
	if got != stagePath {
		t.Fatalf("findStageConfig() = %q, want %q", got, stagePath)
	}
}

func TestFindStageConfigJSONFallback(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "stage.json")
	if err := os.WriteFile(jsonPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write stage.json: %v", err)
	}

	got, err := findStageConfig(dir)
	if err != nil {
		t.Fatalf("findStageConfig() error: %v", err)
	}
	if got != jsonPath {
		t.Fatalf("findStageConfig() = %q, want %q", got, jsonPath)
	}
}

func writeTestPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("encode %s: %v", path, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close %s: %v", path, err)
	}
}

func TestStageToolTabsLayoutInOneRow(t *testing.T) {
	if got, want := stageToolTabsHeight(), stageToolTabH; got != want {
		t.Fatalf("stageToolTabsHeight() = %d, want %d", got, want)
	}
	if got, want := stageToolTabsWidth(), 3*stageToolTabW+2*stageToolTabGap; got != want {
		t.Fatalf("stageToolTabsWidth() = %d, want %d", got, want)
	}

	for i := range stageToolTabs {
		gotX, gotY := stageToolTabOrigin(i, 100, 50)
		wantX := 100 + i*(stageToolTabW+stageToolTabGap)
		if gotX != wantX || gotY != 50 {
			t.Errorf("stageToolTabOrigin(%d) = (%d,%d), want (%d,%d)", i, gotX, gotY, wantX, 50)
		}
	}

	wantIcons := []string{"hammer", "square-pen", "info"}
	for i, tab := range stageToolTabs {
		if tab.Icon != wantIcons[i] {
			t.Errorf("stageToolTabs[%d].Icon = %q, want %q", i, tab.Icon, wantIcons[i])
		}
	}
}

func TestStageBuildGridFillsAvailableWidth(t *testing.T) {
	width := 900
	cols := stageBuildColumns(width)
	if cols != len(stageBuildToggles) {
		t.Fatalf("stageBuildColumns(%d) = %d, want %d", width, cols, len(stageBuildToggles))
	}
	used := 0
	for col := 0; col < cols; col++ {
		used += stageBuildColumnWidth(width, cols, col)
	}
	used += (cols - 1) * stageToolSubGap
	if used != width {
		t.Fatalf("grid uses %d px, want %d", used, width)
	}

	narrow := 300
	cols = stageBuildColumns(narrow)
	if cols < 2 || cols >= len(stageBuildToggles) {
		t.Fatalf("stageBuildColumns(%d) = %d, want a wrapped grid", narrow, cols)
	}
	cellW := stageBuildCellWidth(narrow, cols)
	if cellW < stageToolSubMinW {
		t.Fatalf("narrow cell width = %d, want at least %d", cellW, stageToolSubMinW)
	}
}

func TestSelectStageObjectDefaultsToEdit(t *testing.T) {
	app := NewEditorApp()
	app.mode = modeStage
	app.scaleHandle = -1
	app.proj.Stage = project.NewStageData("Stage1")
	app.proj.Stage.Objects = []project.StageObject{{
		Category: project.StageCategoryBlock,
		File:     "block.png",
		X:        25,
		Y:        25,
		Width:    50,
		Height:   50,
		Opacity:  1,
	}}
	app.st.toolTab = "info"
	app.st.toolInfoAuto = true
	app.st.toolInfoHold = 10

	app.selectStageObject(0)
	if app.st.toolTab != "edit" {
		t.Fatalf("toolTab = %q after selecting, want edit", app.st.toolTab)
	}
	if app.st.toolInfoAuto {
		t.Fatal("toolInfoAuto = true after selecting, want false")
	}
}

func TestSelectStageObjectNoSelectionKeepsBuildTab(t *testing.T) {
	app := NewEditorApp()
	app.mode = modeStage
	app.st.toolTab = "build"

	app.selectStageObject(-1)
	if app.st.toolTab != "build" {
		t.Fatalf("toolTab = %q after clearing selection, want build", app.st.toolTab)
	}
}

func TestStageInfoAutoReturnsToEditAfterIdle(t *testing.T) {
	app := NewEditorApp()
	app.mode = modeStage
	app.proj.Stage = project.NewStageData("Stage1")
	app.proj.Stage.Objects = []project.StageObject{{
		Category: project.StageCategoryBlock,
		File:     "block.png",
		Width:    50,
		Height:   50,
		Opacity:  1,
	}}
	app.selectStageObject(0)

	app.markStageObjectLiveEdit()
	if app.st.toolTab != "info" || !app.st.toolInfoAuto {
		t.Fatalf("live edit tab = %q auto=%v, want info/true", app.st.toolTab, app.st.toolInfoAuto)
	}

	app.st.toolInfoHold = 1
	app.tickStageInfoAuto()
	if app.st.toolTab != "edit" || app.st.toolInfoAuto {
		t.Fatalf("idle tab = %q auto=%v, want edit/false", app.st.toolTab, app.st.toolInfoAuto)
	}
}

func TestStageInfoStaysWhilePropertyFocused(t *testing.T) {
	app := NewEditorApp()
	app.mode = modeStage
	app.proj.Stage = project.NewStageData("Stage1")
	app.proj.Stage.Objects = []project.StageObject{{
		Category: project.StageCategoryBlock,
		File:     "block.png",
		Width:    50,
		Height:   50,
		Opacity:  1,
	}}
	app.selectStageObject(0)
	app.markStageObjectLiveEdit()
	app.props[0].Focused = true
	app.st.toolInfoHold = 1

	app.tickStageInfoAuto()
	if app.st.toolTab != "info" {
		t.Fatalf("toolTab = %q while property focused, want info", app.st.toolTab)
	}
	if !app.st.toolInfoAuto {
		t.Fatal("toolInfoAuto = false while property focused, want true")
	}

	app.props[0].Focused = false
	app.tickStageInfoAuto()
	if app.st.toolTab != "edit" {
		t.Fatalf("toolTab = %q after focus lost, want edit", app.st.toolTab)
	}
}

func TestStageTabClickClearsAutomaticInfo(t *testing.T) {
	app := NewEditorApp()
	app.mode = modeStage
	app.proj.Stage = project.NewStageData("Stage1")
	app.proj.Stage.Objects = []project.StageObject{{
		Category: project.StageCategoryBlock,
		File:     "block.png",
		Width:    50,
		Height:   50,
		Opacity:  1,
	}}
	app.selectStageObject(0)
	app.markStageObjectLiveEdit()
	app.props[0].Focused = true
	app.st.toolTabRects = []stageToolRect{{Key: "edit", X: 0, Y: 0, W: 10, H: 10}}

	app.handleStageToolbarMouse(5, 5, true)
	if app.st.toolTab != "edit" || app.st.toolInfoAuto {
		t.Fatalf("tab = %q auto=%v, want edit/false", app.st.toolTab, app.st.toolInfoAuto)
	}
	if app.props[0].Focused {
		t.Fatal("property input stayed focused after tab click")
	}
}

func TestStageEditTabButtonsDeleteAndDuplicate(t *testing.T) {
	newApp := func() *EditorApp {
		app := NewEditorApp()
		app.mode = modeStage
		app.proj.Stage = project.NewStageData("Stage1")
		app.proj.Stage.Objects = []project.StageObject{{
			Category: project.StageCategoryBlock,
			File:     "block.png",
			Width:    50,
			Height:   50,
			Opacity:  1,
		}}
		app.selectStageObject(0)
		app.st.toolTab = "edit"
		app.st.deleteBtn.X, app.st.deleteBtn.Y = 100, 0
		app.st.deleteBtn.W, app.st.deleteBtn.H = 20, 20
		app.st.deleteBtn.Visible = true
		app.st.duplicateBtn.X, app.st.duplicateBtn.Y = 200, 0
		app.st.duplicateBtn.W, app.st.duplicateBtn.H = 20, 20
		app.st.duplicateBtn.Visible = true
		return app
	}

	app := newApp()
	app.handleStageToolbarMouse(105, 5, true)
	if len(app.proj.Stage.Objects) != 0 {
		t.Fatalf("objects after delete = %d, want 0", len(app.proj.Stage.Objects))
	}

	app = newApp()
	app.handleStageToolbarMouse(205, 5, true)
	if len(app.proj.Stage.Objects) != 2 {
		t.Fatalf("objects after duplicate = %d, want 2", len(app.proj.Stage.Objects))
	}
}

func TestStageBuildToggleIcons(t *testing.T) {
	wantIcons := []string{
		"palette", "move", "ban", "heart-pulse", "opacity", "toggle-right",
		"star", "refresh-cw", "square-dimensions", "circle-chevrons-right", "activity",
	}
	for i, toggle := range stageBuildToggles {
		if toggle.Icon != wantIcons[i] {
			t.Errorf("stageBuildToggles[%d].Icon = %q, want %q", i, toggle.Icon, wantIcons[i])
		}
	}
}

func TestStageToolbarSelectsTabsAndToggles(t *testing.T) {
	app := &EditorApp{}
	app.st.toolTabRects = []stageToolRect{
		{Key: "build", X: 0, Y: 0, W: 10, H: 10},
		{Key: "edit", X: 12, Y: 0, W: 10, H: 10},
		{Key: "info", X: 24, Y: 0, W: 10, H: 10},
	}
	app.st.toolToggleRects = []stageToolRect{
		{Key: "pulse", X: 40, Y: 0, W: 10, H: 10},
	}

	app.handleStageToolbarMouse(13, 5, false)
	if app.st.toolTab != "" {
		t.Fatalf("toolTab changed without a press: %q", app.st.toolTab)
	}

	app.handleStageToolbarMouse(13, 5, true)
	if app.st.toolTab != "edit" {
		t.Fatalf("toolTab = %q, want edit", app.st.toolTab)
	}

	app.handleStageToolbarMouse(45, 5, true)
	if !app.st.toolToggles["pulse"] {
		t.Fatal("pulse toggle did not turn on")
	}
	app.handleStageToolbarMouse(45, 5, true)
	if app.st.toolToggles["pulse"] {
		t.Fatal("pulse toggle did not turn off")
	}
}

func newStageObjectApp(count int) *EditorApp {
	app := NewEditorApp()
	app.mode = modeStage
	app.scaleHandle = -1
	app.proj.Stage = project.NewStageData("Stage1")
	for i := 0; i < count; i++ {
		app.proj.Stage.Objects = append(app.proj.Stage.Objects, project.StageObject{
			Category: project.StageCategoryBlock,
			File:     "block.png",
			X:        10 + float64(i)*100,
			Y:        20,
			Width:    50,
			Height:   40,
			Opacity:  1,
		})
	}
	return app
}

func TestStageMultiSelectionAddToggleAndClear(t *testing.T) {
	app := newStageObjectApp(3)
	app.selectStageObject(0)
	app.addStageObjectSelection(1)

	if got := app.stageSelectionCount(); got != 2 {
		t.Fatalf("selection count = %d, want 2", got)
	}
	if app.st.selectedObj != 1 {
		t.Fatalf("primary = %d, want 1", app.st.selectedObj)
	}

	app.toggleStageObjectSelection(0)
	if app.stageObjectSelected(0) {
		t.Fatal("object 0 still selected after toggle")
	}
	if got := app.stageSelectionCount(); got != 1 {
		t.Fatalf("selection count after toggle = %d, want 1", got)
	}

	app.toggleStageObjectSelection(1)
	if got := app.stageSelectionCount(); got != 0 {
		t.Fatalf("selection count after clearing = %d, want 0", got)
	}
	if app.st.selectedObj != -1 {
		t.Fatalf("primary = %d, want -1", app.st.selectedObj)
	}
	if app.canvas.Selection.Visible {
		t.Fatal("selection rect visible with no selection")
	}
}

func TestStageDeleteSelectedObjects(t *testing.T) {
	app := newStageObjectApp(3)
	app.selectStageObject(0)
	app.addStageObjectSelection(2)

	app.deleteSelectedStageObject()

	if len(app.proj.Stage.Objects) != 1 {
		t.Fatalf("objects after multi-delete = %d, want 1", len(app.proj.Stage.Objects))
	}
	if app.proj.Stage.Objects[0].X != 110 {
		t.Fatalf("remaining object X = %g, want 110", app.proj.Stage.Objects[0].X)
	}
	if app.stageSelectionCount() != 0 {
		t.Fatalf("selection count after delete = %d, want 0", app.stageSelectionCount())
	}
}

func TestStageDuplicateSelectedObjects(t *testing.T) {
	app := newStageObjectApp(2)
	app.st.snap = false
	app.selectStageObject(0)
	app.addStageObjectSelection(1)

	app.duplicateSelectedStageObject()

	if len(app.proj.Stage.Objects) != 4 {
		t.Fatalf("objects after multi-duplicate = %d, want 4", len(app.proj.Stage.Objects))
	}
	if got := app.stageSelectionCount(); got != 2 {
		t.Fatalf("selection count after duplicate = %d, want 2", got)
	}
	if app.st.selectedObj != 3 {
		t.Fatalf("primary after duplicate = %d, want 3", app.st.selectedObj)
	}
	for _, idx := range []int{2, 3} {
		obj := app.proj.Stage.Objects[idx]
		if obj.X != 10+float64(idx-2)*100+project.DefaultBlockSize {
			t.Fatalf("duplicate %d X = %g, want offset", idx, obj.X)
		}
	}
}

func TestStageMultiPropertyMovePreservesOffsets(t *testing.T) {
	app := newStageObjectApp(2)
	app.selectStageObject(0)
	app.addStageObjectSelection(1)
	app.props[0].Focused = true
	app.props[0].SetNumeric(120)

	app.flushStageInputs()

	if app.proj.Stage.Objects[0].X != 20 {
		t.Fatalf("object 0 X = %g, want 20", app.proj.Stage.Objects[0].X)
	}
	if app.proj.Stage.Objects[1].X != 120 {
		t.Fatalf("primary object X = %g, want 120", app.proj.Stage.Objects[1].X)
	}
}

func TestStageMultiSelectionDrawsExtraOutlines(t *testing.T) {
	app := newStageObjectApp(2)
	app.selectStageObject(0)
	app.addStageObjectSelection(1)

	app.buildStageRenders()

	if len(app.canvas.ExtraSelections) != 1 {
		t.Fatalf("extra selections = %d, want 1", len(app.canvas.ExtraSelections))
	}
	if !app.canvas.Selection.Visible {
		t.Fatal("primary selection not visible")
	}
}

func TestStageDragMovesEntireSelection(t *testing.T) {
	app := newStageObjectApp(2)
	app.selectStageObject(0)
	app.addStageObjectSelection(1)
	app.beginStageObjectDrag(0, 0)

	app.applyStageObjectDrag(app.st.dragObjStartX+25, app.st.dragObjStartY+15)

	if app.proj.Stage.Objects[0].X != 35 || app.proj.Stage.Objects[0].Y != 35 {
		t.Fatalf("object 0 = (%g,%g), want (35,35)", app.proj.Stage.Objects[0].X, app.proj.Stage.Objects[0].Y)
	}
	if app.proj.Stage.Objects[1].X != 135 || app.proj.Stage.Objects[1].Y != 35 {
		t.Fatalf("object 1 = (%g,%g), want (135,35)", app.proj.Stage.Objects[1].X, app.proj.Stage.Objects[1].Y)
	}
}

func TestStageMultiPropertyDoesNotNormalizeUneditedFields(t *testing.T) {
	app := newStageObjectApp(2)
	app.proj.Stage.Objects[1].Width = 80
	app.selectStageObject(0)
	app.addStageObjectSelection(1)
	app.props[0].Focused = true
	app.props[0].SetNumeric(20)

	app.flushStageInputs()

	if app.proj.Stage.Objects[0].Width != 50 || app.proj.Stage.Objects[1].Width != 80 {
		t.Fatalf("widths changed without width edit: %g, %g", app.proj.Stage.Objects[0].Width, app.proj.Stage.Objects[1].Width)
	}
}

func TestStageMultiPropertyWidthAppliesToAll(t *testing.T) {
	app := newStageObjectApp(2)
	app.proj.Stage.Objects[1].Width = 80
	app.selectStageObject(0)
	app.addStageObjectSelection(1)
	app.props[2].Focused = true
	app.props[2].SetNumeric(70)

	app.flushStageInputs()

	if app.proj.Stage.Objects[0].Width != 70 || app.proj.Stage.Objects[1].Width != 70 {
		t.Fatalf("widths = %g, %g, want 70,70", app.proj.Stage.Objects[0].Width, app.proj.Stage.Objects[1].Width)
	}
}

func TestStageMarqueeSelectsIntersectingObjects(t *testing.T) {
	app := newStageObjectApp(3)
	app.selectStageObject(2)

	app.beginStageMarquee(0, 0)
	if got := app.stageSelectionCount(); got != 0 {
		t.Fatalf("selection count after marquee start = %d, want 0", got)
	}

	app.updateStageMarquee(150, 50)
	if got := app.stageSelectionCount(); got != 2 {
		t.Fatalf("marquee selection count = %d, want 2", got)
	}
	if !app.stageObjectSelected(0) || !app.stageObjectSelected(1) {
		t.Fatal("marquee did not select objects 0 and 1")
	}
	if app.st.selectedObj != 0 {
		t.Fatalf("marquee primary = %d, want 0", app.st.selectedObj)
	}

	app.updateStageMarquee(60, 50)
	if got := app.stageSelectionCount(); got != 1 {
		t.Fatalf("narrow marquee count = %d, want 1", got)
	}
	if !app.stageObjectSelected(0) {
		t.Fatal("narrow marquee lost object 0")
	}

	app.finishStageMarquee()
	if app.st.marqueeActive {
		t.Fatal("marquee still active after finish")
	}
}

func TestDragStagePositionReturnsToOffGridOrigin(t *testing.T) {
	x, y := dragStagePosition(10, 20, 0, 0, 0, 0, 50, 40, true)
	if x != 10 || y != 20 {
		t.Fatalf("zero-delta drag = (%v, %v), want original (10, 20)", x, y)
	}

	x, y = dragStagePosition(10, 20, 0, 0, 50, 50, 50, 40, true)
	if x != 60 || y != 70 {
		t.Fatalf("grid delta drag = (%v, %v), want (60, 70)", x, y)
	}
}

func TestStageDragCanReturnToCenterCell(t *testing.T) {
	app := newStageObjectApp(1)
	obj := &app.proj.Stage.Objects[0]
	obj.X, obj.Y = 125, 125
	app.selectStageObject(0)

	// Move from center e to each neighboring cell and back to e.
	for _, delta := range [][2]float64{{50, 0}, {0, 50}, {-50, 0}, {0, -50}, {100, 100}} {
		app.beginStageObjectDrag(obj.X, obj.Y)
		app.applyStageObjectDrag(obj.X+delta[0], obj.Y+delta[1])
		if obj.X == 125 && obj.Y == 125 {
			t.Fatalf("object did not move for delta %v", delta)
		}
		app.beginStageObjectDrag(obj.X, obj.Y)
		app.applyStageObjectDrag(125, 125)
		if obj.X != 125 || obj.Y != 125 {
			t.Fatalf("object did not return to center: (%g,%g)", obj.X, obj.Y)
		}
	}
}

func TestStageDragReturnsToStartingCellWithinSameDrag(t *testing.T) {
	app := newStageObjectApp(1)
	obj := &app.proj.Stage.Objects[0]
	obj.X, obj.Y = 125, 125
	app.selectStageObject(0)

	app.beginStageObjectDrag(125, 125)
	app.applyStageObjectDrag(175, 125)
	if obj.X != 175 {
		t.Fatalf("object did not move away: X=%g", obj.X)
	}

	// Returning the cursor to the drag origin must restore the original cell.
	app.applyStageObjectDrag(125, 125)
	if obj.X != 125 || obj.Y != 125 {
		t.Fatalf("object did not return to starting cell within one drag: (%g,%g)", obj.X, obj.Y)
	}
}

func TestStageNormalizeLayersMigratesFlatObjects(t *testing.T) {
	stage := project.NewStageData("Legacy")
	stage.Layers = nil
	stage.Objects = []project.StageObject{
		{Category: project.StageCategoryBG, File: "bg.png"},
		{Category: project.StageCategoryBlock, File: "block.png"},
	}

	stage.NormalizeLayers()

	if len(stage.Layers) != 1 || stage.Layers[0].Name != project.DefaultStageLayer {
		t.Fatalf("layers = %+v, want one default layer", stage.Layers)
	}
	for i, obj := range stage.Objects {
		if obj.Layer != project.DefaultStageLayer {
			t.Fatalf("object %d layer = %q, want %q", i, obj.Layer, project.DefaultStageLayer)
		}
	}
}

func TestStageNormalizeLayersKeepsReferencedLayers(t *testing.T) {
	stage := project.NewStageData("Layered")
	stage.Layers = []project.StageLayer{{Name: "Back"}, {Name: "Back"}, {Name: "Front"}}
	stage.Objects = []project.StageObject{
		{Layer: "Front", File: "front.png"},
		{Layer: "Orphan", File: "orphan.png"},
		{File: "unassigned.png"},
	}

	stage.NormalizeLayers()

	if len(stage.Layers) != 3 {
		t.Fatalf("layers = %+v, want Back, Front, Orphan", stage.Layers)
	}
	if stage.Layers[0].Name != "Back" || stage.Layers[1].Name != "Front" || stage.Layers[2].Name != "Orphan" {
		t.Fatalf("layers = %+v, want Back, Front, Orphan", stage.Layers)
	}
	if stage.Objects[0].Layer != "Front" {
		t.Fatalf("object layer = %q, want Front", stage.Objects[0].Layer)
	}
	if stage.Objects[1].Layer != "Orphan" {
		t.Fatalf("orphan layer = %q, want Orphan", stage.Objects[1].Layer)
	}
	if stage.Objects[2].Layer != "Back" {
		t.Fatalf("unassigned layer = %q, want Back", stage.Objects[2].Layer)
	}
}

func TestAddStageLayerAndAssignNewObjects(t *testing.T) {
	app := newStageObjectApp(1)
	app.proj.Stage.Assets = []project.StageAsset{{
		Category: project.StageCategoryBlock,
		Name:     "block",
		File:     "block.png",
		Width:    50,
		Height:   50,
	}}
	app.proj.Stage.Objects[0].Layer = "Layer 1"
	app.st.activeLayer = "Layer 1"

	app.addStageLayer()
	if len(app.proj.Stage.Layers) != 2 {
		t.Fatalf("layers = %+v, want two layers", app.proj.Stage.Layers)
	}
	if app.st.activeLayer != "Layer 2" {
		t.Fatalf("active layer = %q, want Layer 2", app.st.activeLayer)
	}

	app.addStageObjectFromAsset(0, 100, 100)
	if got := app.proj.Stage.Objects[1].Layer; got != "Layer 2" {
		t.Fatalf("new object layer = %q, want Layer 2", got)
	}
	if app.proj.Stage.Objects[0].Layer != "Layer 1" {
		t.Fatalf("old object layer changed to %q", app.proj.Stage.Objects[0].Layer)
	}
}

func TestMoveActiveStageLayerReordersLayers(t *testing.T) {
	app := newStageObjectApp(1)
	app.addStageLayer()
	app.addStageLayer()
	if len(app.proj.Stage.Layers) != 3 || app.st.activeLayer != "Layer 3" {
		t.Fatalf("setup layers = %+v active=%q", app.proj.Stage.Layers, app.st.activeLayer)
	}

	app.moveActiveStageLayer(-1)
	want := []string{"Layer 1", "Layer 3", "Layer 2"}
	for i, name := range want {
		if app.proj.Stage.Layers[i].Name != name {
			t.Fatalf("layers = %+v, want %v", app.proj.Stage.Layers, want)
		}
	}
	if app.activeStageLayerIndex() != 1 {
		t.Fatalf("active index = %d, want 1", app.activeStageLayerIndex())
	}

	app.moveActiveStageLayer(-1)
	if app.proj.Stage.Layers[0].Name != "Layer 3" || app.activeStageLayerIndex() != 0 {
		t.Fatalf("layers = %+v active index=%d, want Layer 3 at 0", app.proj.Stage.Layers, app.activeStageLayerIndex())
	}

	app.moveActiveStageLayer(-1) // already at bottom: no-op
	if app.proj.Stage.Layers[0].Name != "Layer 3" {
		t.Fatalf("backward past bottom changed layers: %+v", app.proj.Stage.Layers)
	}
}

func TestDeleteActiveStageLayerOnlyWhenEmpty(t *testing.T) {
	app := newStageObjectApp(1)
	app.proj.Stage.Objects[0].Layer = "Layer 1"
	app.addStageLayer()
	app.addStageLayer()
	app.deleteActiveStageLayer()
	if len(app.proj.Stage.Layers) != 2 {
		t.Fatalf("layers after deleting empty layer = %+v, want 2", app.proj.Stage.Layers)
	}
	if app.st.activeLayer != "Layer 2" {
		t.Fatalf("active layer = %q, want Layer 2", app.st.activeLayer)
	}

	app.setActiveStageLayer(0)
	app.deleteActiveStageLayer()
	if len(app.proj.Stage.Layers) != 2 {
		t.Fatalf("non-empty layer deleted: %+v", app.proj.Stage.Layers)
	}
}

func TestStageClickUsesTopLayer(t *testing.T) {
	app := newStageObjectApp(2)
	app.proj.Stage.Layers = []project.StageLayer{{Name: "Back"}, {Name: "Front"}}
	app.proj.Stage.Objects[0].Layer = "Back"
	app.proj.Stage.Objects[1].Layer = "Front"
	app.proj.Stage.Objects[0].X, app.proj.Stage.Objects[0].Y = 100, 100
	app.proj.Stage.Objects[1].X, app.proj.Stage.Objects[1].Y = 100, 100
	app.proj.Stage.Objects[0].Width, app.proj.Stage.Objects[0].Height = 200, 200
	app.proj.Stage.Objects[1].Width, app.proj.Stage.Objects[1].Height = 200, 200

	if got := app.stageObjectAtWorld(100, 100); got != 1 {
		t.Fatalf("hit object = %d, want top-layer object 1", got)
	}
}

func TestStageLayerFocusDimsOtherLayers(t *testing.T) {
	app := newStageObjectApp(2)
	app.proj.Stage.Layers = []project.StageLayer{{Name: "Back"}, {Name: "Front"}}
	app.proj.Stage.Objects[0].Layer = "Back"
	app.proj.Stage.Objects[1].Layer = "Front"

	app.setActiveStageLayer(1)

	if got := app.stageObjectOpacity(&app.proj.Stage.Objects[0]); got != stageInactiveLayerOpacity {
		t.Fatalf("back opacity = %g, want %g", got, stageInactiveLayerOpacity)
	}
	if got := app.stageObjectOpacity(&app.proj.Stage.Objects[1]); got != 1 {
		t.Fatalf("front opacity = %g, want 1", got)
	}

	app.setGeneralStageView()

	for i := range app.proj.Stage.Objects {
		if got := app.stageObjectOpacity(&app.proj.Stage.Objects[i]); got != 1 {
			t.Fatalf("object %d opacity = %g in General, want 1", i, got)
		}
	}
	if app.st.activeLayer != "" {
		t.Fatalf("active layer = %q after General, want empty", app.st.activeLayer)
	}
}

func TestGeneralViewAddsToTopLayerAndStaysGeneral(t *testing.T) {
	app := newStageObjectApp(0)
	app.proj.Stage.Layers = []project.StageLayer{{Name: "Back"}, {Name: "Front"}}
	app.proj.Stage.Assets = []project.StageAsset{{
		Category: project.StageCategoryBlock,
		Name:     "block",
		File:     "block.png",
		Width:    50,
		Height:   50,
	}}
	app.st.activeLayer = ""

	app.addStageObjectFromAsset(0, 100, 100)

	if len(app.proj.Stage.Objects) != 1 {
		t.Fatalf("objects = %d, want 1", len(app.proj.Stage.Objects))
	}
	if got := app.proj.Stage.Objects[0].Layer; got != "Front" {
		t.Fatalf("new object layer = %q, want Front", got)
	}
	if app.st.activeLayer != "" {
		t.Fatalf("active layer = %q after adding in General, want General", app.st.activeLayer)
	}
}

func TestStageObjectDefaultCollisionMatchesBlock(t *testing.T) {
	block := project.StageObject{
		Category: project.StageCategoryBlock,
		Width:    50,
		Height:   40,
	}
	boxes := block.EffectiveCollisionBoxes()
	if len(boxes) != 1 {
		t.Fatalf("default collision boxes = %d, want 1", len(boxes))
	}
	if boxes[0].Width != 50 || boxes[0].Height != 40 || boxes[0].X != 0 || boxes[0].Y != 0 {
		t.Fatalf("default collision box = %+v, want 50x40 at center", boxes[0])
	}

	block.CollisionDisabled = true
	if got := block.EffectiveCollisionBoxes(); len(got) != 0 {
		t.Fatalf("disabled block collision = %+v, want none", got)
	}

	bg := project.StageObject{Category: project.StageCategoryBG, Width: 50, Height: 40}
	if got := bg.EffectiveCollisionBoxes(); len(got) != 0 {
		t.Fatalf("background collision = %+v, want none", got)
	}
}

func TestStageCollisionToggleDisablesSelection(t *testing.T) {
	app := newStageObjectApp(2)
	app.selectStageObject(0)
	app.addStageObjectSelection(1)

	app.toggleStageCollisionForSelection()
	for i := range app.proj.Stage.Objects {
		if !app.proj.Stage.Objects[i].CollisionDisabled {
			t.Fatalf("object %d collision still enabled", i)
		}
	}

	app.toggleStageCollisionForSelection()
	for i := range app.proj.Stage.Objects {
		if app.proj.Stage.Objects[i].CollisionDisabled {
			t.Fatalf("object %d collision still disabled", i)
		}
	}
}

func TestStageCollisionEditorSeedsDefaultAndManagesHitboxes(t *testing.T) {
	app := newStageObjectApp(1)
	app.selectStageObject(0)
	app.openStageCollisionEditor()

	if !app.st.collisionEdit || !app.st.showCollisions {
		t.Fatalf("collision editor state edit=%v show=%v", app.st.collisionEdit, app.st.showCollisions)
	}
	obj := &app.proj.Stage.Objects[0]
	if len(obj.CollisionBoxes) != 1 {
		t.Fatalf("seeded collision boxes = %d, want 1", len(obj.CollisionBoxes))
	}
	if obj.CollisionBoxes[0].Width != 50 || obj.CollisionBoxes[0].Height != 40 {
		t.Fatalf("seeded collision = %+v, want 50x40", obj.CollisionBoxes[0])
	}

	app.setStageCollisionTab(1)
	app.addStageCollisionBox()
	if len(obj.Hitboxes) != 1 {
		t.Fatalf("hitboxes after add = %d, want 1", len(obj.Hitboxes))
	}
	app.removeStageCollisionBox()
	if len(obj.Hitboxes) != 0 {
		t.Fatalf("hitboxes after remove = %d, want 0", len(obj.Hitboxes))
	}

	app.setStageCollisionTab(0)
	if app.st.collisionTab != "collision" {
		t.Fatalf("collision tab = %q, want collision", app.st.collisionTab)
	}
}

func TestStageBoxRenderTransformsWithObjectRotation(t *testing.T) {
	obj := &project.StageObject{X: 100, Y: 50, Rotation: 90}
	box := project.StageBox{X: 10, Y: 0, Width: 20, Height: 10, Rotation: 30}

	got := stageBoxRender(obj, box, 0, true)
	if math.Abs(got.OffsetX-100) > 1e-9 || math.Abs(got.OffsetY-60) > 1e-9 {
		t.Fatalf("box offset = (%g, %g), want (100, 60)", got.OffsetX, got.OffsetY)
	}
	if math.Abs(got.Rotation-math.Pi*2/3) > 1e-9 {
		t.Fatalf("box rotation = %g, want %g", got.Rotation, math.Pi*2/3)
	}
	if got.WorldWidth != 20 || got.WorldHeight != 10 {
		t.Fatalf("box size = %gx%g, want 20x10", got.WorldWidth, got.WorldHeight)
	}
	if !got.Selected {
		t.Fatal("box selected flag lost")
	}
}

func TestExportImportStageCollisionBoxes(t *testing.T) {
	dir := t.TempDir()
	stage := project.NewStageData("Collision Stage")
	stage.Objects = []project.StageObject{{
		Category:          project.StageCategoryBlock,
		File:              filepath.Join(dir, "block.png"),
		X:                 25,
		Y:                 25,
		Width:             50,
		Height:            50,
		Opacity:           1,
		CollisionDisabled: true,
		CollisionBoxes:    []project.StageBox{{X: 2, Y: 3, Width: 20, Height: 10, Rotation: 15}},
		Hitboxes:          []project.StageBox{{X: 1, Y: 1, Width: 8, Height: 6}},
	}}

	outPath := filepath.Join(dir, "stage.go")
	if err := ExportStage(outPath, stage); err != nil {
		t.Fatalf("ExportStage() error: %v", err)
	}
	saved, err := ImportStage(outPath)
	if err != nil {
		t.Fatalf("ImportStage() error: %v", err)
	}
	if len(saved.Objects) != 1 {
		t.Fatalf("saved objects = %d, want 1", len(saved.Objects))
	}
	obj := saved.Objects[0]
	if !obj.CollisionDisabled {
		t.Fatal("CollisionDisabled lost in round trip")
	}
	if len(obj.CollisionBoxes) != 1 || obj.CollisionBoxes[0].Width != 20 || obj.CollisionBoxes[0].Rotation != 15 {
		t.Fatalf("collision boxes = %+v, want one 20-wide box rotated 15", obj.CollisionBoxes)
	}
	if len(obj.Hitboxes) != 1 || obj.Hitboxes[0].Width != 8 || obj.Hitboxes[0].Height != 6 {
		t.Fatalf("hitboxes = %+v, want one 8x6 box", obj.Hitboxes)
	}
}

func TestStageCollisionRendersOnlyWhenEnabled(t *testing.T) {
	app := newStageObjectApp(1)
	app.selectStageObject(0)

	if got := app.stageCollisionRenders(); len(got) != 0 {
		t.Fatalf("collision renders while hidden = %d, want 0", len(got))
	}

	app.st.showCollisions = true
	if got := app.stageCollisionRenders(); len(got) != 1 {
		t.Fatalf("collision renders while shown = %d, want 1", len(got))
	}

	app.proj.Stage.Objects[0].CollisionDisabled = true
	if got := app.stageCollisionRenders(); len(got) != 0 {
		t.Fatalf("collision renders while disabled = %d, want 0", len(got))
	}
}

func TestStageObjectDeepCopyKeepsCollisionSlicesIndependent(t *testing.T) {
	obj := project.StageObject{
		CollisionBoxes: []project.StageBox{{X: 1, Width: 20}},
		Hitboxes:       []project.StageBox{{X: 2, Width: 8}},
	}
	cp := obj.DeepCopy()
	cp.CollisionBoxes[0].X = 99
	cp.Hitboxes[0].X = 88

	if obj.CollisionBoxes[0].X != 1 || obj.Hitboxes[0].X != 2 {
		t.Fatalf("original mutated: collision=%+v hitbox=%+v", obj.CollisionBoxes, obj.Hitboxes)
	}
}

func TestScanStageAssetsUsesFolderCategory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "props")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create props dir: %v", err)
	}
	writeTestPNG(t, filepath.Join(dir, "table.png"), 120, 80)
	writeTestPNG(t, filepath.Join(dir, "wide_bg.png"), 40, 20)

	assets, err := scanStageAssets(dir)
	if err != nil {
		t.Fatalf("scanStageAssets() error: %v", err)
	}
	if len(assets) != 2 {
		t.Fatalf("assets = %d, want 2", len(assets))
	}
	byName := map[string]project.StageAsset{}
	for _, asset := range assets {
		byName[asset.Name] = asset
	}
	if got := byName["table"].Category; got != project.StageCategoryProp {
		t.Fatalf("table category = %q, want prop", got)
	}
	if got := byName["wide_bg"].Category; got != project.StageCategoryBG {
		t.Fatalf("wide_bg category = %q, want bg", got)
	}
}

func TestStagePropKeepsImageSizeAndStartsWithoutCollision(t *testing.T) {
	app := newStageObjectApp(0)
	app.proj.Stage.Assets = []project.StageAsset{{
		Category: project.StageCategoryProp,
		Name:     "table",
		File:     "table.png",
		Width:    120,
		Height:   80,
	}}

	app.addStageObjectFromAsset(0, 100, 100)

	if len(app.proj.Stage.Objects) != 1 {
		t.Fatalf("objects = %d, want 1", len(app.proj.Stage.Objects))
	}
	obj := &app.proj.Stage.Objects[0]
	if obj.Width != 120 || obj.Height != 80 {
		t.Fatalf("prop size = %gx%g, want 120x80", obj.Width, obj.Height)
	}
	if !obj.CollisionDisabled {
		t.Fatal("new prop collision should start disabled")
	}
	if got := obj.EffectiveCollisionBoxes(); len(got) != 0 {
		t.Fatalf("new prop collision = %+v, want none", got)
	}
}

func TestStagePropCanEnableCollisionManually(t *testing.T) {
	app := newStageObjectApp(0)
	app.proj.Stage.Objects = append(app.proj.Stage.Objects, project.StageObject{
		Category:          project.StageCategoryProp,
		File:              "table.png",
		X:                 100,
		Y:                 100,
		Width:             120,
		Height:            80,
		Opacity:           1,
		CollisionDisabled: true,
	})
	app.selectStageObject(0)

	app.toggleStageCollisionForSelection()
	obj := &app.proj.Stage.Objects[0]
	if obj.CollisionDisabled {
		t.Fatal("prop collision still disabled after toggle")
	}
	if len(obj.CollisionBoxes) != 1 {
		t.Fatalf("prop collision boxes = %d, want seeded 1", len(obj.CollisionBoxes))
	}
	if obj.CollisionBoxes[0].Width != 120 || obj.CollisionBoxes[0].Height != 80 {
		t.Fatalf("seeded prop box = %+v, want 120x80", obj.CollisionBoxes[0])
	}
	if got := obj.EffectiveCollisionBoxes(); len(got) != 1 {
		t.Fatalf("effective prop collision = %+v, want 1", got)
	}

	app.toggleStageCollisionForSelection()
	if !obj.CollisionDisabled {
		t.Fatal("prop collision should be disabled after second toggle")
	}
	if got := obj.EffectiveCollisionBoxes(); len(got) != 0 {
		t.Fatalf("disabled prop collision = %+v, want none", got)
	}
}

func TestStagePropCollisionEditorDoesNotSeedUntilAdded(t *testing.T) {
	app := newStageObjectApp(0)
	app.proj.Stage.Objects = append(app.proj.Stage.Objects, project.StageObject{
		Category:          project.StageCategoryProp,
		File:              "table.png",
		Width:             120,
		Height:            80,
		Opacity:           1,
		CollisionDisabled: true,
	})
	app.selectStageObject(0)
	app.openStageCollisionEditor()

	obj := &app.proj.Stage.Objects[0]
	if len(obj.CollisionBoxes) != 0 {
		t.Fatalf("opening prop editor seeded %+v, want none", obj.CollisionBoxes)
	}

	app.addStageCollisionBox()
	if len(obj.CollisionBoxes) != 1 {
		t.Fatalf("prop collision after add = %d, want 1", len(obj.CollisionBoxes))
	}
	if obj.CollisionDisabled {
		t.Fatal("adding a prop collision box should enable collision")
	}

	app.removeStageCollisionBox()
	if len(obj.CollisionBoxes) != 0 {
		t.Fatalf("prop collision after remove = %+v, want none", obj.CollisionBoxes)
	}
	if !obj.CollisionDisabled {
		t.Fatal("removing the last prop box should disable collision again")
	}
}

func TestStageClickPrefersBlockThenPropThenBackground(t *testing.T) {
	app := newStageObjectApp(0)
	app.proj.Stage.Objects = []project.StageObject{
		{Category: project.StageCategoryBG, File: "bg.png", X: 50, Y: 50, Width: 200, Height: 200, Opacity: 1},
		{Category: project.StageCategoryProp, File: "prop.png", X: 50, Y: 50, Width: 100, Height: 100, Opacity: 1},
		{Category: project.StageCategoryBlock, File: "block.png", X: 50, Y: 50, Width: 50, Height: 50, Opacity: 1},
	}

	if got := app.stageObjectAtWorld(50, 50); got != 2 {
		t.Fatalf("hit = %d, want block 2", got)
	}

	app.proj.Stage.Objects = app.proj.Stage.Objects[:2]
	if got := app.stageObjectAtWorld(50, 50); got != 1 {
		t.Fatalf("hit = %d, want prop 1", got)
	}

	app.proj.Stage.Objects = app.proj.Stage.Objects[:1]
	if got := app.stageObjectAtWorld(50, 50); got != 0 {
		t.Fatalf("hit = %d, want background 0", got)
	}
}
