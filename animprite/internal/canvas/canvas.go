package canvas

import (
	"image/color"
	"math"

	"animprite/internal/camera"
	"animprite/internal/theme"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type SpriteRender struct {
	Image    *ebiten.Image
	OffsetX  float64
	OffsetY  float64
	ScaleX   float64
	ScaleY   float64
	Rotation float64
	OriginX  float64
	OriginY  float64
}

type SelectionRect struct {
	X, Y, W, H float64
	Rotation   float64 // radians
	Visible    bool
}

type HurtboxRender struct {
	OffsetX     float64
	OffsetY     float64
	WorldWidth  float64
	WorldHeight float64
	Rotation    float64 // radians
	Selected    bool
}

const handleHitRadius = 8.0

type Canvas struct {
	X, Y, Width, Height int
	BoundWidth          float64
	BoundHeight         float64
	Cam                 *camera.Camera
	th                  *theme.Manager
	drawBuf             *ebiten.Image

	spriteRenders     []SpriteRender
	Selection         SelectionRect
	HighlightedHandle int
	hurtboxRenders    []HurtboxRender
	whitePixel        *ebiten.Image
	fillPixel         *ebiten.Image
}

func (c *Canvas) SetSpriteRenders(r []SpriteRender) {
	c.spriteRenders = r
}

func (c *Canvas) SetSelectionRect(r SelectionRect) {
	c.Selection = r
}

func (c *Canvas) SetHurtboxRenders(r []HurtboxRender) {
	c.hurtboxRenders = r
}

func New(x, y, w, h int, th *theme.Manager) *Canvas {
	wp := ebiten.NewImage(1, 1)
	wp.Set(0, 0, color.RGBA{255, 255, 255, 255})
	fp := ebiten.NewImage(1, 1)
	fp.Set(0, 0, color.RGBA{255, 255, 0, 64})
	return &Canvas{
		X: x, Y: y, Width: w, Height: h,
		BoundWidth:  1280,
		BoundHeight: 720,
		Cam:         camera.New(),
		th:          th,
		whitePixel:  wp,
		fillPixel:   fp,
	}
}

func (c *Canvas) Draw(screen *ebiten.Image) {
	p := c.th.Current
	ui.FillRect(screen, c.X, c.Y, c.Width, c.Height, p.CanvasBG)

	if c.drawBuf == nil || c.drawBuf.Bounds().Dx() != c.Width || c.drawBuf.Bounds().Dy() != c.Height {
		if c.drawBuf != nil {
			c.drawBuf.Deallocate()
		}
		c.drawBuf = ebiten.NewImage(c.Width, c.Height)
	}
	c.drawBuf.Clear()

	cw := float64(c.Width)
	ch := float64(c.Height)

	cam := c.Cam
	cam.SetViewport(cw, ch)

	c.drawGrid(c.drawBuf, cam, p)
	c.drawOriginCrosshair(c.drawBuf, cam, p)
	c.drawBoundary(c.drawBuf, cam, p)
	if len(c.spriteRenders) > 0 {
		c.drawSprites(c.drawBuf, cam)
	}
	if c.Selection.Visible {
		c.drawSelection(c.drawBuf, cam, p)
	}
	if len(c.hurtboxRenders) > 0 {
		c.drawHurtboxes(c.drawBuf, cam)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(c.X), float64(c.Y))
	screen.DrawImage(c.drawBuf, op)
}

func (c *Canvas) drawGrid(screen *ebiten.Image, cam *camera.Camera, p theme.Palette) {
	gridSize := 50.0 * cam.Zoom
	if gridSize < 4 {
		return
	}
	camX := cam.X
	camY := cam.Y
	cw := cam.Width
	ch := cam.Height

	totalOffsetX := camX + cw/2
	startX := math.Mod(totalOffsetX, gridSize)
	if startX > 0 {
		startX -= gridSize
	}
	totalOffsetY := camY + ch/2
	startY := math.Mod(totalOffsetY, gridSize)
	if startY > 0 {
		startY -= gridSize
	}

	for gx := startX; gx < cw; gx += gridSize {
		strokeLine(screen, float32(gx), 0, float32(gx), float32(ch), p.CanvasGridLine)
	}
	for gy := startY; gy < ch; gy += gridSize {
		strokeLine(screen, 0, float32(gy), float32(cw), float32(gy), p.CanvasGridLine)
	}
}

func (c *Canvas) drawOriginCrosshair(screen *ebiten.Image, cam *camera.Camera, p theme.Palette) {
	ox, oy := cam.WorldToCanvas(0, 0)
	s := 16.0 * cam.Zoom
	if s < 4 {
		s = 4
	}
	strokeLine(screen, float32(ox-s), float32(oy), float32(ox+s), float32(oy), p.CanvasGridAxis)
	strokeLine(screen, float32(ox), float32(oy-s), float32(ox), float32(oy+s), p.CanvasGridAxis)
}

func (c *Canvas) drawBoundary(screen *ebiten.Image, cam *camera.Camera, p theme.Palette) {
	x1, y1 := cam.WorldToCanvas(-c.BoundWidth/2, -c.BoundHeight/2)
	x2, y2 := cam.WorldToCanvas(c.BoundWidth/2, c.BoundHeight/2)
	drawDashedRect(screen, float32(x1), float32(y1), float32(x2-x1), float32(y2-y1), p.CanvasBoundaryStroke)
}

func (c *Canvas) drawSprites(screen *ebiten.Image, cam *camera.Camera) {
	for _, s := range c.spriteRenders {
		if s.Image == nil {
			continue
		}
		w := float64(s.Image.Bounds().Dx())
		h := float64(s.Image.Bounds().Dy())
		ox := w * s.OriginX
		oy := h * s.OriginY

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-ox, -oy)
		op.GeoM.Scale(s.ScaleX, s.ScaleY)
		op.GeoM.Rotate(s.Rotation)
		op.GeoM.Translate(s.OffsetX, s.OffsetY)
		op.GeoM.Scale(cam.Zoom, cam.Zoom)
		op.GeoM.Translate(cam.Width/2+cam.X, cam.Height/2+cam.Y)

		screen.DrawImage(s.Image, op)
	}
}

func (c *Canvas) drawSelection(screen *ebiten.Image, cam *camera.Camera, p theme.Palette) {
	sel := c.Selection
	cx := sel.X + sel.W/2
	cy := sel.Y + sel.H/2
	cos := math.Cos(sel.Rotation)
	sin := math.Sin(sel.Rotation)

	// 8 local-space points (relative to center, half-dimensions)
	hw := sel.W / 2
	hh := sel.H / 2
	local := [][2]float64{
		{-hw, -hh}, // top-left
		{hw, -hh},  // top-right
		{hw, hh},   // bottom-right
		{-hw, hh},  // bottom-left
		{0, -hh},   // top-center
		{0, hh},    // bottom-center
		{-hw, 0},   // mid-left
		{hw, 0},    // mid-right
	}

	// Transform to world space (rotate around center), then to canvas space
	var canvasPts [][2]float64
	for _, p := range local {
		rx := cx + p[0]*cos - p[1]*sin
		ry := cy + p[0]*sin + p[1]*cos
		xx, yy := cam.WorldToCanvas(rx, ry)
		canvasPts = append(canvasPts, [2]float64{xx, yy})
	}

	// Dashed rectangle edges (0-3 are corners)
	clr := p.CanvasBoundaryStroke
	for i := 0; i < 4; i++ {
		next := (i + 1) % 4
		drawDashedLine(screen, float32(canvasPts[i][0]), float32(canvasPts[i][1]),
			float32(canvasPts[next][0]), float32(canvasPts[next][1]), 6, 4, clr)
	}

	// 8 handle squares (6px canvas pixels)
	hSz := float32(6)
	hH := hSz / 2
	for i := 0; i < 8; i++ {
		vector.DrawFilledRect(screen, float32(canvasPts[i][0])-hH, float32(canvasPts[i][1])-hH, hSz, hSz, clr, false)
	}

	// Highlighted handle (hover)
	if c.HighlightedHandle >= 0 && c.HighlightedHandle < 8 {
		hl := c.HighlightedHandle
		hlSz := float32(9)
		hlH := hlSz / 2
		hlClr := color.RGBA{255, 200, 0, 255}
		vector.DrawFilledRect(screen, float32(canvasPts[hl][0])-hlH, float32(canvasPts[hl][1])-hlH, hlSz, hlSz, hlClr, false)
	}
}

func (c *Canvas) HandleHitTest(mx, my int) int {
	if !c.Selection.Visible {
		return -1
	}
	sel := c.Selection
	cx := sel.X + sel.W/2
	cy := sel.Y + sel.H/2
	hw := sel.W / 2
	hh := sel.H / 2
	cos := math.Cos(sel.Rotation)
	sin := math.Sin(sel.Rotation)
	local := [8][2]float64{
		{-hw, -hh},
		{hw, -hh},
		{hw, hh},
		{-hw, hh},
		{0, -hh},
		{0, hh},
		{-hw, 0},
		{hw, 0},
	}
	for i, p := range local {
		wx := cx + p[0]*cos - p[1]*sin
		wy := cy + p[0]*sin + p[1]*cos
		sx, sy := c.Cam.WorldToCanvas(wx, wy)
		dx := float64(mx-c.X) - sx
		dy := float64(my-c.Y) - sy
		if dx*dx+dy*dy <= handleHitRadius*handleHitRadius {
			return i
		}
	}
	return -1
}

func strokeLine(screen *ebiten.Image, x1, y1, x2, y2 float32, clr color.Color) {
	if c, ok := clr.(color.RGBA); ok && c.A < 255 {
		a := float64(c.A) / 255
		c.R = uint8(float64(c.R) * a)
		c.G = uint8(float64(c.G) * a)
		c.B = uint8(float64(c.B) * a)
		c.A = 255
		clr = c
	}
	vector.StrokeLine(screen, x1, y1, x2, y2, 1, clr, false)
}

func drawDashedRect(screen *ebiten.Image, x, y, w, h float32, clr color.Color) {
	dash := float32(6)
	gap := float32(4)
	drawDashedLine(screen, x, y, x+w, y, dash, gap, clr)
	drawDashedLine(screen, x+w, y, x+w, y+h, dash, gap, clr)
	drawDashedLine(screen, x+w, y+h, x, y+h, dash, gap, clr)
	drawDashedLine(screen, x, y+h, x, y, dash, gap, clr)
}

func drawDashedLine(screen *ebiten.Image, x1, y1, x2, y2, dash, gap float32, clr color.Color) {
	if c, ok := clr.(color.RGBA); ok && c.A < 255 {
		a := float64(c.A) / 255
		c.R = uint8(float64(c.R) * a)
		c.G = uint8(float64(c.G) * a)
		c.B = uint8(float64(c.B) * a)
		c.A = 255
		clr = c
	}
	dx := x2 - x1
	dy := y2 - y1
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if dist == 0 {
		return
	}
	ux := dx / dist
	uy := dy / dist
	drawing := true
	cx, cy := x1, y1
	for i := float32(0); i < dist; {
		seg := dash
		if !drawing {
			seg = gap
		}
		nx := cx + ux*seg
		ny := cy + uy*seg
		if i+seg > dist {
			nx = x2
			ny = y2
		}
		if drawing {
			vector.StrokeLine(screen, cx, cy, nx, ny, 1, clr, false)
		}
		cx, cy = nx, ny
		i += seg
		drawing = !drawing
	}
}

func (c *Canvas) drawHurtboxes(screen *ebiten.Image, cam *camera.Camera) {
	for _, h := range c.hurtboxRenders {
		if h.WorldWidth <= 0 || h.WorldHeight <= 0 {
			continue
		}
		if h.Selected {
			hwFull := h.WorldWidth / 2
			hhFull := h.WorldHeight / 2
			cos := math.Cos(h.Rotation)
			sin := math.Sin(h.Rotation)
			local := [4][2]float64{
				{-hwFull, -hhFull},
				{hwFull, -hhFull},
				{-hwFull, hhFull},
				{hwFull, hhFull},
			}
			fillClr := color.RGBA{255, 255, 0, 207}
			fillA := float32(fillClr.A) / 255
			verts := [4]ebiten.Vertex{}
			for i, l := range local {
				wx := h.OffsetX + l[0]*cos - l[1]*sin
				wy := h.OffsetY + l[0]*sin + l[1]*cos
				cx, cy := cam.WorldToCanvas(wx, wy)
				verts[i] = ebiten.Vertex{
					DstX: float32(cx), DstY: float32(cy),
					SrcX: 0, SrcY: 0,
					ColorR: float32(fillClr.R) / 255 * fillA,
					ColorG: float32(fillClr.G) / 255 * fillA,
					ColorB: float32(fillClr.B) / 255 * fillA,
					ColorA: fillA,
				}
			}
			indices := []uint16{0, 1, 2, 1, 3, 2}
			screen.DrawTriangles(verts[:], indices, c.whitePixel, nil)
		}

		// Inward border: 4 rotated corners
		borderW := 4.0
		margin := math.Min(borderW/2/cam.Zoom, math.Min(h.WorldWidth, h.WorldHeight)/2)
		hw := h.WorldWidth/2 - margin
		hh := h.WorldHeight/2 - margin
		cos := math.Cos(h.Rotation)
		sin := math.Sin(h.Rotation)
		clr := color.RGBA{250, 232, 111, 255}
		if h.Selected {
			clr = color.RGBA{255, 241, 113, 255}
		}
		corners := [4][2]float64{
			{-hw, -hh},
			{hw, -hh},
			{hw, hh},
			{-hw, hh},
		}
		for i := 0; i < 4; i++ {
			n := (i + 1) % 4
			wx1 := h.OffsetX + corners[i][0]*cos - corners[i][1]*sin
			wy1 := h.OffsetY + corners[i][0]*sin + corners[i][1]*cos
			wx2 := h.OffsetX + corners[n][0]*cos - corners[n][1]*sin
			wy2 := h.OffsetY + corners[n][0]*sin + corners[n][1]*cos
			cx1, cy1 := cam.WorldToCanvas(wx1, wy1)
			cx2, cy2 := cam.WorldToCanvas(wx2, wy2)
			vector.StrokeLine(screen, float32(cx1), float32(cy1), float32(cx2), float32(cy2), float32(borderW), clr, false)
		}
	}
}
