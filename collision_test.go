package engine

import (
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
)

func DescribePathCollision(c gospec.Context) {
	var pathA, pathB PathAction
	var collision CollisionTemp

	c.Specify("when path A is following into path B's position", func() {
	})

	c.Specify("when path A has the same destination as path B", func() {
	})

	c.Specify("when path A and path B are inverses of each other", func() {
		a, b := WorldCoord{0, 0}, WorldCoord{1, 0}
		pathA = PathAction{Orig: a, Dest: b}
		pathB = PathAction{Orig: b, Dest: a}

		pathA.TimeSpan = NewTimeSpan(10, 30)
		pathB.TimeSpan = NewTimeSpan(10, 30)

		collision = pathA.CollidesWith(pathB)
		c.Assume(collision.Type(), Equals, CT_SWAP)
		c.Assume(collision.(PathCollision).A, Equals, pathA)
		c.Assume(collision.(PathCollision).B, Equals, pathB)

		collision = pathB.CollidesWith(pathA)
		c.Assume(collision.Type(), Equals, CT_SWAP)
		c.Assume(collision.(PathCollision).A, Equals, pathB)
		c.Assume(collision.(PathCollision).B, Equals, pathA)

		c.Specify("the collision begins when either start moving", func() {
			c.Specify("path A starts first", func() {
				pathB.TimeSpan.start = pathA.TimeSpan.start + 1

				collision = pathA.CollidesWith(pathB)
				c.Expect(collision.Start(), Equals, pathA.Start())

				collision = pathB.CollidesWith(pathA)
				c.Expect(collision.Start(), Equals, pathA.Start())
			})

			c.Specify("path B starts first", func() {
				pathA.TimeSpan.start = pathB.TimeSpan.start + 1

				collision = pathA.CollidesWith(pathB)
				c.Expect(collision.Start(), Equals, pathB.Start())

				collision = pathB.CollidesWith(pathA)
				c.Expect(collision.Start(), Equals, pathB.Start())
			})
		})

		c.Specify("the collision ends when the last one moving ends", func() {
			c.Specify("path A ends last", func() {
				pathA.TimeSpan.end = pathB.TimeSpan.end + 1

				collision = pathA.CollidesWith(pathB)
				c.Expect(collision.End(), Equals, pathA.End())

				collision = pathB.CollidesWith(pathA)
				c.Expect(collision.End(), Equals, pathA.End())
			})

			c.Specify("path B ends last", func() {
				pathB.TimeSpan.end = pathA.TimeSpan.end + 1

				collision = pathA.CollidesWith(pathB)
				c.Expect(collision.End(), Equals, pathB.End())

				collision = pathB.CollidesWith(pathA)
				c.Expect(collision.End(), Equals, pathB.End())
			})
		})

		specs := [...]struct {
			A, B        TimeSpan
			description string
		}{{
			pathA.TimeSpan,
			pathB.TimeSpan,
			"and the time span's are equal",
		}, {
			NewTimeSpan(10, 90),
			NewTimeSpan(31, 69),
			"and one starts before and ends after the other",
		}, {
			NewTimeSpan(28, 30),
			NewTimeSpan(30, 32),
			"and one starts as the other ends",
		}, {
			NewTimeSpan(20, 30),
			NewTimeSpan(40, 50),
			"and one starts after the other ends",
		}}

		overlapSpec := func(collision PathCollision) {
			start, end := collision.Start(), collision.End()

			// Going to fix this requirement when I implement floating point time
			// The collision needs > 3 steps and 2 of them have to be after the peak
			// This is the easiest way to enforce this
			c.Assume(collision.A.duration, Satisfies, collision.A.duration > 1)
			c.Assume(collision.B.duration, Satisfies, collision.B.duration > 1)

			overlap := collision.OverlapAt(start)
			c.Expect(overlap, Equals, 0.0)

			prevOverlap := overlap
			peak := overlap

			t := start + 1
			for ; t < end; t++ {
				overlap = collision.OverlapAt(t)
				if overlap <= prevOverlap || t == end-1 {
					peak = prevOverlap
					prevOverlap = overlap
					break
				}
				c.Expect(overlap, Satisfies, overlap > prevOverlap)
				prevOverlap = overlap
			}

			c.Expect(peak, Satisfies, peak > collision.OverlapAt(start))

			for ; t <= end; t++ {
				overlap = collision.OverlapAt(t)
				c.Expect(overlap, Satisfies, overlap <= peak)
				c.Expect(overlap, Satisfies, overlap <= prevOverlap)
				prevOverlap = overlap
			}

			overlap = collision.OverlapAt(end)
			c.Expect(overlap, Equals, 0.0)
		}

		for _, spec := range specs {
			c.Specify(spec.description+" the overlap will grow to a peak and then decreases", func() {
				pathA.TimeSpan = spec.A
				pathB.TimeSpan = spec.B

				overlapSpec(pathCollision(pathA, pathB))
				overlapSpec(pathCollision(pathB, pathA))
			})
		}
	})
}

func DescribeCoordCollision(c gospec.Context) {
	c.Specify("a coord-path collision can be calculated", func() {
		c.Specify("as not happening", func() {
			coord := WorldCoord{0, 0}
			path := PathAction{
				TimeSpan: NewTimeSpan(10, 20),
				Orig:     WorldCoord{1, 1},
				Dest:     WorldCoord{1, 0},
			}
			collision := path.CollidesWith(coord)
			c.Expect(collision.Type(), Equals, CT_NONE)
		})

		c.Specify("if the path's origin is the coord", func() {
			coord := WorldCoord{0, 0}
			path := PathAction{
				TimeSpan: NewTimeSpan(10, 20),
				Orig:     coord,
				Dest:     WorldCoord{1, 0},
			}
			collision := path.CollidesWith(coord)
			c.Assume(collision.Type(), Equals, CT_COORD_ORIG)
			c.Assume(collision.(CoordCollision).Coord, Equals, coord)
			c.Assume(collision.(CoordCollision).Path, Equals, path)

			c.Specify("the overlap will begin at 1.0 and decrease to 0.0", func() {
				start, end := collision.Start(), collision.End()

				overlap := collision.OverlapAt(start)
				c.Expect(overlap, Equals, 1.0)

				prevOverlap := overlap
				for i := start + 1; i < end; i++ {
					overlap = collision.OverlapAt(i)
					c.Expect(overlap, Satisfies, overlap < prevOverlap)
					prevOverlap = overlap
				}

				c.Expect(collision.OverlapAt(end), Equals, 0.0)
			})
		})

		c.Specify("if the path's destination is the coord", func() {
			coord := WorldCoord{0, 0}
			path := PathAction{
				TimeSpan: NewTimeSpan(10, 30),
				Orig:     WorldCoord{1, 0},
				Dest:     coord,
			}
			collision := path.CollidesWith(coord)
			c.Assume(collision.Type(), Equals, CT_COORD_DEST)
			c.Assume(collision.(CoordCollision).Coord, Equals, coord)
			c.Assume(collision.(CoordCollision).Path, Equals, path)

			c.Specify("the overlap will begin at 0.0 and grow to 1.0", func() {
				start, end := collision.Start(), collision.End()

				overlap := collision.OverlapAt(start)
				c.Expect(overlap, Equals, 0.0)

				prevOverlap := overlap
				for i := start + 1; i < end; i++ {
					overlap = collision.OverlapAt(i)
					c.Expect(overlap, Satisfies, overlap > prevOverlap)
					prevOverlap = overlap
				}

				c.Expect(collision.OverlapAt(end), Equals, 1.0)
			})
		})
	})
}
