package editor

import (
	"fmt"
	"time"

	"animprite/internal/canvas"
	"animprite/internal/filedialog"
	animfonts "animprite/internal/fonts"
	"animprite/internal/project"
	"animprite/internal/theme"
	"animprite/internal/ui"
	"animprite/internal/windrag"

	"github.com/hajimehoshi/ebiten/v2"
)

var ErrWindowClose = fmt.Errorf("window close requested")

type windowState struct {
	dragMgr            *windrag.DragManager
	prevLeftBtn        bool
	titleBarImg        *ebiten.Image
	titleBarW          int
	clickTimer         time.Time
	lastClickMX        int
	lastClickMY        int
	dragPending        bool
	outsideWidth       int
	outsideHeight      int
	barLogicH          int
	btnLogicW          int
	resizing           resizeInfo
	prevW, prevH       int
	restoreState       int
	restoreW, restoreH int
	hoveredBtn         titleBarBtn

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

	canvas    *canvas.Canvas
	th        *theme.Manager
	titleLogo *ebiten.Image

	modeDropdown *ui.Dropdown
	themeBtn     *ui.Button
	openBtn      *ui.Button
	saveBtn      *ui.Button

	props               [5]*ui.TextInput
	phaseDropdown       *ui.Dropdown
	frameSpriteDropdown *ui.Dropdown
	originInputs        [2]*ui.TextInput
	baseRotInput        *ui.TextInput
	atkTimingInputs     [5]*ui.TextInput
	fpsInput            *ui.TextInput

	loopInput          *ui.TextInput

	loadedSprites map[int]*ebiten.Image

	animTable    *ui.Table
	spriteTable  *ui.Table
	hurtboxTable *ui.Table
	hitboxTable  *ui.Table

	animNameInputs      []*ui.TextInput

	animAddFrameBtns    []*ui.Button
	animRemoveFrameBtns []*ui.Button
	animFramePrevBtns   []*ui.Button
	animFrameNextBtns   []*ui.Button
	animFrameInputs     []*ui.TextInput
	spriteBrowseBtns    []*ui.Button
	spriteFramePrevBtns []*ui.Button
	spriteFrameNextBtns []*ui.Button
	spriteWidthInputs   []*ui.TextInput
	spriteHeightInputs  []*ui.TextInput
	hurtboxWidthInputs  []*ui.TextInput
	hurtboxHeightInputs []*ui.TextInput
	hurtboxDmgMultInputs    []*ui.TextInput
	hitboxWidthInputs   []*ui.TextInput
	hitboxHeightInputs  []*ui.TextInput
	hurtboxXInputs      []*ui.TextInput
	hurtboxYInputs      []*ui.TextInput

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

	scaleHandle int
	scaleOrig   scaleOrigData

	undoStack    []project.ProjectData
	redoStack    []project.ProjectData
	wheelChanged bool
}

func (a *EditorApp) Layout(w, h int) (int, int) {
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
		win: windowState{
			dragMgr: &windrag.DragManager{},
		},
	}

	app.win.outsideWidth = DefaultWinW
	app.win.outsideHeight = DefaultWinH

	canvasTop := titleBarH + modeIndicatorH + app.topPanelH
	app.canvas = canvas.New(
		0,
		canvasTop,
		DefaultWinW-rightPanelW,
		DefaultWinH-canvasTop-statusbarH,
		th,
	)

	groupH := dropdownH + btnGap + rightBtnH + btnGap + rightBtnH + btnGap + rightBtnH
	groupTop := titleBarH + modeIndicatorH + (app.topPanelH-groupH)/2
	app.modeDropdown = ui.NewDropdown(panelPad, groupTop, dropdownW, dropdownH, th)
	app.modeDropdown.Options = []string{"Movement Editor", "Attack Editor"}
	app.modeDropdown.DisplayText = "Movement Editor"
	app.modeDropdown.OnChange = func(idx int) {
		if idx == 0 {
			app.mode = modeMovement
			app.modeDropdown.DisplayText = "Movement Editor"
		} else {
			app.mode = modeAttack
			app.modeDropdown.DisplayText = "Attack Editor"
		}
	}

	btnY0 := groupTop + dropdownH + btnGap
	btnY1 := btnY0 + rightBtnH + btnGap
	btnY2 := btnY1 + rightBtnH + btnGap
	app.openBtn = ui.NewButton(panelPad, btnY0, rightBtnW, rightBtnH, "Open", th)
	app.openBtn.BtnType = ui.BtnOrange
	app.saveBtn = ui.NewButton(panelPad, btnY1, rightBtnW, rightBtnH, "Save", th)
	app.themeBtn = ui.NewButton(panelPad, btnY2, rightBtnW, rightBtnH, "Dark", th)

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
		if app.mode == modeAttack {
			title = "Save Attack"
		}
		path, err := filedialog.SaveFile(title, "Go Files\000*.go\000All Files\000*.*")
		if err != nil {
			app.setStatus("Save cancelled")
			return
		}
		app.flushInputsToData()
		var saveErr error
		if app.mode == modeAttack {
			saveErr = app.saveAttackFile(path)
		} else {
			saveErr = app.saveMovementFile(path)
		}
		if saveErr != nil {
			app.setStatus("Save error: " + saveErr.Error())
			return
		}
		app.setStatus("Saved: " + path)
	}

	app.openBtn.OnClick = func() {
		title := "Open Movement"
		if app.mode == modeAttack {
			title = "Open Attack"
		}
		path, err := filedialog.OpenFile(title, "Go Files\000*.go\000All Files\000*.*")
		if err != nil {
			app.setStatus("Open cancelled")
			return
		}
		var openErr error
		if app.mode == modeAttack {
			openErr = app.openAttackFile(path)
		} else {
			openErr = app.openMovementFile(path)
		}
		if openErr != nil {
			app.setStatus("Open error: " + openErr.Error())
			return
		}
		app.setStatus("Opened: " + path)
	}

	app.proj.AssetName = "MyMovement"
	app.proj.AssetKey = "my_movement"
	app.proj.DefaultOriginX = 0.5
	app.proj.DefaultOriginY = 0.5
	app.proj.Sprites = []project.SpriteRow{
		{Name: "Base", File: "base.png", Width: 256, Height: 256, FrameCount: 1, CurrentIdx: 0, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
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
	app.syncLayout()
	app.navigateToAnim(0)

	app.prevSelectedHurtboxIdx = -1

	if logoImg, err := LoadICO("assets/logo.ico", 16); err == nil && logoImg != nil {
		app.titleLogo = ebiten.NewImageFromImage(logoImg)
	}

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
		if a.frameSpriteEntry(frame, spriteIdx) == nil {
			frame.Sprites = append(frame.Sprites, project.FrameSpriteEntry{
				SpriteIdx:      spriteIdx,
				SpriteFrameIdx: 0,
				OffsetX:        a.proj.Sprites[spriteIdx].OffsetX,
				OffsetY:        a.proj.Sprites[spriteIdx].OffsetY,
				Rotation:       a.proj.Sprites[spriteIdx].Rotation,
				OriginX:        a.proj.Sprites[spriteIdx].OriginX,
				OriginY:        a.proj.Sprites[spriteIdx].OriginY,
				ScaleX:         a.proj.Sprites[spriteIdx].ScaleX,
				ScaleY:         a.proj.Sprites[spriteIdx].ScaleY,
			})
		}
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
