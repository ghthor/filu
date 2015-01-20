package coord

import (
	"errors"
	"fmt"

	"github.com/ghthor/engine/time"
)

type (
	Cell struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	PartialCell struct {
		Cell
		Percentage float64
	}
)

func (c Cell) Neighbor(d Direction) Cell {
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

func (c Cell) Add(x, y int) Cell {
	c.X += x
	c.Y += y
	return c
}

func (c Cell) DirectionTo(other Cell) Direction {
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

func (p PartialCell) String() string {
	return fmt.Sprintf("PL{%v %v}", p.Cell, p.Percentage)
}

type (
	MoveAction interface {
		Start() time.Time
		End() time.Time
		CanHappenAfter(action MoveAction) bool
	}

	TurnAction struct {
		From, To Direction
		Time     time.Time
	}

	PathAction struct {
		time.Span
		Orig, Dest Cell
	}

	PathActionJson struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
		Orig  Cell      `json:"orig"`
		Dest  Cell      `json:"dest"`
	}
)

// Currenting in Frames, optimized for 40fps
// TODO Conditionalize this with the fps
const TurnActionDelay = 10

func (a TurnAction) Start() time.Time { return a.Time }
func (a TurnAction) End() time.Time   { return a.Time }
func (a TurnAction) CanHappenAfter(anAction MoveAction) bool {
	if anAction == nil {
		return true
	}

	switch action := anAction.(type) {
	case TurnAction:
		if a.Time-action.Time > TurnActionDelay {
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
	return fmt.Sprintf("PA{s:%v d:%v e:%v f:%v t:%v}",
		pa.Span.Start,
		pa.Span.Duration,
		pa.Span.End,
		pa.Orig, pa.Dest)
}

func (pa PathAction) Json() PathActionJson {
	return PathActionJson{
		pa.Span.Start,
		pa.Span.End,
		pa.Orig,
		pa.Dest,
	}
}

func (pa *PathAction) Start() time.Time { return pa.Span.Start }
func (pa *PathAction) End() time.Time   { return pa.Span.End }

func (pa *PathAction) CanHappenAfter(anAction MoveAction) bool {
	if anAction == nil {
		return true
	}

	switch action := anAction.(type) {
	case TurnAction:
		if pa.Start()-action.End() > TurnActionDelay && action.To == pa.Direction() {
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

func (pa PathAction) OrigPartial(now time.Time) (pc PartialCell) {
	pc.Cell = pa.Orig
	if now <= pa.Span.Start {
		pc.Percentage = 1.0
	} else if now >= pa.Span.End {
		pc.Percentage = 0.0
	} else {
		pc.Percentage = float64(pa.Remaining(now)) / float64(pa.Span.Duration)
	}
	return
}

func (pa PathAction) DestPartial(now time.Time) (pc PartialCell) {
	pc.Cell = pa.Dest
	if now <= pa.Span.Start {
		pc.Percentage = 0.0
	} else if now >= pa.Span.End {
		pc.Percentage = 1.0
	} else {
		pc.Percentage = 1.0 - (float64(pa.Remaining(now)) / float64(pa.Span.Duration))
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

func (pa PathAction) Traverses(c Cell) bool {
	return pa.Orig == c || pa.Dest == c
}

func (pa PathAction) TraversesAt(c Cell, t time.Time) (pc PartialCell, err error) {
	if t < pa.Span.Start || t > pa.Span.End {
		return pc, errors.New("timeOutOfRange")
	}

	if c == pa.Orig {
		if t == pa.Span.End {
			return pc, errors.New("miss")
		}
		pc = pa.OrigPartial(t)

	} else if pa.Dest == c {
		if t == pa.Span.Start {
			return pc, errors.New("miss")
		}
		pc = pa.DestPartial(t)

	} else {
		return pc, errors.New("wcOutOfRange")
	}
	return
}

func (pa PathAction) Crosses(path PathAction) bool {
	return pa.Traverses(path.Orig) || pa.Traverses(path.Dest)
}

type Direction byte

const (
	N, North Direction = iota, iota
	E, East
	S, South
	W, West
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
