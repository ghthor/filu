package coord

import "errors"

type Bounds struct {
	TopL Cell `json:"tl"`
	BotR Cell `json:"br"`
}

func (b Bounds) Contains(c Cell) bool {
	// TODO there might be some optimizations laying right here
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
	return abs(b.BotR.X-b.TopL.X) + 1
}

func (b Bounds) Height() int {
	return abs(b.TopL.Y-b.BotR.Y) + 1
}

func (b Bounds) TopR() Cell { return Cell{b.BotR.X, b.TopL.Y} }
func (b Bounds) BotL() Cell { return Cell{b.TopL.X, b.BotR.Y} }

func (b Bounds) Area() int {
	return (abs(b.BotR.X-b.TopL.X) + 1) * (abs(b.TopL.Y-b.BotR.Y) + 1)
}

func (b Bounds) Overlaps(other Bounds) bool {
	if other.TopL == other.BotR {
		return b.Contains(other.TopL)
	}

	if b.TopL == b.BotR {
		return other.Contains(b.TopL)
	}

	if b.Contains(other.TopL) || b.Contains(other.BotR) ||
		other.Contains(b.TopL) || other.Contains(b.BotR) {
		return true
	}

	return b.Contains(other.TopR()) || b.Contains(other.BotL()) ||
		other.Contains(b.TopR()) || other.Contains(b.BotL())
}

func abs(a int) int {
	if a < 0 {
		a = -a
	}
	return a
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

var ErrNoOverlap = errors.New("no overlap")

func (b Bounds) Intersection(other Bounds) (Bounds, error) {
	if !b.Overlaps(other) {
		return Bounds{}, ErrNoOverlap
	}

	return Bounds{
		Cell{max(b.TopL.X, other.TopL.X), min(b.TopL.Y, other.TopL.Y)},
		Cell{min(b.BotR.X, other.BotR.X), max(b.BotR.Y, other.BotR.Y)},
	}, nil
}

func (b Bounds) Join(with Bounds) Bounds {
	if b.IsInverted() {
		b = b.Invert()
	}

	if with.IsInverted() {
		with = with.Invert()
	}

	return Bounds{
		TopL: Cell{
			X: min(b.TopL.X, with.TopL.X),
			Y: max(b.TopL.Y, with.TopL.Y),
		},
		BotR: Cell{
			X: max(b.BotR.X, with.BotR.X),
			Y: min(b.BotR.Y, with.BotR.Y),
		}}
}

func (b Bounds) JoinAll(bounds ...Bounds) Bounds {
	for i := 0; i < len(bounds); i++ {
		b = b.Join(bounds[i])
	}

	return b
}

func JoinBounds(bounds ...Bounds) Bounds {
	if len(bounds) == 1 {
		return bounds[0]
	}

	return bounds[0].JoinAll(bounds[1:]...)
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

	bounds[NW] = nw
	bounds[NE] = ne
	bounds[SE] = se
	bounds[SW] = sw

	return bounds, nil
}

// Diffs 2 overlapping bounds and returns a slice
// of rectangles that are contained within b and
// not contained within other. The slice of rectangles
// will either have 1,2,3,5,8 rectangles.
//
//    1 rect - Iff w && h are the same and it shares 2 parallel edges
//             aka, its been translated in 1 of the 4 directions
//    2 rect - TODO Iff w || h are different but it shares 2 perpendicular edges
//    3 rect - TODO If w && h are different and it shares 2 perpendicular edges
//    3 rect - *TODO If w && h are the same, but it's been translated by 2 directions
//    5 rect - TODO Iff w && h are different and other shares 1 edge with b
//    8 rect - TODO Iff b contains other and shares no edges
func (a Bounds) DiffFrom(b Bounds) []Bounds {
	switch {
	case !a.Overlaps(b):
		// No Overlap
	case a.TopL == b.TopL && a.BotR == b.BotR:
		// Same Bounds

	case a.Width() == b.Width() && a.Height() == b.Height():
		switch {
		case a.TopL.X == b.TopL.X && a.BotR.X == b.BotR.X:
			// b Translated North/South
			return []Bounds{a.vdiff(b)}
		case a.TopL.Y == b.TopL.Y && a.BotR.Y == b.BotR.Y:
			// b Translated East/West
			return []Bounds{a.hdiff(b)}
		default:
			rects := a.diff3(b)
			return rects[0:]
		}

	default:
	}
	return nil
}

func (a Bounds) vdiff(b Bounds) Bounds {
	switch max(a.TopL.Y, b.TopL.Y) {
	case a.TopL.Y:
		// b Translated South
		return Bounds{
			Cell{
				a.TopL.X,
				a.BotR.Y - 1,
			},
			b.BotR,
		}

	case b.TopL.Y:
		// b Translated North
		return Bounds{
			b.TopL,
			Cell{
				b.BotR.X,
				a.TopL.Y + 1,
			},
		}
	}

	return Bounds{}
}

func (a Bounds) hdiff(b Bounds) Bounds {
	switch max(a.BotR.X, b.BotR.X) {
	case b.BotR.X:
		// b Translated East
		return Bounds{
			Cell{
				a.BotR.X + 1,
				a.TopL.Y,
			},
			b.BotR,
		}

	case a.BotR.X:
		// b Translated West
		return Bounds{
			b.TopL,
			Cell{
				a.TopL.X - 1,
				b.BotR.Y,
			},
		}
	}

	return Bounds{}
}

func (a Bounds) diff3(b Bounds) (rects [3]Bounds) {
	switch {
	case a.Contains(b.BotL()):
		rects = a.diff3ne(b)

	case a.Contains(b.TopL):
		rects = a.diff3se(b)

	case a.Contains(b.TopR()):
		rects = a.diff3sw(b)

	case a.Contains(b.BotR):
		rects = a.diff3nw(b)
	}
	return
}

func (a Bounds) diff3ne(b Bounds) [3]Bounds {
	return [3]Bounds{{
		b.TopL,
		Cell{a.BotR.X, a.TopL.Y + 1},
	}, {
		Cell{a.BotR.X + 1, b.TopL.Y},
		Cell{b.BotR.X, a.TopL.Y + 1},
	}, {
		Cell{a.BotR.X + 1, a.TopL.Y},
		b.BotR,
	}}
}

func (a Bounds) diff3se(b Bounds) (rects [3]Bounds) {
	return [3]Bounds{{
		Cell{a.BotR.X + 1, b.TopL.Y},
		Cell{b.BotR.X, a.BotR.Y},
	}, {
		Cell{a.BotR.X + 1, a.BotR.Y - 1},
		b.BotR,
	}, {
		Cell{b.TopL.X, a.BotR.Y - 1},
		Cell{a.BotR.X, b.BotR.Y},
	}}
}

func (a Bounds) diff3sw(b Bounds) (rects [3]Bounds) {
	return [3]Bounds{{
		Cell{a.TopL.X, a.BotR.Y - 1},
		b.BotR,
	}, {
		Cell{b.TopL.X, a.BotR.Y - 1},
		Cell{a.TopL.X - 1, b.BotR.Y},
	}, {
		b.TopL,
		Cell{a.TopL.X - 1, a.BotR.Y},
	}}
}

func (a Bounds) diff3nw(b Bounds) (rects [3]Bounds) {
	return [3]Bounds{{
		Cell{b.TopL.X, a.TopL.Y},
		Cell{a.TopL.X - 1, b.BotR.Y},
	}, {
		b.TopL,
		Cell{a.TopL.X - 1, a.TopL.Y + 1},
	}, {
		Cell{a.TopL.X, b.TopL.Y},
		Cell{b.BotR.X, a.TopL.Y + 1},
	}}
}
