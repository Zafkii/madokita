package editor

import (
	"fmt"
	"log"

	assetIcons "animprite/assets/icons"
	"animprite/internal/canvas"
	"animprite/internal/filedialog"
	animfonts "animprite/internal/fonts"
	"animprite/internal/icons"
	"animprite/internal/project"
	"animprite/internal/theme"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

var ErrWindowClose = fmt.Errorf("window close requested")

type windowState struct {
	outsideWidth                                               int
	outsideHeight                                              int
	resetViewBtnX, resetViewBtnY, resetViewBtnW, resetViewBtnH int
}

type previewState struct {
	speedSlider    *ui.Slider
	previewPlaying bool
	loopChecked    bool
	previewSpeed   float64

	previewAccumulator float64
	previewAnimIdx     int
	armedElapsed       float64

	previewBtnH                                        int
	previewRow2Y                                       int
	previewChkX, previewChkY, previewChkW, previewChkH int
	previewMinusX, previewMinusY                       int
	previewPlusX, previewPlusY                         int
	previewPlayX, previewPlayY                         int

	previewPlayHovered  bool
	previewLoopHovered  bool
	previewMinusHovered bool
	previewPlusHovered  bool
}

type rightPanelState struct {
	scroll   int
	contentH int
	buf      *ebiten.Image
}

type EditorApp struct {
	mode editorMode

	canvas *canvas.Canvas
	th     *theme.Manager

	modeDropdown *ui.Dropdown
	themeBtn     *ui.Button
	openBtn      *ui.Button
	openDirBtn   *ui.Button
	saveBtn      *ui.Button

	props               [5]*ui.TextInput
	phaseDropdown       *ui.Dropdown
	frameSpriteDropdown *ui.Dropdown
	originInputs        [2]*ui.TextInput
	baseRotInput        *ui.TextInput
	atkTimingInputs     [5]*ui.TextInput
	fpsInput            *ui.TextInput

	loopInput *ui.TextInput

	loadedSprites map[int]*ebiten.Image

	animTable    *ui.Table
	spriteTable  *ui.Table
	hurtboxTable *ui.Table
	hitboxTable  *ui.Table

	animNameInputs []*ui.TextInput

	animAddFrameBtns     []*ui.Button
	animRemoveFrameBtns  []*ui.Button
	animFramePrevBtns    []*ui.Button
	animFrameNextBtns    []*ui.Button
	animFrameInputs      []*ui.TextInput
	spriteBrowseBtns     []*ui.Button
	spriteFramePrevBtns  []*ui.Button
	spriteFrameNextBtns  []*ui.Button
	spriteWidthInputs    []*ui.TextInput
	spriteHeightInputs   []*ui.TextInput
	hurtboxWidthInputs   []*ui.TextInput
	hurtboxHeightInputs  []*ui.TextInput
	hurtboxDmgMultInputs []*ui.TextInput
	hitboxWidthInputs    []*ui.TextInput
	hitboxHeightInputs   []*ui.TextInput
	hurtboxXInputs       []*ui.TextInput
	hurtboxYInputs       []*ui.TextInput

	topPanelH                int
	prevMouseX, prevMouseY   int
	prevSelectedAnimIdx      int
	prevSelectedAnimFrameIdx int
	prevSelectedSpriteIdx    int
	prevSelectedHurtboxIdx   int
	spriteEditIdx            int
	hurtboxAnimCtx           int
	panelMode                rightPanelMode
	statusMsg                string
	statusTime               int
	hoveredFilePath          string

	win  windowState
	proj project.ProjectData
	prev previewState
	rp   rightPanelState

	scaleHandle  int
	scaleOrig    scaleOrigData
	rotationDrag rotationDragData

	st      stageState
	iconSet *icons.Set

	currentFilePath string

	dirty         bool
	confirm       confirmState
	confirmYesBtn *ui.Button
	confirmNoBtn  *ui.Button

	undoStack    []project.ProjectData
	redoStack    []project.ProjectData
	wheelChanged bool
}

func (a *EditorApp) Layout(w, h int) (int, int) {
	// Keep the two-column top panel at the width where its fixed controls
	// fit. If the window manager ignores the size limits, the logical screen
	// is scaled down instead of letting the columns overlap.
	if w < MinWinW {
		w = MinWinW
	}
	if h < MinWinH {
		h = MinWinH
	}
	if w != a.win.outsideWidth || h != a.win.outsideHeight {
		a.win.outsideWidth = w
		a.win.outsideHeight = h
		a.syncLayout()
	}
	return w, h
}

func NewEditorApp() *EditorApp {
	if err := ui.SetFontFromTTF(animfonts.NotoSansRegularTTF, 13); err != nil {
		panic("font load: " + err.Error())
	}
	th := theme.NewManager()

	app := &EditorApp{
		th:                       th,
		mode:                     modeMovement,
		topPanelH:                topPanelMinH,
		prevSelectedSpriteIdx:    -1,
		prevSelectedAnimIdx:      -1,
		prevSelectedAnimFrameIdx: -1,
		spriteEditIdx:            0,
		loadedSprites:            make(map[int]*ebiten.Image),
		st: stageState{
			images:      make(map[string]*ebiten.Image),
			expanded:    map[string]bool{project.StageCategoryBG: true, project.StageCategoryBlock: true},
			toolTab:     "build",
			toolToggles: make(map[string]bool),
			snap:        true,
			selectedObj: -1,
			dragAsset:   -1,
		},
	}

	iconSet, err := icons.Load(assetIcons.FS, icons.RasterSize)
	if err != nil {
		log.Printf("load editor icons: %v", err)
	}
	app.iconSet = iconSet

	app.win.outsideWidth = DefaultWinW
	app.win.outsideHeight = DefaultWinH

	app.confirmYesBtn = ui.NewButton(0, 0, 130, 26, "Discard & Open", th)
	app.confirmYesBtn.BtnType = ui.BtnRed
	app.confirmYesBtn.OnClick = app.acceptConfirm
	app.confirmNoBtn = ui.NewButton(0, 0, 130, 26, "Cancel", th)
	app.confirmNoBtn.BtnType = ui.BtnBlue
	app.confirmNoBtn.OnClick = app.cancelConfirm

	canvasTop := modeIndicatorH + app.topPanelH
	app.canvas = canvas.New(
		0,
		canvasTop,
		DefaultWinW-rightPanelW,
		DefaultWinH-canvasTop-statusbarH,
		th,
	)

	groupH := dropdownH + btnGap + rightBtnH + btnGap + rightBtnH + btnGap + rightBtnH + btnGap + rightBtnH
	groupTop := modeIndicatorH + (app.topPanelH-groupH)/2
	app.modeDropdown = ui.NewDropdown(panelPad, groupTop, dropdownW, dropdownH, th)
	app.modeDropdown.Options = []string{"Movement Editor", "Attack Editor", "Stage Editor"}
	app.modeDropdown.DisplayText = "Movement Editor"
	app.modeDropdown.OnChange = func(idx int) {
		switch idx {
		case 1:
			app.setEditorMode(modeAttack)
		case 2:
			app.setEditorMode(modeStage)
		default:
			app.setEditorMode(modeMovement)
		}
	}

	btnY0 := groupTop + dropdownH + btnGap
	btnY1 := btnY0 + rightBtnH + btnGap
	btnY2 := btnY1 + rightBtnH + btnGap
	btnY3 := btnY2 + rightBtnH + btnGap
	app.openBtn = ui.NewButton(panelPad, btnY0, rightBtnW, rightBtnH, "Open", th)
	app.openBtn.BtnType = ui.BtnOrange
	app.openBtn.Tooltip = "Open project file"
	app.openDirBtn = ui.NewButton(panelPad, btnY1, rightBtnW, rightBtnH, "Open Dir", th)
	app.openDirBtn.BtnType = ui.BtnOrange
	app.openDirBtn.Tooltip = "Open project folder"
	app.saveBtn = ui.NewButton(panelPad, btnY2, rightBtnW, rightBtnH, "Save", th)
	app.saveBtn.Tooltip = "Save project"
	app.themeBtn = ui.NewButton(panelPad, btnY3, rightBtnW, rightBtnH, "Dark", th)
	app.themeBtn.Tooltip = "Toggle dark or light theme"

	app.themeBtn.OnClick = func() {
		app.th.Toggle()
		if app.th.IsLight {
			app.themeBtn.Text = "Light"
		} else {
			app.themeBtn.Text = "Dark"
		}
	}

	app.saveBtn.OnClick = func() {
		title := "Save Movement"
		filter := "Go Files\000*.go\000All Files\000*.*"
		if app.mode == modeAttack {
			title = "Save Attack"
		} else if app.mode == modeStage {
			title = "Save Stage"
			filter = "Go Files\000*.go\000JSON Files\000*.json\000All Files\000*.*"
		}
		path, err := filedialog.SaveFile(title, filter)
		if err != nil {
			app.setStatus("Save cancelled")
			return
		}
		app.flushInputsToData()
		var saveErr error
		switch app.mode {
		case modeAttack:
			saveErr = app.saveAttackFile(path)
		case modeStage:
			saveErr = app.saveStageFile(path)
		default:
			saveErr = app.saveMovementFile(path)
		}
		if saveErr != nil {
			app.setStatus("Save error: " + saveErr.Error())
			return
		}
		app.currentFilePath = path
		app.markClean()
		app.setStatus("Saved: " + path)
	}

	app.openBtn.OnClick = func() {
		path, err := filedialog.OpenFile("Open File", "Go Files\000*.go\000JSON Files\000*.json\000All Files\000*.*")
		if err != nil {
			app.setStatus("Open cancelled")
			return
		}
		app.requestOpenFile(path)
	}

	app.openDirBtn.OnClick = func() {
		dir, err := filedialog.OpenFolder("Open Project Folder")
		if err != nil {
			app.setStatus("Open directory cancelled")
			return
		}
		app.requestOpenFolder(dir)
	}

	app.proj.AssetName = "MyMovement"
	app.proj.AssetKey = "my_movement"
	app.proj.DefaultOriginX = 0.5
	app.proj.DefaultOriginY = 0.5
	app.proj.Sprites = []project.SpriteRow{
		{Name: "base", File: "base.png", Width: 256, Height: 256, FrameCount: 1, CurrentIdx: 0, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
	}
	makeEntry := func() []project.FrameSpriteEntry {
		return []project.FrameSpriteEntry{
			{SpriteIdx: 0, OriginX: 0.5, OriginY: 0.5, ScaleX: 1, ScaleY: 1},
		}
	}
	app.proj.Animations = []project.AnimationRow{
		{
			Name: "idle", CurrentIdx: 0, Loop: true,
			Frames: []project.AnimationFrame{
				{Sprites: makeEntry(), Phase: project.PhaseWindup},
				{Sprites: makeEntry(), Phase: project.PhaseWindup},
			},
			FPS: 14,
		},
	}

	app.initTables()
	app.initRightPanelWidgets()
	app.initStageWidgets()
	app.syncLayout()
	app.navigateToAnim(0)

	app.prevSelectedHurtboxIdx = -1

	app.markClean()
	return app
}

func (a *EditorApp) initRightPanelWidgets() {
	px := DefaultWinW - rightPanelW + rightPanelPad
	iw := rightPanelInner
	th := a.th

	newNum := func(label string, def, min, max, step float64) *ui.TextInput {
		inp := ui.NewTextInput(px, 0, iw, rightPanelInputH, th)
		inp.SetLabel(label)
		inp.SetNumeric(def)
		inp.Numeric = true
		inp.Min = min
		inp.Max = max
		inp.Step = step
		return inp
	}

	a.props[0] = newNum("Offset X", 0, -99999, 99999, 1)
	a.props[1] = newNum("Offset Y", 0, -99999, 99999, 1)
	a.props[2] = newNum("Rotation (°)", 0, -360, 360, 0.5)
	a.props[3] = newNum("Scale X", 1, -99999, 99999, 0.05)
	a.props[4] = newNum("Scale Y", 1, -99999, 99999, 0.05)

	a.phaseDropdown = ui.NewDropdown(px, 0, 70, rightPanelInputH, th)
	a.phaseDropdown.Options = []string{"WU", "AC", "RC", "ARM"}
	a.phaseDropdown.Selected = 0
	a.phaseDropdown.SetLabel("Phase")
	a.phaseDropdown.OnChange = func(idx int) {
		animIdx := a.animTable.SelectedIdx
		if animIdx < 0 || animIdx >= len(a.proj.Animations) {
			return
		}
		anim := &a.proj.Animations[animIdx]
		startFrame := anim.CurrentIdx
		if startFrame < 0 {
			return
		}
		a.dirty = true
		for fi := startFrame; fi < len(anim.Frames); fi++ {
			anim.Frames[fi].Phase = project.FramePhase(idx)
		}
	}

	a.frameSpriteDropdown = ui.NewDropdown(px, 0, rightPanelInner, rightPanelInputH, th)
	a.frameSpriteDropdown.SetLabel("Frame Sprite")
	a.frameSpriteDropdown.Selected = -1
	a.frameSpriteDropdown.Options = []string{"(none)"}
	a.frameSpriteDropdown.OnChange = func(idx int) {
		animIdx := a.animTable.SelectedIdx
		if animIdx < 0 || animIdx >= len(a.proj.Animations) {
			return
		}
		frameIdx := a.proj.Animations[animIdx].CurrentIdx
		if frameIdx < 0 || frameIdx >= len(a.proj.Animations[animIdx].Frames) {
			return
		}
		frame := &a.proj.Animations[animIdx].Frames[frameIdx]
		spriteIdx := idx - 1
		if spriteIdx < 0 {
			a.spriteEditIdx = 0
			return
		}
		a.spriteEditIdx = spriteIdx
		a.ensureFrameEntry(frame, spriteIdx)
		a.loadAnimFrameProps(animIdx, frameIdx)
	}

	a.originInputs[0] = newNum("Origin X", 0.5, 0, 1, 0.01)
	a.originInputs[1] = newNum("Origin Y", 0.5, 0, 1, 0.01)
	a.baseRotInput = newNum("Base Rot (°)", 0, -360, 360, 0.5)

	a.atkTimingInputs[0] = newNum("Windup", 200, 0, 9999, 1)
	a.atkTimingInputs[1] = newNum("Active", 250, 0, 9999, 1)
	a.atkTimingInputs[2] = newNum("Recover", 800, 0, 9999, 1)
	a.atkTimingInputs[3] = newNum("Armed", 3000, 0, 9999, 1)
	a.atkTimingInputs[4] = newNum("Armed FPS", 14, 1, 120, 1)

	a.fpsInput = newNum("FPS", 14, 1, 120, 1)

	a.loopInput = ui.NewTextInput(px, 0, iw, rightPanelInputH, th)
	a.loopInput.SetLabel("Loop")
	a.loopInput.Text = "true"

	a.prev.speedSlider = ui.NewSlider(0, 0, 80, 16, th)
	a.prev.speedSlider.Min = 0.1
	a.prev.speedSlider.Max = 5.0
	a.prev.speedSlider.Value = 1.0
	a.prev.speedSlider.Step = 0.1
	a.prev.speedSlider.Visible = true

	a.prev.speedSlider.OnChange = func(v float64) {
		a.prev.previewSpeed = v
	}

	a.prev.previewSpeed = 1.0
	a.prev.loopChecked = true
}

func (a *EditorApp) initStageWidgets() {
	px := DefaultWinW - rightPanelW + rightPanelPad
	iw := rightPanelInner
	th := a.th

	opacity := ui.NewTextInput(px, 0, iw, rightPanelInputH, th)
	opacity.SetLabel("Opacity")
	opacity.Numeric = true
	opacity.Min = 0
	opacity.Max = 1
	opacity.HasMax = true
	opacity.Step = 0.05
	opacity.SetNumeric(1)
	a.st.opacityInput = opacity

	a.st.deleteBtn = ui.NewButton(0, 0, 1, 22, "Delete", th)
	a.st.deleteBtn.BtnType = ui.BtnRed
	a.st.deleteBtn.OnClick = func() {
		a.deleteSelectedStageObject()
	}

	a.st.duplicateBtn = ui.NewButton(0, 0, 1, 22, "Duplicate", th)
	a.st.duplicateBtn.BtnType = ui.BtnBlue
	a.st.duplicateBtn.OnClick = func() {
		a.duplicateSelectedStageObject()
	}
}
