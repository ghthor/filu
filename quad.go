package engine

import (
	"errors"
)

type (
	AABB struct {
		TopL, BotR WorldCoord
	}
)

func (aabb AABB) Contains(c WorldCoord) bool {
	return (aabb.TopL.X <= c.X && aabb.BotR.X >= c.X &&
		aabb.TopL.Y >= c.Y && aabb.BotR.Y <= c.Y)
}

func (aabb AABB) Width() int {
	return aabb.BotR.X - aabb.TopL.X + 1
}

func (aabb AABB) Height() int {
	return aabb.TopL.Y - aabb.BotR.Y + 1
}

func (aabb AABB) TopR() WorldCoord { return WorldCoord{aabb.BotR.X, aabb.TopL.Y} }
func (aabb AABB) BotL() WorldCoord { return WorldCoord{aabb.TopL.X, aabb.BotR.Y} }

func (aabb AABB) Area() int {
	return (aabb.BotR.X - aabb.TopL.X + 1) * (aabb.TopL.Y - aabb.BotR.Y + 1)
}

func (aabb AABB) Overlaps(other AABB) bool {
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

func (aabb AABB) Intersection(other AABB) (AABB, error) {
	if !aabb.Overlaps(other) {
		return AABB{}, errors.New("no overlap")
	}

	return AABB{
		WorldCoord{max(aabb.TopL.X, other.TopL.X), min(aabb.TopL.Y, other.TopL.Y)},
		WorldCoord{min(aabb.BotR.X, other.BotR.X), max(aabb.BotR.Y, other.BotR.Y)},
	}, nil
}

// Is BotR actually TopL?
func (aabb AABB) IsInverted() bool {
	return aabb.BotR.Y > aabb.TopL.Y && aabb.BotR.X < aabb.TopL.X
}

// Flip TopL and BotR
func (aabb AABB) Invert() AABB {
	return AABB{
		aabb.BotR, aabb.TopL,
	}
}

const (
	QUAD_NW = iota
	QUAD_NE
	QUAD_SE
	QUAD_SW
)

func splitAABBToQuads(aabb AABB) ([4]AABB, error) {
	var aabbs [4]AABB

	if aabb.IsInverted() {
		return aabbs, errors.New("aabb is inverted")
	}

	w, h := aabb.Width(), aabb.Height()

	if w < 2 || h < 2 {
		return aabbs, errors.New("aabb is too small to split")
	}

	// NorthWest
	nw := AABB{
		aabb.TopL,
		WorldCoord{aabb.TopL.X + (w/2 - 1), aabb.TopL.Y - (h/2 - 1)},
	}

	// NorthEast
	ne := AABB{
		WorldCoord{nw.BotR.X + 1, aabb.TopL.Y},
		WorldCoord{aabb.BotR.X, nw.BotR.Y},
	}

	se := AABB{
		WorldCoord{ne.TopL.X, ne.BotR.Y - 1},
		aabb.BotR,
	}

	sw := AABB{
		WorldCoord{aabb.TopL.X, se.TopL.Y},
		WorldCoord{nw.BotR.X, aabb.BotR.Y},
	}

	aabbs[QUAD_NW] = nw
	aabbs[QUAD_NE] = ne
	aabbs[QUAD_SE] = se
	aabbs[QUAD_SW] = sw

	return aabbs, nil
}
