package engine

import (
	"math"
)

type (
	CollisionType int

	// This will be renamed to Collision
	Collision interface {
		Type() CollisionType
		Start() WorldTime
		End() WorldTime
		OverlapAt(WorldTime) float64
	}

	PathCollision struct {
		CollisionType
		TimeSpan
		A, B PathAction
	}

	CellCollision struct {
		CollisionType
		TimeSpan
		Cell Cell
		Path PathAction
	}
)

// TODO Implement this as bitflags
const (
	CT_NONE               CollisionType = iota
	CT_HEAD_TO_HEAD       CollisionType = iota
	CT_FROM_SIDE          CollisionType = iota
	CT_A_INTO_B           CollisionType = iota
	CT_A_INTO_B_FROM_SIDE CollisionType = iota
	CT_SWAP               CollisionType = iota
	CT_SAME_ORIG          CollisionType = iota
	CT_SAME_ORIG_PERP     CollisionType = iota
	CT_SAME_ORIG_DEST     CollisionType = iota
	CT_CELL_DEST          CollisionType = iota
	CT_CELL_ORIG          CollisionType = iota
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
		return "same origin"
	case CT_SAME_ORIG_DEST:
		return "same orign and destination"
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
	var start, end WorldTime
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
			goto EXIT
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
			if a.start >= b.start && a.end >= b.end {
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
	if a.start < b.start {
		start = a.start
	} else {
		start = b.start
	}

	if a.end > b.end {
		end = a.end
	} else {
		end = b.end
	}

	c.TimeSpan = NewTimeSpan(start, end)
	goto EXIT

CT_SAME_ORIG_PERP_TIMESPAN:
	if a.start < b.start {
		start = a.start
	} else {
		start = b.start
	}

	if a.end < b.end {
		end = a.end
	} else {
		end = b.end
	}
	c.TimeSpan = NewTimeSpan(start, end)
	goto EXIT

CT_HEAD_TO_HEAD_TIMESPAN:
	// Start of collision
	if a.start == b.end {
		start = a.start
	} else if b.start == a.end {
		start = b.start
	} else {
		var at, as, bt, bs float64
		// Starts
		at, bt = float64(a.start), float64(b.start)
		// Speeds
		as, bs = float64(a.end-a.start), float64(b.end-b.start)

		start = WorldTime(math.Floor((at*bs + bt*as + as*bs) / (bs + as)))

		// TODO Check if this floating point work around hack can be avoided or done differently
		if c.OverlapAt(start+1) == 0.0 {
			start += 1
		}
	}

	// End of Collision
	if a.end >= b.end {
		end = a.end
	} else {
		end = b.end
	}

	c.TimeSpan = NewTimeSpan(start, end)
	goto EXIT

CT_FROM_SIDE_TIMESPAN:
	if a.start > b.start {
		start = a.start
	} else {
		start = b.start
	}

	if a.end > b.end {
		end = a.end
	} else {
		end = b.end
	}

	c.TimeSpan = NewTimeSpan(start, end)
	goto EXIT

CT_SWAP_TIMESPAN:
	// TODO this is a.TimeSpan.Add(b.TimeSpan)
	if a.start <= b.start {
		start = a.start
	} else {
		start = b.start
	}

	if a.end >= b.end {
		end = a.end
	} else {
		end = b.end
	}

	c.TimeSpan = NewTimeSpan(start, end)
	goto EXIT

CT_A_INTO_B_TIMESPAN:
	if a.start <= b.start {
		start = a.start
	} else {
		var as, ae, bs, be float64
		as, ae = float64(a.start), float64(a.end)
		bs, be = float64(b.start), float64(b.end)

		start = WorldTime(math.Floor(((as / (ae - as)) - (bs / (be - bs))) / ((1 / (ae - as)) - (1 / (be - bs)))))

		// TODO Check if this floating point work around hack can be avoided or done differently
		if c.OverlapAt(start+1) == 0.0 {
			start += 1
		}
	}
	c.TimeSpan = NewTimeSpan(start, b.end)
	goto EXIT

CT_A_INTO_B_FROM_SIDE_TIMESPAN:
	start = a.start
	end = b.end
	c.TimeSpan = NewTimeSpan(start, end)

EXIT:
	return
}

func cellCollision(p PathAction, c Cell) (cc CellCollision) {
	cc.Path, cc.Cell = p, c
	switch c {
	case p.Dest:
		cc.CollisionType = CT_CELL_DEST
		cc.TimeSpan = p.TimeSpan
	case p.Orig:
		cc.CollisionType = CT_CELL_ORIG
		cc.TimeSpan = p.TimeSpan
	}
	return
}

func (c PathCollision) Type() CollisionType { return c.CollisionType }
func (c PathCollision) Start() WorldTime    { return c.TimeSpan.start }
func (c PathCollision) End() WorldTime      { return c.TimeSpan.end }
func (c PathCollision) OverlapAt(t WorldTime) (overlap float64) {

	switch c.CollisionType {
	case CT_SAME_ORIG_PERP:
		if t == c.end {
			overlap = 0.0
			return
		}

		p := [...]PartialCell{
			c.A.OrigPartial(t),
			c.B.OrigPartial(t),
		}

		overlap = p[0].Percentage * p[1].Percentage

	case CT_HEAD_TO_HEAD:
		if t == c.end {
			overlap = 1.0
			return
		}

		p := [...]PartialCell{
			c.A.DestPartial(t),
			c.B.DestPartial(t),
		}

		sum := p[0].Percentage + p[1].Percentage
		if sum > 1.0 {
			overlap = sum - 1.0
		}

	case CT_FROM_SIDE:
		if t == c.end {
			overlap = 1.0
			return
		}

		p := [...]PartialCell{
			c.A.DestPartial(t),
			c.B.DestPartial(t),
		}

		overlap = p[0].Percentage * p[1].Percentage

	case CT_SWAP:
		switch {
		case t <= c.start || t >= c.end:
			overlap = 0.0
		default:
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
func (c CellCollision) Start() WorldTime    { return c.TimeSpan.start }
func (c CellCollision) End() WorldTime      { return c.TimeSpan.end }
func (c CellCollision) OverlapAt(t WorldTime) (overlap float64) {
	switch c.CollisionType {
	case CT_CELL_DEST:
		overlap = c.Path.DestPartial(t).Percentage
	case CT_CELL_ORIG:
		overlap = c.Path.OrigPartial(t).Percentage
	}
	return
}
