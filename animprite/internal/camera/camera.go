package camera

import "math"

type Camera struct {
	X, Y   float64
	Zoom   float64
	Width  float64
	Height float64
}

func New() *Camera {
	return &Camera{
		Zoom: 1.0,
	}
}

func (c *Camera) WorldToCanvas(wx, wy float64) (float64, float64) {
	return wx*c.Zoom + c.Width/2 + c.X, wy*c.Zoom + c.Height/2 + c.Y
}

func (c *Camera) CanvasToWorld(cx, cy float64) (float64, float64) {
	return (cx - c.Width/2 - c.X) / c.Zoom, (cy - c.Height/2 - c.Y) / c.Zoom
}

func (c *Camera) Pan(dx, dy float64) {
	c.X += dx
	c.Y += dy
}

func (c *Camera) ZoomAt(amount float64, cx, cy float64) {
	beforeX, beforeY := c.CanvasToWorld(cx, cy)
	c.Zoom = clamp(c.Zoom*amount, 0.05, 20)
	afterX, afterY := c.CanvasToWorld(cx, cy)
	c.X += (afterX - beforeX) * c.Zoom
	c.Y += (afterY - beforeY) * c.Zoom
}

func (c *Camera) Reset() {
	c.X, c.Y = 0, 0
	c.Zoom = 1.0
}

func (c *Camera) SetViewport(w, h float64) {
	c.Width = w
	c.Height = h
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}
