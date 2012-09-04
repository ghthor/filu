package engine

import (
	"errors"
	"fmt"
)

type (
	WorldCoord struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	PartialWorldCoord struct {
		WorldCoord
		Percentage float64
	}
)

func (c WorldCoord) Neighbor(d Direction) WorldCoord {
	switch d {
	case North:
		c.Y++
	case South:
		c.Y--
	case East:
		c.X++
	case West:
		c.X--
	}
	return c
}

func (c WorldCoord) Add(x, y int) WorldCoord {
	c.X += x
	c.Y += y
	return c
}

func (c WorldCoord) DirectionTo(other WorldCoord) Direction {
	x := other.X - c.X
	y := other.Y - c.Y

	switch {
	case x == 0 && y < 0:
		return South

	case x == 0 && y > 0:
		return North

	case x < 0 && y == 0:
		return West

	case x > 0 && y == 0:
		return East

	default:
	}

	panic("unable to calculate Direction")
}

func (p PartialWorldCoord) String() string {
	return fmt.Sprintf("PL{%v %v}", p.WorldCoord, p.Percentage)
}

type (
	MoveAction interface {
		Start() WorldTime
		End() WorldTime
		CanHappenAfter(action MoveAction) bool
	}

	TurnAction struct {
		from, to Direction
		time     WorldTime
	}

	PathAction struct {
		TimeSpan
		Orig, Dest WorldCoord
	}

	PathActionJson struct {
		Start WorldTime  `json:"start"`
		End   WorldTime  `json:"end"`
		Orig  WorldCoord `json:"orig"`
		Dest  WorldCoord `json:"dest"`
	}
)

// Currenting in Frames, optimized for 40fps
// TODO Conditionalize this with the fps
const TurnActionDelay = 10

func (a TurnAction) Start() WorldTime { return a.time }
func (a TurnAction) End() WorldTime   { return a.time }
func (a TurnAction) CanHappenAfter(anAction MoveAction) bool {
	if anAction == nil {
		return true
	}

	switch action := anAction.(type) {
	case TurnAction:
		if a.time-action.time > TurnActionDelay {
			return true
		} else {
			return false
		}
	case *PathAction:
		return true
	default:
	}
	panic("unknown MoveAction type")
}

func (pa PathAction) String() string {
	return fmt.Sprintf("PA{s:%v d:%v e:%v f:%v t:%v}", pa.start, pa.duration, pa.end, pa.Orig, pa.Dest)
}

func (pa PathAction) Json() PathActionJson {
	return PathActionJson{
		pa.start,
		pa.end,
		pa.Orig,
		pa.Dest,
	}
}

func (pa *PathAction) Start() WorldTime { return pa.TimeSpan.start }
func (pa *PathAction) End() WorldTime   { return pa.TimeSpan.end }

func (pa *PathAction) CanHappenAfter(anAction MoveAction) bool {
	if anAction == nil {
		return true
	}

	switch action := anAction.(type) {
	case TurnAction:
		if pa.Start()-action.End() > TurnActionDelay && action.to == pa.Direction() {
			return true
		} else {
			return false
		}
	case *PathAction:
		if pa.Start() == action.End() {
			return true
		} else if pa.Direction() == action.Direction() {
			return true
		} else {
			return false
		}
	default:
	}
	panic("unknown MoveAction type")
}

func (pa PathAction) OrigPartial(now WorldTime) (pwc PartialWorldCoord) {
	pwc.WorldCoord = pa.Orig
	if now <= pa.start {
		pwc.Percentage = 1.0
	} else if now >= pa.end {
		pwc.Percentage = 0.0
	} else {
		pwc.Percentage = float64(pa.Remaining(now)) / float64(pa.duration)
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
		pwc.Percentage = 1.0 - (float64(pa.Remaining(now)) / float64(pa.duration))
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

	default:
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
	CT_SWAP               CollisionType = iota
	CT_SAME_ORIG          CollisionType = iota
	CT_SAME_ORIG_DEST     CollisionType = iota
)

func (A PathAction) Collides(B PathAction) (c Collision) {

	// Check if time's overlap
	if !A.Overlaps(B.TimeSpan) {
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

	case A.Dest == B.Orig && B.Dest == A.Orig:
		// A and B are swapping positions
		c.Type = CT_SWAP

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

	case CT_SWAP:
		if c.A.start <= c.B.start {
			c.T = c.A.start
		} else {
			c.T = c.B.start
		}

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

func (d Direction) String() string {
	switch d {
	case North:
		return "north"

	case East:
		return "east"

	case South:
		return "south"

	case West:
		return "west"
	}
	panic("never reached")
}
