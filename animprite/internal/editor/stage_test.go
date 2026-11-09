package editor

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
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

func TestStageToolbarSelectsTabsAndToggles(t *testing.T) {
	app := &EditorApp{}
	app.st.toolTabRects = []stageToolRect{
		{Key: "build", X: 0, Y: 0, W: 10, H: 10},
		{Key: "edit", X: 12, Y: 0, W: 10, H: 10},
		{Key: "delete", X: 24, Y: 0, W: 10, H: 10},
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
