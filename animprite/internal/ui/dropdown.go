package ui

import (
	"animprite/internal/theme"

	"github.com/hajimehoshi/ebiten/v2"
)

type Dropdown struct {
	X, Y, W, H    int
	Label         string
	Options       []string
	Selected      int
	Visible       bool
	Enabled       bool
	OnChange      func(int)
	th            *theme.Manager
	open          bool
	labelW        int
	hoveredIdx    int
	buttonHovered bool
	DisplayText   string
}

func NewDropdown(x, y, w, h int, th *theme.Manager) *Dropdown {
	return &Dropdown{
		X: x, Y: y, W: w, H: h,
		Visible: true,
		Enabled: true,
		th:      th,
	}
}

func (d *Dropdown) SetLabel(l string) {
	d.Label = l
	d.labelW = TextWidth(l) + 4
}

func (d *Dropdown) HandleMouse(cx, cy int, justPressed bool) {
	if !d.Visible || !d.Enabled {
		d.buttonHovered = false
		return
	}
	x := d.X + d.labelW
	d.hoveredIdx = -1
	d.buttonHovered = cx >= x && cx <= x+d.W && cy >= d.Y && cy <= d.Y+d.H

	if d.open {
		for i := range d.Options {
			oy := d.Y + d.H + i*d.H
			if cx >= x && cx <= x+d.W && cy >= oy && cy <= oy+d.H {
				d.hoveredIdx = i
				break
			}
		}
	}

	if justPressed && cx >= x && cx <= x+d.W && cy >= d.Y && cy <= d.Y+d.H {
		d.open = !d.open
		return
	}
	if d.open && justPressed {
		for i := range d.Options {
			oy := d.Y + d.H + i*d.H
			if cx >= x && cx <= x+d.W && cy >= oy && cy <= oy+d.H {
				d.Selected = i
				d.open = false
				if d.OnChange != nil {
					d.OnChange(i)
				}
				return
			}
		}
		d.open = false
	}
}

func (d *Dropdown) Draw(screen *ebiten.Image) {
	if !d.Visible {
		return
	}
	p := d.th.Current
	x := d.X + d.labelW
	if d.Label != "" {
		DrawText(screen, d.Label, d.X, d.Y+2, 1, p.LabelColor)
	}
	FillRect(screen, x, d.Y, d.W, d.H, p.DropdownBG)
	FillBorder(screen, x, d.Y, d.W, d.H, p.DropdownBorder)

	if d.buttonHovered && d.Enabled {
		FillRect(screen, x, d.Y, d.W, d.H, p.DropdownHoverBG)
	}

	text := d.DisplayText
	if text == "" && d.Selected >= 0 && d.Selected < len(d.Options) {
		text = d.Options[d.Selected]
	}
	d.drawButtonText(screen, p, x, text)

	arrowClr := p.DropdownText
	if d.buttonHovered && d.Enabled {
		arrowClr = p.BtnTextHover
	}
	DrawText(screen, "▼", x+d.W-12, d.Y+2, 1, arrowClr)
}

func (d *Dropdown) drawButtonText(screen *ebiten.Image, p theme.Palette, x int, text string) {
	textClr := p.DropdownText
	if d.buttonHovered && d.Enabled {
		textClr = p.BtnTextHover
	}
	lines := splitLines(text)
	if len(lines) == 1 {
		DrawText(screen, lines[0], x+4, d.Y+2, 1, textClr)
		return
	}
	lineH := TextHeight()
	totalH := len(lines) * lineH
	topPad := (d.H - totalH) / 2
	if topPad < 0 {
		topPad = 0
	}
	for i, line := range lines {
		DrawText(screen, line, x+4, d.Y+topPad+i*lineH, 1, textClr)
	}
}

func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

func (d *Dropdown) IsOpen() bool {
	return d.open
}

func (d *Dropdown) PopupHit(cx, cy int) bool {
	if !d.open {
		return false
	}
	x := d.X + d.labelW
	popupH := len(d.Options) * d.H
	return cx >= x && cx <= x+d.W && cy >= d.Y+d.H && cy < d.Y+d.H+popupH
}

func (d *Dropdown) DrawPopup(screen *ebiten.Image) {
	if !d.Visible || !d.open {
		return
	}
	p := d.th.Current
	x := d.X + d.labelW
	for i := range d.Options {
		oy := d.Y + d.H + i*d.H
		bg := p.DropdownPopupBG
		textClr := p.DropdownPopupText
		if i == d.Selected {
			bg = p.DropdownPopupSelectedBG
			textClr = p.DropdownPopupSelectedText
		}
		FillRect(screen, x, oy, d.W, d.H, bg)
		if i == d.hoveredIdx && i != d.Selected {
			FillRect(screen, x, oy, d.W, d.H, p.DropdownPopupHover)
		}
		FillBorder(screen, x, oy, d.W, d.H, p.DropdownPopupBorder)
		DrawText(screen, d.Options[i], x+4, oy+2, 1, textClr)
	}
}
