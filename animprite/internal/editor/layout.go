package editor

const (
	DefaultWinW = 1280
	DefaultWinH = 720
)

// The top panel is a two-column table grid. Below this width the fixed
// controls inside each table start colliding, so the window is clamped here
// instead of letting the columns overlap.
const (
	MinWinW = 720
	MinWinH = 480
)

const (
	modeIndicatorH = 22

	topPanelMinH = 136
	topPanelPad  = 6
	topTableRowH = 22
	topTableGap  = 4
	topPanelBtnH = 18

	panelPad    = 8
	tableColPad = 4
	dropdownW   = 140
	dropdownH   = 22

	// Fixed column widths keep row controls inside their table when the
	// window is near MinWinW. Flexible columns absorb the remaining space.
	animFrameColW   = 160
	spriteNameColW  = 40
	spriteNumColW   = 34
	spriteFrameColW = 84

	rightBtnW = 60
	rightBtnH = 20
	btnGap    = 8

	rightPanelW      = 280
	rightPanelPad    = 8
	rightPanelInputH = 20
	rightPanelInner  = rightPanelW - rightPanelPad*2

	statusbarH = 28

	scrollStep   = 30
	scrollMargin = 4
)
