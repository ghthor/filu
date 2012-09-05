package engine

import (
	"math"
)

type (
	CollisionType int

	// This will be renamed to Collision
	CollisionTemp interface {
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

	CoordCollision struct {
		CollisionType
		TimeSpan
		Coord WorldCoord
		Path  PathAction
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
	CT_SAME_ORIG_DEST     CollisionType = iota
	CT_COORD_DEST         CollisionType = iota
	CT_COORD_ORIG         CollisionType = iota
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
	case CT_COORD_DEST:
		return "coord destination"
	case CT_COORD_ORIG:
		return "coord origin"
	}
	return "unknown collision type"
}

func (A PathAction) CollidesWith(B interface{}) (c CollisionTemp) {
	switch b := B.(type) {
	case PathAction:
		return pathCollision(A, b)
	case WorldCoord:
		return coordCollision(A, b)
	default:
	}
	panic("unknown collision attempt")
}

func pathCollision(a, b PathAction) (c PathCollision) {
	var start, end WorldTime
	c.A, c.B = a, b

	switch {
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
		// A is moving into the WorldCoord B is leaving
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

func coordCollision(p PathAction, wc WorldCoord) (c CoordCollision) {
	c.Path, c.Coord = p, wc
	switch wc {
	case p.Dest:
		c.CollisionType = CT_COORD_DEST
		c.TimeSpan = p.TimeSpan
	case p.Orig:
		c.CollisionType = CT_COORD_ORIG
		c.TimeSpan = p.TimeSpan
	}
	return
}

func (c PathCollision) Type() CollisionType { return c.CollisionType }
func (c PathCollision) Start() WorldTime    { return c.TimeSpan.start }
func (c PathCollision) End() WorldTime      { return c.TimeSpan.end }
func (c PathCollision) OverlapAt(t WorldTime) (overlap float64) {

	switch c.CollisionType {
	case CT_SWAP:
		switch {
		case t <= c.start || t >= c.end:
			overlap = 0.0
		default:
			p := [...]PartialWorldCoord{
				c.A.DestPartial(t),
				c.B.DestPartial(t),
			}

			overlap = p[0].Percentage + p[1].Percentage

			if overlap > 1.0 {
				p = [...]PartialWorldCoord{
					c.A.OrigPartial(t),
					c.B.OrigPartial(t),
				}
				overlap = p[0].Percentage + p[1].Percentage
			}
		}

	case CT_A_INTO_B:
		p := [...]PartialWorldCoord{
			c.A.DestPartial(t),
			c.B.OrigPartial(t),
		}

		sum := p[0].Percentage + p[1].Percentage
		if sum > 1.0 {
			overlap = sum - 1.0
		}

	case CT_A_INTO_B_FROM_SIDE:
		p := [...]PartialWorldCoord{
			c.A.DestPartial(t),
			c.B.OrigPartial(t),
		}

		overlap = p[0].Percentage * p[1].Percentage
	}
	return
}

func (c CoordCollision) Type() CollisionType { return c.CollisionType }
func (c CoordCollision) Start() WorldTime    { return c.TimeSpan.start }
func (c CoordCollision) End() WorldTime      { return c.TimeSpan.end }
func (c CoordCollision) OverlapAt(t WorldTime) (overlap float64) {
	switch c.CollisionType {
	case CT_COORD_DEST:
		overlap = c.Path.DestPartial(t).Percentage
	case CT_COORD_ORIG:
		overlap = c.Path.OrigPartial(t).Percentage
	}
	return
}
