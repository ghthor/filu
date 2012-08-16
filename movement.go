package engine

import (
	"errors"
	"fmt"
)

type Action struct {
	start, end WorldTime
	duration   int64
}

func NewAction(start, end WorldTime) Action {
	return Action{
		start,
		end,
		int64(end) - int64(start),
	}
}

func (a Action) TimeLeft(from WorldTime) int64 {
	return int64(a.end) - int64(from)
}

func (a Action) ExistsAt(t WorldTime) bool {
	return a.start <= t && t <= a.end
}

func (a Action) HappensDuring(other Action) bool {
	return a.ExistsAt(other.start) ||
		a.ExistsAt(other.end) ||
		other.ExistsAt(a.start) ||
		other.ExistsAt(a.end)
}

type (
	WorldCoord struct {
		X, Y int
	}

	PartialWorldCoord struct {
		WorldCoord
		Percentage float64
	}
)

type (
	StandAction struct {
		WorldCoord
		Facing Direction
	}

	PathAction struct {
		Action
		Orig, Dest WorldCoord
	}
)

func (p PartialWorldCoord) String() string {
	return fmt.Sprintf("PL{%v %v}", p.WorldCoord, p.Percentage)
}

func (pa PathAction) String() string {
	return fmt.Sprintf("PA{s:%v d:%v e:%v f:%v t:%v}", pa.start, pa.duration, pa.end, pa.Orig, pa.Dest)
}

func (pa PathAction) OrigPartial(now WorldTime) (pwc PartialWorldCoord) {
	pwc.WorldCoord = pa.Orig
	if now <= pa.start {
		pwc.Percentage = 1.0
	} else if now >= pa.end {
		pwc.Percentage = 0.0
	} else {
		pwc.Percentage = float64(pa.TimeLeft(now)) / float64(pa.duration)
	}
	return
}

func (pa PathAction) DestPartial(now WorldTime) (pwc PartialWorldCoord) {
	pwc.WorldCoord = pa.Dest
	if now <= pa.start {
		pwc.Percentage = 0.0
	} else if now >= pa.end {
		pwc.Percentage = 1.0
	} else {
		pwc.Percentage = 1.0 - (float64(pa.TimeLeft(now)) / float64(pa.duration))
	}
	return
}

func (pa PathAction) Direction() Direction {
	x := pa.Dest.X - pa.Orig.X
	y := pa.Dest.Y - pa.Orig.Y

	switch {
	case x == 0 && y < 0:
		return South

	case x == 0 && y > 0:
		return North

	case x < 0 && y == 0:
		return West

	case x > 0 && y == 0:
		return East

	case x == 0 && y == 0:
	}

	panic("invalid PathAction")
}

func (pa PathAction) IsParallelTo(pa2 PathAction) bool {
	return pa.Direction().IsParallelTo(pa2.Direction())
}

func (pa PathAction) Traverses(wc WorldCoord) bool {
	return pa.Orig == wc || pa.Dest == wc
}

func (pa PathAction) TraversesAt(wc WorldCoord, t WorldTime) (pwc PartialWorldCoord, err error) {
	if t < pa.start || t > pa.end {
		return pwc, errors.New("timeOutOfRange")
	}

	if wc == pa.Orig {
		if t == pa.end {
			return pwc, errors.New("miss")
		}
		pwc = pa.OrigPartial(t)

	} else if pa.Dest == wc {
		if t == pa.start {
			return pwc, errors.New("miss")
		}
		pwc = pa.DestPartial(t)

	} else {
		return pwc, errors.New("wcOutOfRange")
	}
	return
}

func (pa PathAction) Crosses(path PathAction) bool {
	return pa.Traverses(path.Orig) || pa.Traverses(path.Dest)
}

type (
	CollisionType int

	Collision struct {
		Type CollisionType
		T    WorldTime
		A, B PathAction
	}
)

func (c Collision) Overlap() (overlap float64) {

	switch c.Type {
	case CT_HEAD_TO_HEAD:
		p := [...]PartialWorldCoord{
			c.A.DestPartial(c.T),
			c.B.DestPartial(c.T),
		}

		sum := p[0].Percentage + p[1].Percentage
		if sum > 1.0 {
			overlap = sum - 1.0
		}
	case CT_FROM_SIDE:
		p := [...]PartialWorldCoord{
			c.A.DestPartial(c.T),
			c.B.DestPartial(c.T),
		}

		overlap = p[0].Percentage * p[1].Percentage
	case CT_A_INTO_B:
		p := [...]PartialWorldCoord{
			c.A.DestPartial(c.T),
			c.B.OrigPartial(c.T),
		}

		sum := p[0].Percentage + p[1].Percentage
		if sum > 1.0 {
			overlap = sum - 1.0
		}

	case CT_A_INTO_B_FROM_SIDE:
		p := [...]PartialWorldCoord{
			c.A.DestPartial(c.T),
			c.B.OrigPartial(c.T),
		}

		overlap = p[0].Percentage * p[1].Percentage

	default:
	}
	return
}

// TODO Implement this as bitflags
const (
	CT_NONE               CollisionType = iota
	CT_HEAD_TO_HEAD       CollisionType = iota
	CT_FROM_SIDE          CollisionType = iota
	CT_A_INTO_B           CollisionType = iota
	CT_A_INTO_B_FROM_SIDE CollisionType = iota
	CT_SAME_ORIG          CollisionType = iota
	CT_SAME_ORIG_DEST     CollisionType = iota
)

func (A PathAction) Collides(B PathAction) (c Collision) {

	// Check if time's overlap
	if !A.HappensDuring(B.Action) {
		return
	}

	c.A, c.B = A, B

	switch {
	case A.Orig == B.Orig && A.Dest == B.Dest:
		// A & B are moving out of the same WorldCoord in the same direction
		c.Type = CT_SAME_ORIG_DEST

	case A.Orig == B.Orig:
		// A & B are moving out of the same WorldCoord in different directions
		c.Type = CT_SAME_ORIG

	case A.Dest == B.Dest:
		// A & B are moving into the same WorldCoord
		if A.Direction() == B.Direction().Reverse() {
			// Head to Head
			c.Type = CT_HEAD_TO_HEAD
		} else {
			// From the Side
			c.Type = CT_FROM_SIDE
		}

	case B.Dest == A.Orig:
		// Need to flip A and B
		c.A, c.B = B, A
		fallthrough

	case A.Dest == B.Orig:
		// A is moving into the WorldCoord B is leaving
		if c.A.Direction() == c.B.Direction() {
			if c.A.start >= c.B.start && c.A.end >= c.B.end {
				return
			}
			c.Type = CT_A_INTO_B
		} else {
			c.Type = CT_A_INTO_B_FROM_SIDE
		}

	default:
		// No Collision
		return
	}

	c.findStart()

	return
}

func (c *Collision) findStart() {
	switch c.Type {
	case CT_HEAD_TO_HEAD, CT_FROM_SIDE:
		switch {
		case c.A.start <= c.B.start:
			c.T = c.B.start

		case c.A.start > c.B.start:
			c.T = c.A.start
		}

	case CT_A_INTO_B, CT_A_INTO_B_FROM_SIDE:
		c.T = c.A.start

	default:
	}
}

type Direction byte

const (
	N, North Direction = iota, iota
	E, East  Direction = iota, iota
	S, South Direction = iota, iota
	W, West  Direction = iota, iota
)

func (d Direction) IsParallelTo(p Direction) bool {
	switch {
	case d == North || d == South:
		return p == North || p == South

	case d == East || d == West:
		return p == East || p == West
	}
	panic("never reached")
}

func (d Direction) Reverse() Direction {
	switch d {
	case North:
		return South

	case East:
		return West

	case South:
		return North

	case West:
		return East
	}
	panic("never reached")
}
