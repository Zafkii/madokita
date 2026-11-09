package editor

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"animprite/internal/project"
)

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestModeFromFolderName(t *testing.T) {
	tests := []struct {
		name string
		want editorMode
		ok   bool
	}{
		{name: "sayaka_miki_atk", want: modeAttack, ok: true},
		{name: "sayaka_miki_moves", want: modeMovement, ok: true},
		{name: "sayaka_miki_move", want: modeMovement, ok: true},
		{name: "Character_ATK", want: modeAttack, ok: true},
		{name: "test_stage", want: modeStage, ok: true},
		{name: "Test_STAGE", want: modeStage, ok: true},
		{name: "stage", want: modeStage, ok: true},
		{name: "sayaka_miki", want: modeMovement, ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := modeFromFolderName(tt.name)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("mode = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenFolderDoesNotInheritStageMode(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sayaka_miki_atk")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	a := newScaleTestApp()
	a.mode = modeStage

	if err := a.openFolder(dir); err != nil {
		t.Fatal(err)
	}
	if a.mode != modeAttack {
		t.Fatalf("mode = %v, want modeAttack", a.mode)
	}
}

func TestNeedsFrameCountRecalc(t *testing.T) {
	sprites := []project.SpriteRow{
		{Width: 256, Height: 256, FrameCount: 1},
		{File: "sayaka.png", Width: 256, Height: 256, FrameCount: 25},
	}

	tests := []struct {
		name       string
		assignment spriteAssignment
		want       bool
	}{
		{name: "new row", assignment: spriteAssignment{RowIdx: -1}, want: true},
		{name: "imported placeholder row", assignment: spriteAssignment{RowIdx: 0}, want: true},
		{name: "declared spritesheet", assignment: spriteAssignment{RowIdx: 1}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsFrameCountRecalc(sprites, tt.assignment); got != tt.want {
				t.Fatalf("needsFrameCountRecalc = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportedMovementWithoutSpritesNeedsFrameCountRecalc(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "moves.go")
	writeFile(t, path, []byte(`package movements

import . "madokita/internal/animation"

var TestMovement = Movement{
	AssetKey: "test",
	Animations: map[string]MovementAnimDef{
		"idle": Anim(3, true,
			F(S(0, 0, 0, 0, 0, 1, 1, 0.5, 0.5)),
			F(S(0, 1, 0, 0, 0, 1, 1, 0.5, 0.5)),
		),
	},
}
`))

	proj, err := ImportMovement(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(proj.Sprites) != 1 {
		t.Fatalf("sprites = %d, want 1", len(proj.Sprites))
	}
	row := proj.Sprites[0]
	if row.File != "" || row.FrameCount != 1 {
		t.Fatalf("imported placeholder row = %+v, want empty File and FrameCount 1", row)
	}

	// The folder opener must map the convention PNG onto that placeholder row
	// and derive its frame count from the image instead of keeping FrameCount 1.
	writeFile(t, filepath.Join(dir, "base.png"), []byte("not decoded by the planner"))
	assignments := planFolderSprites(dir, proj.Sprites, true)
	if len(assignments) != 1 {
		t.Fatalf("assignments = %d, want 1", len(assignments))
	}
	if assignments[0].RowIdx != 0 {
		t.Fatalf("assignment row = %d, want 0", assignments[0].RowIdx)
	}
	if !needsFrameCountRecalc(proj.Sprites, assignments[0]) {
		t.Fatal("imported placeholder row must recalculate its frame count from the PNG")
	}
}

func TestPlanFolderSpritesUsesConventionForStalePaths(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "base.png"), nil)
	writeFile(t, filepath.Join(dir, "sprite2.png"), nil)

	sprites := []project.SpriteRow{
		{File: "/old/project/vertical_attack.png", Width: 256, Height: 256, FrameCount: 20},
		{File: "/old/project/test_weapon.png", Width: 145, Height: 145, FrameCount: 6},
	}

	got := planFolderSprites(dir, sprites, false)
	if len(got) != 2 {
		t.Fatalf("assignments = %d, want 2", len(got))
	}
	if got[0].RowIdx != 0 || got[0].File != "base.png" {
		t.Fatalf("row 0 = %+v, want base.png", got[0])
	}
	if got[1].RowIdx != 1 || got[1].File != "sprite2.png" {
		t.Fatalf("row 1 = %+v, want sprite2.png", got[1])
	}
	if got[0].Path != filepath.Join(dir, "base.png") {
		t.Fatalf("path = %q, want absolute path inside dir", got[0].Path)
	}

	// The planner must not recalculate imported frame metadata.
	if sprites[0].FrameCount != 20 || sprites[1].Width != 145 || sprites[1].Height != 145 {
		t.Fatalf("planner mutated imported sprite rows: %+v", sprites)
	}
}

func TestPlanFolderSpritesResolvesBasenameFirst(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "base.png"), nil)
	writeFile(t, filepath.Join(dir, "weapon.png"), nil)

	sprites := []project.SpriteRow{
		{File: "/old/project/weapon.png", Width: 128, Height: 128, FrameCount: 4},
	}

	got := planFolderSprites(dir, sprites, false)
	if len(got) != 1 {
		t.Fatalf("assignments = %d, want 1", len(got))
	}
	if got[0].File != "weapon.png" {
		t.Fatalf("resolved file = %q, want weapon.png (basename before convention)", got[0].File)
	}
}

func TestPlanFolderSpritesResolvesRecordedRelativePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sheets"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "sheets", "weapon.png"), nil)

	sprites := []project.SpriteRow{{File: "sheets/weapon.png"}}

	got := planFolderSprites(dir, sprites, false)
	if len(got) != 1 {
		t.Fatalf("assignments = %d, want 1", len(got))
	}
	if got[0].File != "sheets/weapon.png" {
		t.Fatalf("resolved file = %q, want sheets/weapon.png", got[0].File)
	}
}

func TestPlanFolderSpritesAppendsExtras(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "base.png"), nil)
	writeFile(t, filepath.Join(dir, "sprite2.png"), nil)
	writeFile(t, filepath.Join(dir, "weapon.png"), nil)

	sprites := []project.SpriteRow{
		{File: "/old/project/a.png"},
		{File: "/old/project/b.png"},
	}

	got := planFolderSprites(dir, sprites, true)
	if len(got) != 3 {
		t.Fatalf("assignments = %d, want 3", len(got))
	}
	extra := got[2]
	if extra.RowIdx != -1 {
		t.Fatalf("extra row index = %d, want -1", extra.RowIdx)
	}
	if extra.File != "weapon.png" {
		t.Fatalf("extra file = %q, want weapon.png", extra.File)
	}
	if extra.Name != spriteDefaultName(2) {
		t.Fatalf("extra name = %q, want %q", extra.Name, spriteDefaultName(2))
	}

	noExtras := planFolderSprites(dir, sprites, false)
	if len(noExtras) != 2 {
		t.Fatalf("without extras assignments = %d, want 2", len(noExtras))
	}
}

func TestFindProjectGo(t *testing.T) {
	t.Run("no go files", func(t *testing.T) {
		got, err := findProjectGo(t.TempDir(), modeMovement)
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Fatalf("path = %q, want empty", got)
		}
	})

	t.Run("single file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "custom.go")
		writeFile(t, path, []byte("package custom\n"))

		got, err := findProjectGo(dir, modeMovement)
		if err != nil {
			t.Fatal(err)
		}
		if got != path {
			t.Fatalf("path = %q, want %q", got, path)
		}
	})

	t.Run("preferred attack file", func(t *testing.T) {
		dir := t.TempDir()
		preferred := filepath.Join(dir, "atk.go")
		writeFile(t, preferred, []byte("package attacks\n"))
		writeFile(t, filepath.Join(dir, "other.go"), []byte("package other\n"))

		got, err := findProjectGo(dir, modeAttack)
		if err != nil {
			t.Fatal(err)
		}
		if got != preferred {
			t.Fatalf("path = %q, want %q", got, preferred)
		}
	})

	t.Run("preferred movement file", func(t *testing.T) {
		dir := t.TempDir()
		preferred := filepath.Join(dir, "moves.go")
		writeFile(t, preferred, []byte("package movements\n"))
		writeFile(t, filepath.Join(dir, "other.go"), []byte("package other\n"))

		got, err := findProjectGo(dir, modeMovement)
		if err != nil {
			t.Fatal(err)
		}
		if got != preferred {
			t.Fatalf("path = %q, want %q", got, preferred)
		}
	})

	t.Run("multiple files by content", func(t *testing.T) {
		dir := t.TempDir()
		attackPath := filepath.Join(dir, "aaa.go")
		movementPath := filepath.Join(dir, "bbb.go")
		writeFile(t, attackPath, []byte("package attacks\nvar X = Attack{}\n"))
		writeFile(t, movementPath, []byte("package movements\nvar Y = Movement{}\n"))

		got, err := findProjectGo(dir, modeAttack)
		if err != nil {
			t.Fatal(err)
		}
		if got != attackPath {
			t.Fatalf("attack path = %q, want %q", got, attackPath)
		}

		got, err = findProjectGo(dir, modeMovement)
		if err != nil {
			t.Fatal(err)
		}
		if got != movementPath {
			t.Fatalf("movement path = %q, want %q", got, movementPath)
		}
	})

	t.Run("multiple files without matching declaration", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "one.go"), []byte("package one\n"))
		writeFile(t, filepath.Join(dir, "two.go"), []byte("package two\n"))

		if _, err := findProjectGo(dir, modeMovement); err == nil {
			t.Fatal("expected an error for ambiguous .go files, got nil")
		}
	})
}

func TestRelativeSpritePath(t *testing.T) {
	dir := t.TempDir()

	if got := relativeSpritePath(dir, filepath.Join(dir, "base.png")); got != "base.png" {
		t.Fatalf("inside path = %q, want base.png", got)
	}
	if got := relativeSpritePath(dir, filepath.Join(dir, "nested", "sprite.png")); got != "nested/sprite.png" {
		t.Fatalf("nested path = %q, want nested/sprite.png", got)
	}
	if got := relativeSpritePath(dir, "base.png"); got != "base.png" {
		t.Fatalf("relative path = %q, want base.png", got)
	}
	outside := filepath.Join(filepath.Dir(dir), "elsewhere", "sprite.png")
	if got := relativeSpritePath(dir, outside); got != outside {
		t.Fatalf("outside path = %q, want unchanged %q", got, outside)
	}
	if got := relativeSpritePath(dir, ""); got != "" {
		t.Fatalf("empty path = %q, want empty", got)
	}
}

func TestExportRelativizesSpritePathsUnderOutputDir(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(filepath.Dir(dir), "elsewhere", "weapon.png")
	proj := &project.ProjectData{
		AssetName:      "Test",
		AssetKey:       "test",
		DefaultOriginX: 0.5,
		DefaultOriginY: 0.5,
		Sprites: []project.SpriteRow{
			{File: filepath.Join(dir, "base.png"), Width: 256, Height: 256, FrameCount: 1},
			{File: outside, Width: 128, Height: 128, FrameCount: 1},
		},
	}

	tests := []struct {
		name   string
		export func(string, *project.ProjectData) error
	}{
		{name: "movement", export: ExportMovement},
		{name: "attack", export: ExportAttack},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, tt.name+".go")
			if err := tt.export(path, proj); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(data, []byte(`File: "base.png"`)) {
				t.Fatalf("export did not relativize sprite inside output dir:\n%s", data)
			}
			if !bytes.Contains(data, []byte("File: "+strconv.Quote(outside))) {
				t.Fatalf("export changed sprite outside output dir:\n%s", data)
			}
		})
	}
}
