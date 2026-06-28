package math

type Rect struct {
	X, Y, W, H float64
}

func NewRect(x, y, w, h float64) Rect {
	return Rect{X: x, Y: y, W: w, H: h}
}

func (r Rect) Overlaps(other Rect) bool {
	return r.X < other.X+other.W &&
		r.X+r.W > other.X &&
		r.Y < other.Y+other.H &&
		r.Y+r.H > other.Y
}

func (r Rect) Center() Vec2 {
	return Vec2{X: r.X + r.W/2, Y: r.Y + r.H/2}
}
