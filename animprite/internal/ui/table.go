package ui

import (
	"animprite/internal/theme"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	tableTitleH  = 22
	tableHeaderH = 16
	tablePad     = 6
	tableBtnGap  = 8
)

type TableColumn struct {
	Label string
	Width int // 0 = flexible, >0 = fixed px
}

type Table struct {
	X, Y, W int
	Visible bool

	Title      string
	Columns    []TableColumn
	RowCount   int
	MaxVisible int
	RowH       int
	ColPad     int

	AddBtn    *Button
	RemoveBtn *Button
	ExtraBtns []*Button

	Scroll      int
	SelectedIdx int

	// Data bounds (set during Draw, used by HitRow/HitTest)
	DataX, DataY, DataW int

	// Per-row callback: screen, palette, table-left-x, row-top-y, content-width, column-xs[], row-index
	DrawRow func(screen *ebiten.Image, p theme.Palette, tx, rowY, cw int, xs []int, idx int)

	th *theme.Manager
}

func NewTable(title string, columns []TableColumn, maxVisible int, th *theme.Manager) *Table {
	return &Table{
		Title:       title,
		Columns:     columns,
		MaxVisible:  maxVisible,
		RowH:        22,
		ColPad:      4,
		Visible:     true,
		SelectedIdx: -1,
		th:          th,
	}
}

func (t *Table) Height() int {
	if !t.Visible {
		return 0
	}
	rows := t.RowCount
	if rows > t.MaxVisible {
		rows = t.MaxVisible
	}
	if rows == 0 {
		return tableTitleH + 2 + t.RowH + tablePad
	}
	return tableTitleH + 2 + tableHeaderH + rows*t.RowH + tablePad
}

func (t *Table) calcColumnPositions(cw int) []int {
	n := len(t.Columns)
	xs := make([]int, n)

	fixedTotal := 0
	flexCount := 0
	for _, col := range t.Columns {
		if col.Width > 0 {
			fixedTotal += col.Width
		} else {
			flexCount++
		}
	}

	gap := 2
	avail := cw - t.ColPad*2 - fixedTotal - gap*(n-1)
	flexW := 0
	if flexCount > 0 && avail > 0 {
		flexW = avail / flexCount
	}

	cx := t.ColPad
	for i, col := range t.Columns {
		xs[i] = cx
		if col.Width > 0 {
			cx += col.Width + gap
		} else {
			cx += flexW + gap
		}
	}
	return xs
}

func (t *Table) VisibleRange() (start, end int) {
	total := t.RowCount
	start = t.Scroll
	if start > total-t.MaxVisible {
		start = total - t.MaxVisible
	}
	if start < 0 {
		start = 0
		t.Scroll = 0
	}
	end = start + t.MaxVisible
	if end > total {
		end = total
	}
	return
}

func (t *Table) VisibleCount() int {
	start, end := t.VisibleRange()
	return end - start
}

func (t *Table) HitRow(mx, my int) int {
	if !t.Visible || t.DataW <= 0 {
		return -1
	}
	if mx >= t.DataX && mx <= t.DataX+t.DataW {
		if relY := my - t.DataY; relY >= 0 {
			if idx := t.Scroll + (relY / t.RowH); idx < t.RowCount {
				return idx
			}
		}
	}
	return -1
}

func (t *Table) HitTest(mx, my int) bool {
	if !t.Visible || t.DataW <= 0 {
		return false
	}
	start, end := t.VisibleRange()
	count := end - start
	return mx >= t.DataX && mx <= t.DataX+t.DataW &&
		my >= t.DataY && my < t.DataY+count*t.RowH
}

func (t *Table) ScrollBy(amount int) {
	if !t.Visible || t.RowCount == 0 {
		return
	}
	t.Scroll -= amount
	maxScroll := t.RowCount - t.MaxVisible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if t.Scroll < 0 {
		t.Scroll = 0
	}
	if t.Scroll > maxScroll {
		t.Scroll = maxScroll
	}
}

func (t *Table) DrawTitleRow(screen *ebiten.Image, p theme.Palette, tx, iy, cw int) int {
	btnY := iy + (tableTitleH-18)/2

	DrawText(screen, t.Title, tx+4, iy+4, 1.08, p.LabelColor)

	if t.RemoveBtn != nil {
		t.RemoveBtn.X = tx + cw - 2 - t.RemoveBtn.W
		t.RemoveBtn.Y = btnY
		t.RemoveBtn.Draw(screen)
	}

	nextX := tx + cw - 2
	if t.RemoveBtn != nil {
		nextX = t.RemoveBtn.X - tableBtnGap
	}

	for i := len(t.ExtraBtns) - 1; i >= 0; i-- {
		btn := t.ExtraBtns[i]
		btn.X = nextX - btn.W
		btn.Y = btnY
		btn.Draw(screen)
		nextX = btn.X - tableBtnGap
	}

	if t.AddBtn != nil {
		t.AddBtn.X = nextX - tableBtnGap - t.AddBtn.W
		t.AddBtn.Y = btnY
		t.AddBtn.Draw(screen)
	}

	iy += tableTitleH
	FillRect(screen, tx, iy, cw-2, 1, p.Border)
	iy += 2
	return iy
}

func (t *Table) DrawHeaderRow(screen *ebiten.Image, p theme.Palette, tx, iy int, xs []int) int {
	for i, col := range t.Columns {
		DrawText(screen, col.Label, tx+xs[i], iy+2, 1.08, p.LabelColor)
	}
	return iy + tableHeaderH
}

func (t *Table) drawScrollbar(screen *ebiten.Image, p theme.Palette, tx, iy, cw int) {
	if t.RowCount <= t.MaxVisible {
		return
	}
	start, end := t.VisibleRange()
	visible := end - start
	sx := tx + cw - 5
	sw := 3
	dataH := visible * t.RowH
	thumbH := dataH * t.MaxVisible / t.RowCount
	if thumbH < 6 {
		thumbH = 6
	}
	if thumbH > dataH {
		thumbH = dataH
	}
	maxScroll := t.RowCount - t.MaxVisible
	if maxScroll > 0 {
		pos := float64(t.Scroll) / float64(maxScroll)
		thumbY := iy + int(pos*float64(dataH-thumbH))
		FillRect(screen, sx, iy, sw, dataH, p.ScrollbarTrack)
		FillRect(screen, sx, thumbY, sw, thumbH, p.ScrollbarThumb)
	}
}

// HandleTitleRowMouse dispatches mouse events to the title row buttons.
func (t *Table) HandleTitleRowMouse(mx, my int, justL bool) {
	if t.AddBtn != nil {
		t.AddBtn.HandleMouse(mx, my, justL)
	}
	if t.RemoveBtn != nil {
		t.RemoveBtn.HandleMouse(mx, my, justL)
	}
	for _, btn := range t.ExtraBtns {
		if btn != nil {
			btn.HandleMouse(mx, my, justL)
		}
	}
}

// Draw renders the full table and returns the Y position after the table.
func (t *Table) Draw(screen *ebiten.Image, p theme.Palette, x, iy, w int) int {
	if !t.Visible {
		return iy
	}

	tx := x + tablePad
	cw := w - tablePad*2

	// Background box
	th := t.Height()
	FillRect(screen, tx, iy, cw, th, p.TableBG)
	FillBorder(screen, tx, iy, cw, th, p.Border)

	// Title row with add/remove buttons
	iy = t.DrawTitleRow(screen, p, tx, iy, cw)

	if t.RowCount == 0 {
		DrawText(screen, "no items", tx+4, iy+4, 1, p.TextMuted)
		return iy + t.RowH + tablePad
	}

	// Column headers
	xs := t.calcColumnPositions(cw)
	iy = t.DrawHeaderRow(screen, p, tx, iy, xs)

	// Store data bounds for hit-testing
	t.DataX = tx
	t.DataY = iy
	t.DataW = cw

	// Draw visible rows
	start, end := t.VisibleRange()
	visible := end - start

	for i := 0; i < visible; i++ {
		idx := start + i
		rowY := iy + i*t.RowH

		if t.SelectedIdx == idx {
			FillRect(screen, tx+1, rowY, cw-2, t.RowH, p.SelectedRow)
		}

		if t.DrawRow != nil {
			t.DrawRow(screen, p, tx, rowY, cw, xs, idx)
		}
	}

	// Scrollbar
	t.drawScrollbar(screen, p, tx, iy, cw)

	return iy + visible*t.RowH + tablePad
}
