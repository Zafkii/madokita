package ui

import (
	"animprite/internal/theme"
	"image/color"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type TextInput struct {
	X, Y, W, H    int
	Text          string
	Focused       bool
	Visible       bool
	Enabled       bool
	OnChange      func(string)
	OnEnter       func()
	Label         string
	Placeholder   string
	Numeric       bool
	HasMax        bool
	Step          float64
	Min, Max      float64
	th            *theme.Manager
	labelW        int
	ignoreNext    bool
	cursorTick    int
	cursorPos     int
	selStart      int
	clickCount    int
	lastClickTime time.Time
	lastClickCX   int
	lastClickCY   int
}

func NewTextInput(x, y, w, h int, th *theme.Manager) *TextInput {
	return &TextInput{
		X: x, Y: y, W: w, H: h,
		Visible:  true,
		Enabled:  true,
		th:       th,
		Step:     1,
		selStart: -1,
	}
}

func (t *TextInput) SetLabel(l string) {
	t.Label = l
	t.labelW = TextWidth(l) + 6
}

func (t *TextInput) LabelWidth() int {
	return t.labelW
}

func (t *TextInput) InputX() int {
	return t.X + t.labelW
}

func (t *TextInput) InputW() int {
	return t.W - t.labelW
}

func (t *TextInput) HandleMouse(cx, cy int, justPressed bool) {
	if !t.Visible || !t.Enabled || !justPressed {
		return
	}
	ix := t.InputX()
	over := cx >= ix && cx <= ix+t.InputW() && cy >= t.Y && cy <= t.Y+t.H

	if over {
		if !t.Focused {
			t.Focused = true
			if t.Numeric {
				v, err := strconv.ParseFloat(t.Text, 64)
				if err == nil {
					t.Text = strconv.FormatFloat(v, 'f', -1, 64)
				}
			}
		}

		now := time.Now()
		dcx := cx - t.lastClickCX
		dcy := cy - t.lastClickCY
		isMulti := !t.lastClickTime.IsZero() && now.Sub(t.lastClickTime) < 400 &&
			dcx >= -4 && dcx <= 4 && dcy >= -4 && dcy <= 4
		t.lastClickTime = now
		t.lastClickCX = cx
		t.lastClickCY = cy

		if isMulti {
			t.clickCount++
			if t.clickCount > 3 {
				t.clickCount = 1
			}
		} else {
			t.clickCount = 1
		}

		switch t.clickCount {
		case 1:
			t.cursorPos = t.cursorAtX(cx - ix - 4)
			t.selStart = t.cursorPos
		case 2:
			t.selectWord()
		case 3:
			t.selStart = 0
			t.cursorPos = len(t.Text)
			t.clickCount = 1
		}
		t.cursorTick = 0
	} else {
		t.Focused = false
	}
}

func (t *TextInput) HandleRunes(runes []rune) {
	if !t.Focused || !t.Visible || !t.Enabled || len(runes) == 0 {
		return
	}
	changed := false
	for _, r := range runes {
		if unicode.IsPrint(r) {
			if t.Numeric && !unicode.IsDigit(r) && r != '.' && r != '-' {
				continue
			}
			if t.selStart >= 0 {
				t.deleteSelection()
			}
			s := string(r)
			t.Text = t.Text[:t.cursorPos] + s + t.Text[t.cursorPos:]
			t.cursorPos += len(s)
			changed = true
		}
	}
	if changed && t.OnChange != nil {
		t.OnChange(t.Text)
	}
}

func (t *TextInput) HandleKeys() {
	if !t.Focused || !t.Visible || !t.Enabled {
		return
	}

	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	changed := false

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
		if t.selStart >= 0 {
			t.deleteSelection()
			changed = true
		} else if t.cursorPos > 0 {
			_, size := utf8.DecodeLastRuneInString(t.Text[:t.cursorPos])
			t.Text = t.Text[:t.cursorPos-size] + t.Text[t.cursorPos:]
			t.cursorPos -= size
			changed = true
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyDelete):
		if t.selStart >= 0 {
			t.deleteSelection()
			changed = true
		} else if t.cursorPos < len(t.Text) {
			_, size := utf8.DecodeRuneInString(t.Text[t.cursorPos:])
			t.Text = t.Text[:t.cursorPos] + t.Text[t.cursorPos+size:]
			changed = true
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyLeft):
		if shift {
			if t.selStart < 0 {
				t.selStart = t.cursorPos
			}
		} else {
			t.selStart = -1
		}
		if t.cursorPos > 0 {
			_, size := utf8.DecodeLastRuneInString(t.Text[:t.cursorPos])
			t.cursorPos -= size
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyRight):
		if shift {
			if t.selStart < 0 {
				t.selStart = t.cursorPos
			}
		} else {
			t.selStart = -1
		}
		if t.cursorPos < len(t.Text) {
			_, size := utf8.DecodeRuneInString(t.Text[t.cursorPos:])
			t.cursorPos += size
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyHome):
		if shift {
			if t.selStart < 0 {
				t.selStart = t.cursorPos
			}
		} else {
			t.selStart = -1
		}
		t.cursorPos = 0

	case inpututil.IsKeyJustPressed(ebiten.KeyEnd):
		if shift {
			if t.selStart < 0 {
				t.selStart = t.cursorPos
			}
		} else {
			t.selStart = -1
		}
		t.cursorPos = len(t.Text)

	case ctrl && inpututil.IsKeyJustPressed(ebiten.KeyA):
		t.selStart = 0
		t.cursorPos = len(t.Text)

	case ctrl && inpututil.IsKeyJustPressed(ebiten.KeyC):
		if t.selStart >= 0 {
			ClipboardSet(t.selectedText())
		}

	case ctrl && inpututil.IsKeyJustPressed(ebiten.KeyX):
		if t.selStart >= 0 {
			ClipboardSet(t.selectedText())
			t.deleteSelection()
			changed = true
		}

	case ctrl && inpututil.IsKeyJustPressed(ebiten.KeyV):
		if text, ok := ClipboardGet(); ok {
			if t.selStart >= 0 {
				t.deleteSelection()
			}
			t.Text = t.Text[:t.cursorPos] + text + t.Text[t.cursorPos:]
			t.cursorPos += len(text)
			changed = true
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyEnter):
		t.HandleEnter()
	}

	if changed && t.OnChange != nil {
		t.OnChange(t.Text)
	}
}

func (t *TextInput) HandleEnter() {
	if !t.Focused {
		return
	}
	t.Focused = false
	t.selStart = -1
	if t.Numeric {
		t.applyClamp()
	}
	if t.OnChange != nil {
		t.OnChange(t.Text)
	}
	if t.OnEnter != nil {
		t.OnEnter()
	}
}

func (t *TextInput) decimalPlaces() int {
	if t.Step >= 1 {
		return 0
	}
	s := strconv.FormatFloat(t.Step, 'f', -1, 64)
	if idx := strings.IndexByte(s, '.'); idx >= 0 {
		return len(s) - idx - 1
	}
	return 0
}

func (t *TextInput) applyClamp() {
	v, err := strconv.ParseFloat(t.Text, 64)
	if err != nil {
		return
	}
	clamped := false
	if v < t.Min {
		v = t.Min
		clamped = true
	}
	if t.HasMax && v > t.Max {
		v = t.Max
		clamped = true
	}
	if clamped {
		t.Text = strconv.FormatFloat(v, 'f', t.decimalPlaces(), 64)
	}
}

func (t *TextInput) SetNumeric(v float64) {
	t.Text = strconv.FormatFloat(v, 'f', t.decimalPlaces(), 64)
}

func (t *TextInput) NumericValue() float64 {
	v, _ := strconv.ParseFloat(t.Text, 64)
	return v
}

func (t *TextInput) IntValue() int {
	return int(t.NumericValue())
}

func (t *TextInput) cursorAtX(padX int) int {
	if padX <= 0 || t.Text == "" {
		return 0
	}
	var prevPos, prevW int
	for i := 0; i < len(t.Text); {
		w := TextWidth(t.Text[:i]) + 4
		if w >= padX {
			if i > 0 && padX-prevW < w-padX {
				return prevPos
			}
			return i
		}
		prevPos = i
		prevW = w
		_, size := utf8.DecodeRuneInString(t.Text[i:])
		i += size
	}
	// After last rune — check end
	endW := TextWidth(t.Text) + 4
	if padX < endW && padX-prevW < endW-padX {
		return prevPos
	}
	return len(t.Text)
}

func (t *TextInput) selectWord() {
	if t.Text == "" {
		return
	}
	pos := t.cursorPos
	if pos >= len(t.Text) {
		if pos == 0 {
			return
		}
		_, size := utf8.DecodeLastRuneInString(t.Text[:pos])
		pos -= size
	}
	r, _ := utf8.DecodeRuneInString(t.Text[pos:])
	if unicode.IsSpace(r) {
		end := pos
		for end < len(t.Text) {
			r, size := utf8.DecodeRuneInString(t.Text[end:])
			if !unicode.IsSpace(r) {
				break
			}
			end += size
		}
		if end >= len(t.Text) {
			t.selStart = pos
			t.cursorPos = len(t.Text)
			return
		}
		pos = end
	}
	start := pos
	for start > 0 {
		_, size := utf8.DecodeLastRuneInString(t.Text[:start])
		r, _ := utf8.DecodeRuneInString(t.Text[start-size : start])
		if unicode.IsSpace(r) {
			break
		}
		start -= size
	}
	end := pos
	for end < len(t.Text) {
		r, size := utf8.DecodeRuneInString(t.Text[end:])
		if unicode.IsSpace(r) {
			break
		}
		end += size
	}
	t.selStart = start
	t.cursorPos = end
}

func (t *TextInput) HandleDrag(cx, cy int) {
	if !t.Visible || !t.Enabled || !t.Focused || t.selStart < 0 {
		return
	}
	ix := t.InputX()
	over := cx >= ix && cx <= ix+t.InputW() && cy >= t.Y && cy <= t.Y+t.H
	if !over {
		return
	}
	dx := cx - t.lastClickCX
	if dx < 0 {
		dx = -dx
	}
	if dx <= 3 {
		return
	}
	t.cursorPos = t.cursorAtX(cx - ix - 4)
}

func (t *TextInput) deleteSelection() {
	if t.selStart < 0 {
		return
	}
	start := t.selStart
	end := t.cursorPos
	if start > end {
		start, end = end, start
	}
	t.Text = t.Text[:start] + t.Text[end:]
	t.cursorPos = start
	t.selStart = -1
}

func (t *TextInput) selectedText() string {
	if t.selStart < 0 {
		return ""
	}
	start := t.selStart
	end := t.cursorPos
	if start > end {
		start, end = end, start
	}
	return t.Text[start:end]
}

func (t *TextInput) Draw(screen *ebiten.Image) {
	if !t.Visible {
		return
	}
	p := t.th.Current
	ix := t.InputX()
	if t.Label != "" {
		DrawText(screen, t.Label, t.X, t.Y+2, 1, p.LabelColor)
	}
	iw := t.InputW()
	FillRect(screen, ix, t.Y, iw, t.H, p.InputBG)
	borderClr := p.InputBorder
	if t.Focused {
		borderClr = p.InputFocusBorder
	}
	FillBorder(screen, ix, t.Y, iw, t.H, borderClr)

	textX := ix + 4
	textY := t.Y + 2

	var selStart, selEnd int
	hasSel := t.Focused && t.selStart >= 0
	if hasSel {
		selStart = t.selStart
		selEnd = t.cursorPos
		if selStart > selEnd {
			selStart, selEnd = selEnd, selStart
		}
		sx := textX + TextWidth(t.Text[:selStart])
		sw := TextWidth(t.Text[selStart:selEnd])
		FillRect(screen, sx, t.Y+2, sw, t.H-4, p.SelectionBG)
	}

	if t.Text == "" && t.Placeholder != "" && !t.Focused {
		DrawText(screen, t.Placeholder, textX, textY, 1, p.TextMuted)
	} else {
		if hasSel {
			if selStart > 0 {
				DrawText(screen, t.Text[:selStart], textX, textY, 1, p.TextPrimary)
			}
			sx := textX + TextWidth(t.Text[:selStart])
			DrawText(screen, t.Text[selStart:selEnd], sx, textY, 1, p.SelectionText)
			if selEnd < len(t.Text) {
				sx := textX + TextWidth(t.Text[:selEnd])
				DrawText(screen, t.Text[selEnd:], sx, textY, 1, p.TextPrimary)
			}
		} else {
			DrawText(screen, t.Text, textX, textY, 1, p.TextPrimary)
		}

		if t.Focused {
			t.cursorTick++
			if (t.cursorTick/30)%2 == 0 {
				cx := textX + TextWidth(t.Text[:t.cursorPos])
				FillRect(screen, cx, t.Y+2, 1, t.H-4, p.TextPrimary)
			}
		}
	}
}

func FillBorder(screen *ebiten.Image, x, y, w, h int, clr color.Color) {
	FillRect(screen, x, y, w, 1, clr)
	FillRect(screen, x, y+h-1, w, 1, clr)
	FillRect(screen, x, y, 1, h, clr)
	FillRect(screen, x+w-1, y, 1, h, clr)
}
