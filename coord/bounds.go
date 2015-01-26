package coord

import "errors"

type Bounds struct {
	TopL Cell `json:"tl"`
	BotR Cell `json:"br"`
}

func (b Bounds) Contains(c Cell) bool {
	return (b.TopL.X <= c.X && b.BotR.X >= c.X &&
		b.TopL.Y >= c.Y && b.BotR.Y <= c.Y)
}

func (b Bounds) HasOnEdge(c Cell) (onEdge bool) {
	x, y := c.X, c.Y
	switch {
	case (x == b.TopL.X || x == b.BotR.X) && (y <= b.TopL.Y && y >= b.BotR.Y):
		fallthrough
	case (y == b.TopL.Y || y == b.BotR.Y) && (x >= b.TopL.X && x <= b.BotR.X):
		onEdge = true
	default:
	}
	return
}

func (b Bounds) Width() int {
	return b.BotR.X - b.TopL.X + 1
}

func (b Bounds) Height() int {
	return b.TopL.Y - b.BotR.Y + 1
}

func (b Bounds) TopR() Cell { return Cell{b.BotR.X, b.TopL.Y} }
func (b Bounds) BotL() Cell { return Cell{b.TopL.X, b.BotR.Y} }

func (b Bounds) Area() int {
	return (b.BotR.X - b.TopL.X + 1) * (b.TopL.Y - b.BotR.Y + 1)
}

func (b Bounds) Overlaps(other Bounds) bool {
	if b.Contains(other.TopL) || b.Contains(other.BotR) ||
		other.Contains(b.TopL) || other.Contains(b.BotR) {
		return true
	}

	return b.Contains(other.TopR()) || b.Contains(other.BotL()) ||
		other.Contains(b.TopR()) || other.Contains(b.BotL())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (b Bounds) Intersection(other Bounds) (Bounds, error) {
	if !b.Overlaps(other) {
		return Bounds{}, errors.New("no overlap")
	}

	return Bounds{
		Cell{max(b.TopL.X, other.TopL.X), min(b.TopL.Y, other.TopL.Y)},
		Cell{min(b.BotR.X, other.BotR.X), max(b.BotR.Y, other.BotR.Y)},
	}, nil
}

func (b Bounds) Expand(mag int) Bounds {
	b.TopL = b.TopL.Add(-mag, mag)
	b.BotR = b.BotR.Add(mag, -mag)
	return b
}

// Is BotR actually TopL?
func (b Bounds) IsInverted() bool {
	return b.BotR.Y > b.TopL.Y && b.BotR.X < b.TopL.X
}

// Flip TopL and BotR
func (b Bounds) Invert() Bounds {
	return Bounds{
		b.BotR, b.TopL,
	}
}

var ErrBoundsAreInverted = errors.New("bounds are inverted")
var ErrBoundsAreTooSmall = errors.New("bounds are too small to split")

func (b Bounds) Quads() ([4]Bounds, error) {
	var bounds [4]Bounds

	if b.IsInverted() {
		return bounds, ErrBoundsAreInverted
	}

	w, h := b.Width(), b.Height()

	if w < 2 || h < 2 {
		return bounds, ErrBoundsAreTooSmall
	}

	// NorthWest
	nw := Bounds{
		b.TopL,
		Cell{b.TopL.X + (w/2 - 1), b.TopL.Y - (h/2 - 1)},
	}

	// NorthEast
	ne := Bounds{
		Cell{nw.BotR.X + 1, b.TopL.Y},
		Cell{b.BotR.X, nw.BotR.Y},
	}

	se := Bounds{
		Cell{ne.TopL.X, ne.BotR.Y - 1},
		b.BotR,
	}

	sw := Bounds{
		Cell{b.TopL.X, se.TopL.Y},
		Cell{nw.BotR.X, b.BotR.Y},
	}

	bounds[QUAD_NW] = nw
	bounds[QUAD_NE] = ne
	bounds[QUAD_SE] = se
	bounds[QUAD_SW] = sw

	return bounds, nil
}
