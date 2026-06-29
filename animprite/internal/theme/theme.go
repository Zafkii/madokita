package theme

import "image/color"

type Palette struct {
	// Canvas
	CanvasBG             color.Color
	CanvasGridLine       color.Color
	CanvasGridAxis       color.Color
	CanvasBoundaryStroke color.Color

	// Panels
	TopPanelBG   color.Color
	RightPanelBG color.Color
	StatusbarBG  color.Color

	// Text
	TextPrimary  color.Color
	TextMuted    color.Color
	LabelColor   color.Color
	ButtonText   color.Color
	BtnTextHover color.Color

	// Buttons
	BtnGreen    color.Color
	BtnRed      color.Color
	BtnBlue     color.Color
	BtnOrange   color.Color
	BtnDisabled color.Color
	BtnHover    color.Color

	// Dropdown
	DropdownBG                color.Color
	DropdownBorder            color.Color
	DropdownText              color.Color
	DropdownPopupBG           color.Color
	DropdownPopupBorder       color.Color
	DropdownPopupText         color.Color
	DropdownPopupHover        color.Color
	DropdownPopupSelectedBG   color.Color
	DropdownPopupSelectedText color.Color

	// Inputs
	InputBG          color.Color
	InputBorder      color.Color
	InputFocusBorder color.Color
	SelectionBG      color.Color
	SelectionText    color.Color

	// UI chrome
	Border         color.Color
	SelectedRow    color.Color
	ModeIndBG      color.Color
	TableBG        color.Color
	ScrollbarTrack color.Color
	ScrollbarThumb color.Color
}

var Dark = Palette{
	// Canvas
	CanvasBG:             color.RGBA{10, 10, 20, 255},
	CanvasGridLine:       color.RGBA{90, 190, 140, 50},
	CanvasGridAxis:       color.RGBA{62, 119, 144, 255},
	CanvasBoundaryStroke: color.RGBA{40, 90, 110, 255},

	// Panels
	TopPanelBG:   color.RGBA{22, 26, 48, 255},
	RightPanelBG: color.RGBA{24, 28, 50, 255},
	StatusbarBG:  color.RGBA{19, 20, 42, 255},

	// Text
	TextPrimary:  color.RGBA{212, 212, 212, 255},
	TextMuted:    color.RGBA{130, 140, 160, 255},
	LabelColor:   color.RGBA{255, 218, 85, 255},
	ButtonText:   color.RGBA{235, 235, 245, 255},
	BtnTextHover: color.RGBA{255, 212, 40, 255},

	// Buttons
	BtnGreen:    color.RGBA{21, 114, 44, 255},
	BtnRed:      color.RGBA{144, 40, 40, 255},
	BtnBlue:     color.RGBA{47, 96, 120, 255},
	BtnOrange:   color.RGBA{210, 135, 50, 255},
	BtnDisabled: color.RGBA{68, 68, 85, 255},
	BtnHover:    color.RGBA{60, 70, 110, 255},

	// Dropdown
	DropdownBG:                color.RGBA{35, 40, 70, 255},
	DropdownBorder:            color.RGBA{65, 65, 85, 255},
	DropdownText:              color.RGBA{255, 255, 255, 255},
	DropdownPopupBG:           color.RGBA{42, 48, 80, 255},
	DropdownPopupBorder:       color.RGBA{65, 65, 85, 255},
	DropdownPopupText:         color.RGBA{255, 255, 255, 255},
	DropdownPopupHover:        color.RGBA{55, 68, 110, 255},
	DropdownPopupSelectedBG:   color.RGBA{255, 212, 40, 255},
	DropdownPopupSelectedText: color.RGBA{30, 30, 45, 255},

	// Inputs
	InputBG:          color.RGBA{26, 26, 46, 255},
	InputBorder:      color.RGBA{65, 65, 85, 255},
	InputFocusBorder: color.RGBA{102, 170, 255, 255},
	SelectionBG:      color.RGBA{255, 230, 0, 102},
	SelectionText:    color.RGBA{0, 0, 0, 255},

	// UI chrome
	Border:         color.RGBA{65, 65, 85, 255},
	SelectedRow:    color.RGBA{80, 150, 255, 50},
	ModeIndBG:      color.RGBA{15, 16, 35, 255},
	TableBG:        color.RGBA{18, 18, 35, 255},
	ScrollbarTrack: color.RGBA{40, 40, 55, 200},
	ScrollbarThumb: color.RGBA{100, 100, 130, 255},
}

var Light = Palette{
	// Canvas
	CanvasBG:             color.RGBA{200, 215, 225, 255},
	CanvasGridLine:       color.RGBA{180, 190, 200, 255},
	CanvasGridAxis:       color.RGBA{100, 140, 180, 255},
	CanvasBoundaryStroke: color.RGBA{120, 140, 160, 255},

	// Panels
	TopPanelBG:   color.RGBA{220, 225, 235, 255},
	RightPanelBG: color.RGBA{215, 220, 230, 255},
	StatusbarBG:  color.RGBA{200, 208, 220, 255},

	// Text
	TextPrimary:  color.RGBA{30, 35, 50, 255},
	TextMuted:    color.RGBA{130, 140, 155, 255},
	LabelColor:   color.RGBA{60, 100, 140, 255},
	ButtonText:   color.RGBA{255, 255, 255, 255},
	BtnTextHover: color.RGBA{255, 255, 255, 255},

	// Buttons
	BtnGreen:    color.RGBA{50, 160, 80, 255},
	BtnRed:      color.RGBA{200, 55, 55, 255},
	BtnBlue:     color.RGBA{60, 130, 170, 255},
	BtnOrange:   color.RGBA{200, 140, 50, 255},
	BtnDisabled: color.RGBA{190, 190, 195, 255},
	BtnHover:    color.RGBA{0, 188, 182, 216},

	// Dropdown
	DropdownBG:                color.RGBA{177, 211, 255, 255},
	DropdownBorder:            color.RGBA{200, 205, 210, 255},
	DropdownText:              color.RGBA{30, 35, 50, 255},
	DropdownPopupBG:           color.RGBA{235, 238, 245, 255},
	DropdownPopupBorder:       color.RGBA{200, 205, 210, 255},
	DropdownPopupText:         color.RGBA{30, 35, 50, 255},
	DropdownPopupHover:        color.RGBA{0, 188, 182, 216},
	DropdownPopupSelectedBG:   color.RGBA{0, 129, 203, 255},
	DropdownPopupSelectedText: color.RGBA{30, 35, 50, 255},

	// Inputs
	InputBG:          color.RGBA{245, 245, 250, 255},
	InputBorder:      color.RGBA{200, 205, 210, 255},
	InputFocusBorder: color.RGBA{100, 160, 255, 255},
	SelectionBG:      color.RGBA{100, 160, 255, 100},
	SelectionText:    color.RGBA{0, 0, 0, 255},

	// UI chrome
	Border:         color.RGBA{195, 200, 210, 255},
	SelectedRow:    color.RGBA{100, 160, 255, 50},
	ModeIndBG:      color.RGBA{208, 215, 225, 255},
	TableBG:        color.RGBA{230, 235, 245, 255},
	ScrollbarTrack: color.RGBA{200, 200, 220, 200},
	ScrollbarThumb: color.RGBA{150, 150, 180, 255},
}

type Manager struct {
	Current Palette
	Dark    Palette
	Light   Palette
	IsLight bool
}

func NewManager() *Manager {
	return &Manager{Current: Dark, Dark: Dark, Light: Light}
}

func (m *Manager) Toggle() {
	m.IsLight = !m.IsLight
	if m.IsLight {
		m.Current = m.Light
	} else {
		m.Current = m.Dark
	}
}
