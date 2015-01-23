package coord

import "errors"

type Bounds struct {
	TopL Cell `json:"tl"`
	BotR Cell `json:"br"`
}

func (aabb Bounds) Contains(c Cell) bool {
	return (aabb.TopL.X <= c.X && aabb.BotR.X >= c.X &&
		aabb.TopL.Y >= c.Y && aabb.BotR.Y <= c.Y)
}

func (aabb Bounds) HasOnEdge(c Cell) (onEdge bool) {
	x, y := c.X, c.Y
	switch {
	case (x == aabb.TopL.X || x == aabb.BotR.X) && (y <= aabb.TopL.Y && y >= aabb.BotR.Y):
		fallthrough
	case (y == aabb.TopL.Y || y == aabb.BotR.Y) && (x >= aabb.TopL.X && x <= aabb.BotR.X):
		onEdge = true
	default:
	}
	return
}

func (aabb Bounds) Width() int {
	return aabb.BotR.X - aabb.TopL.X + 1
}

func (aabb Bounds) Height() int {
	return aabb.TopL.Y - aabb.BotR.Y + 1
}

func (aabb Bounds) TopR() Cell { return Cell{aabb.BotR.X, aabb.TopL.Y} }
func (aabb Bounds) BotL() Cell { return Cell{aabb.TopL.X, aabb.BotR.Y} }

func (aabb Bounds) Area() int {
	return (aabb.BotR.X - aabb.TopL.X + 1) * (aabb.TopL.Y - aabb.BotR.Y + 1)
}

func (aabb Bounds) Overlaps(other Bounds) bool {
	if aabb.Contains(other.TopL) || aabb.Contains(other.BotR) ||
		other.Contains(aabb.TopL) || other.Contains(aabb.BotR) {
		return true
	}

	return aabb.Contains(other.TopR()) || aabb.Contains(other.BotL()) ||
		other.Contains(aabb.TopR()) || other.Contains(aabb.BotL())
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

func (aabb Bounds) Intersection(other Bounds) (Bounds, error) {
	if !aabb.Overlaps(other) {
		return Bounds{}, errors.New("no overlap")
	}

	return Bounds{
		Cell{max(aabb.TopL.X, other.TopL.X), min(aabb.TopL.Y, other.TopL.Y)},
		Cell{min(aabb.BotR.X, other.BotR.X), max(aabb.BotR.Y, other.BotR.Y)},
	}, nil
}

func (aabb Bounds) Expand(mag int) Bounds {
	aabb.TopL = aabb.TopL.Add(-mag, mag)
	aabb.BotR = aabb.BotR.Add(mag, -mag)
	return aabb
}

// Is BotR actually TopL?
func (aabb Bounds) IsInverted() bool {
	return aabb.BotR.Y > aabb.TopL.Y && aabb.BotR.X < aabb.TopL.X
}

// Flip TopL and BotR
func (aabb Bounds) Invert() Bounds {
	return Bounds{
		aabb.BotR, aabb.TopL,
	}
}

func (aabb Bounds) Quads() ([4]Bounds, error) {
	return splitAABBToQuads(aabb)
}

var ErrBoundsAreInverted = errors.New("aabb is inverted")
var ErrBoundsAreTooSmall = errors.New("aabb is too small to split")

func splitAABBToQuads(aabb Bounds) ([4]Bounds, error) {
	var aabbs [4]Bounds

	if aabb.IsInverted() {
		return aabbs, ErrBoundsAreInverted
	}

	w, h := aabb.Width(), aabb.Height()

	if w < 2 || h < 2 {
		return aabbs, ErrBoundsAreTooSmall
	}

	// NorthWest
	nw := Bounds{
		aabb.TopL,
		Cell{aabb.TopL.X + (w/2 - 1), aabb.TopL.Y - (h/2 - 1)},
	}

	// NorthEast
	ne := Bounds{
		Cell{nw.BotR.X + 1, aabb.TopL.Y},
		Cell{aabb.BotR.X, nw.BotR.Y},
	}

	se := Bounds{
		Cell{ne.TopL.X, ne.BotR.Y - 1},
		aabb.BotR,
	}

	sw := Bounds{
		Cell{aabb.TopL.X, se.TopL.Y},
		Cell{nw.BotR.X, aabb.BotR.Y},
	}

	const (
		QUAD_NW = iota
		QUAD_NE
		QUAD_SE
		QUAD_SW
	)

	aabbs[QUAD_NW] = nw
	aabbs[QUAD_NE] = ne
	aabbs[QUAD_SE] = se
	aabbs[QUAD_SW] = sw

	return aabbs, nil
}
