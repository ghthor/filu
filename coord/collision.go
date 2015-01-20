package coord

import (
	"math"

	"github.com/ghthor/engine/time"
)

type (
	CollisionType int

	Collision interface {
		Type() CollisionType
		Start() time.Time
		End() time.Time
		OverlapAt(time.Time) float64
	}

	PathCollision struct {
		CollisionType
		time.Span
		A, B PathAction
	}

	CellCollision struct {
		CollisionType
		time.Span
		Cell Cell
		Path PathAction
	}
)

// TODO Implement this as bitflags
const (
	CT_NONE CollisionType = iota
	CT_HEAD_TO_HEAD
	CT_FROM_SIDE
	CT_A_INTO_B
	CT_A_INTO_B_FROM_SIDE
	CT_SWAP
	CT_SAME_ORIG
	CT_SAME_ORIG_PERP
	CT_SAME_ORIG_DEST
	CT_CELL_DEST
	CT_CELL_ORIG
)

func (c CollisionType) String() string {
	switch c {
	case CT_NONE:
		return "none"
	case CT_HEAD_TO_HEAD:
		return "head to head"
	case CT_FROM_SIDE:
		return "from the side"
	case CT_A_INTO_B:
		return "A into B"
	case CT_A_INTO_B_FROM_SIDE:
		return "A into B from the side"
	case CT_SWAP:
		return "swap"
	case CT_SAME_ORIG:
		return "same origin inverse facing"
	case CT_SAME_ORIG_PERP:
		return "same origin perpendicular facing"
	case CT_SAME_ORIG_DEST:
		return "same origin and destination"
	case CT_CELL_DEST:
		return "cell destination"
	case CT_CELL_ORIG:
		return "cell origin"
	}
	return "unknown collision type"
}

func (A PathAction) CollidesWith(B interface{}) (c Collision) {
	switch b := B.(type) {
	case PathAction:
		return pathCollision(A, b)
	case Cell:
		return cellCollision(A, b)
	default:
	}
	panic("unknown collision attempt")
}

func pathCollision(a, b PathAction) (c PathCollision) {
	var start, end time.Time
	c.A, c.B = a, b

	switch {
	case a.Orig == b.Orig && a.Dest == b.Dest:
		// A & B are moving out of the same Cell in the same direction
		c.CollisionType = CT_SAME_ORIG_DEST
		goto CT_SAME_ORIG_DEST_TIMESPAN

	case a.Orig == b.Orig:
		// A & B are moving out of the same Cell in different directions
		if a.Direction() == b.Direction().Reverse() {
			c.CollisionType = CT_SAME_ORIG
			goto CT_SAME_ORIG_TIMESPAN
		}
		// A & B are moving perpendicular to each other
		c.CollisionType = CT_SAME_ORIG_PERP
		goto CT_SAME_ORIG_PERP_TIMESPAN

	case a.Dest == b.Dest:
		// A & B are moving into the same Cell
		if a.Direction() == b.Direction().Reverse() {
			// Head to Head
			c.CollisionType = CT_HEAD_TO_HEAD
			goto CT_HEAD_TO_HEAD_TIMESPAN
		}
		// From the Side
		c.CollisionType = CT_FROM_SIDE
		goto CT_FROM_SIDE_TIMESPAN

	case a.Dest == b.Orig && b.Dest == a.Orig:
		// A and B are swapping positions
		c.CollisionType = CT_SWAP
		goto CT_SWAP_TIMESPAN

	case b.Dest == a.Orig:
		// Need to flip A and B
		a, b = b, a
		c.A, c.B = a, b
		fallthrough

	case a.Dest == b.Orig:
		// A is moving into the Cell B is leaving
		if a.Direction() == b.Direction() {
			if a.Span.Start >= b.Span.Start && a.Span.End >= b.Span.End {
				goto EXIT
			}
			c.CollisionType = CT_A_INTO_B
			goto CT_A_INTO_B_TIMESPAN

		} else {
			c.CollisionType = CT_A_INTO_B_FROM_SIDE
			goto CT_A_INTO_B_FROM_SIDE_TIMESPAN
		}
	default:
		goto EXIT
	}

CT_SAME_ORIG_DEST_TIMESPAN:
	if a.Span.Start < b.Span.Start {
		start = a.Span.Start
	} else {
		start = b.Span.Start
	}

	if a.Span.End > b.Span.End {
		end = a.Span.End
	} else {
		end = b.Span.End
	}

	c.Span = time.NewSpan(start, end)
	goto EXIT

CT_SAME_ORIG_TIMESPAN:
	if a.Span.Start < b.Span.Start {
		start = a.Span.Start
	} else {
		start = b.Span.Start
	}

	if a.Span.End == b.Span.Start {
		end = a.Span.End
	} else if b.Span.End == a.Span.Start {
		end = b.Span.End
	} else {
		var at, as, bt, bs float64
		// Starts
		at, bt = float64(a.Span.Start), float64(b.Span.Start)
		// Speeds
		as, bs = float64(a.Span.End-a.Span.Start), float64(b.Span.End-b.Span.Start)

		end = time.Time(math.Ceil((at*bs + bt*as + as*bs) / (bs + as)))

		// TODO Check if this floating point work around hack can be avoided or done differently
		if c.OverlapAt(end-1) == 0.0 {
			end -= 1
		}
	}

	c.Span = time.NewSpan(start, end)
	goto EXIT

CT_SAME_ORIG_PERP_TIMESPAN:
	if a.Span.Start < b.Span.Start {
		start = a.Span.Start
	} else {
		start = b.Span.Start
	}

	if a.Span.End < b.Span.End {
		end = a.Span.End
	} else {
		end = b.Span.End
	}
	c.Span = time.NewSpan(start, end)
	goto EXIT

CT_HEAD_TO_HEAD_TIMESPAN:
	// Start of collision
	if a.Span.Start == b.Span.End {
		start = a.Span.Start
	} else if b.Span.Start == a.Span.End {
		start = b.Span.Start
	} else {
		var at, as, bt, bs float64
		// Starts
		at, bt = float64(a.Span.Start), float64(b.Span.Start)
		// Speeds
		as, bs = float64(a.Span.End-a.Span.Start), float64(b.Span.End-b.Span.Start)

		start = time.Time(math.Floor((at*bs + bt*as + as*bs) / (bs + as)))

		// TODO Check if this floating point work around hack can be avoided or done differently
		if c.OverlapAt(start+1) == 0.0 {
			start += 1
		}
	}

	// End of Collision
	if a.Span.End >= b.Span.End {
		end = a.Span.End
	} else {
		end = b.Span.End
	}

	c.Span = time.NewSpan(start, end)
	goto EXIT

CT_FROM_SIDE_TIMESPAN:
	if a.Span.Start > b.Span.Start {
		start = a.Span.Start
	} else {
		start = b.Span.Start
	}

	if a.Span.End > b.Span.End {
		end = a.Span.End
	} else {
		end = b.Span.End
	}

	c.Span = time.NewSpan(start, end)
	goto EXIT

CT_SWAP_TIMESPAN:
	// TODO this is a.TimeSpan.Add(b.TimeSpan)
	if a.Span.Start <= b.Span.Start {
		start = a.Span.Start
	} else {
		start = b.Span.Start
	}

	if a.Span.End >= b.Span.End {
		end = a.Span.End
	} else {
		end = b.Span.End
	}

	c.Span = time.NewSpan(start, end)
	goto EXIT

CT_A_INTO_B_TIMESPAN:
	if a.Span.Start <= b.Span.Start {
		start = a.Span.Start
	} else {
		var as, ae, bs, be float64
		as, ae = float64(a.Span.Start), float64(a.Span.End)
		bs, be = float64(b.Span.Start), float64(b.Span.End)

		start = time.Time(math.Floor(((as / (ae - as)) - (bs / (be - bs))) / ((1 / (ae - as)) - (1 / (be - bs)))))

		// TODO Check if this floating point work around hack can be avoided or done differently
		if c.OverlapAt(start+1) == 0.0 {
			start += 1
		}
	}
	c.Span = time.NewSpan(start, b.Span.End)
	goto EXIT

CT_A_INTO_B_FROM_SIDE_TIMESPAN:
	start = a.Span.Start
	end = b.Span.End
	c.Span = time.NewSpan(start, end)

EXIT:
	return
}

func cellCollision(p PathAction, c Cell) (cc CellCollision) {
	cc.Path, cc.Cell = p, c
	switch c {
	case p.Dest:
		cc.CollisionType = CT_CELL_DEST
		cc.Span = p.Span
	case p.Orig:
		cc.CollisionType = CT_CELL_ORIG
		cc.Span = p.Span
	}
	return
}

func (c PathCollision) Type() CollisionType { return c.CollisionType }
func (c PathCollision) Start() time.Time    { return c.Span.Start }
func (c PathCollision) End() time.Time      { return c.Span.End }
func (c PathCollision) OverlapAt(t time.Time) (overlap float64) {

	switch c.CollisionType {
	case CT_SAME_ORIG:
		p := [...]PartialCell{
			c.A.OrigPartial(t),
			c.B.OrigPartial(t),
		}

		overlap = p[0].Percentage + p[1].Percentage - 1.0
		if overlap < 0.0 {
			overlap = 0.0
		}

	case CT_SAME_ORIG_PERP:
		p := [...]PartialCell{
			c.A.OrigPartial(t),
			c.B.OrigPartial(t),
		}
		overlap = p[0].Percentage * p[1].Percentage

	case CT_HEAD_TO_HEAD:
		p := [...]PartialCell{
			c.A.DestPartial(t),
			c.B.DestPartial(t),
		}
		sum := p[0].Percentage + p[1].Percentage
		if sum > 1.0 {
			overlap = sum - 1.0
		}

	case CT_FROM_SIDE:
		p := [...]PartialCell{
			c.A.DestPartial(t),
			c.B.DestPartial(t),
		}
		overlap = p[0].Percentage * p[1].Percentage

	case CT_SWAP:
		p := [...]PartialCell{
			c.A.DestPartial(t),
			c.B.DestPartial(t),
		}
		overlap = p[0].Percentage + p[1].Percentage
		if overlap > 1.0 {
			p = [...]PartialCell{
				c.A.OrigPartial(t),
				c.B.OrigPartial(t),
			}
			overlap = p[0].Percentage + p[1].Percentage
		}

	case CT_A_INTO_B:
		p := [...]PartialCell{
			c.A.DestPartial(t),
			c.B.OrigPartial(t),
		}
		sum := p[0].Percentage + p[1].Percentage
		if sum > 1.0 {
			overlap = sum - 1.0
		}

	case CT_A_INTO_B_FROM_SIDE:
		p := [...]PartialCell{
			c.A.DestPartial(t),
			c.B.OrigPartial(t),
		}
		overlap = p[0].Percentage * p[1].Percentage
	}
	return
}

func (c CellCollision) Type() CollisionType { return c.CollisionType }
func (c CellCollision) Start() time.Time    { return c.Span.Start }
func (c CellCollision) End() time.Time      { return c.Span.End }
func (c CellCollision) OverlapAt(t time.Time) (overlap float64) {
	switch c.CollisionType {
	case CT_CELL_DEST:
		overlap = c.Path.DestPartial(t).Percentage
	case CT_CELL_ORIG:
		overlap = c.Path.OrigPartial(t).Percentage
	}
	return
}
