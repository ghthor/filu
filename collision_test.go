package engine

import (
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
)

func overlapPeakAndDecrease(c gospec.Context, collision PathCollision) float64 {
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
	return peak
}

func overlapPeakLevelThenDecrease(c gospec.Context, collision PathCollision) {
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

	for ; t <= collision.A.end; t++ {
		overlap = collision.OverlapAt(t)
		c.Expect(overlap, IsWithin(0.0000000000000001), peak)
		prevOverlap = overlap
	}

	for ; t <= end; t++ {
		overlap = collision.OverlapAt(t)
		c.Expect(overlap, Satisfies, overlap < peak)
		c.Expect(overlap, Satisfies, overlap < prevOverlap)
		prevOverlap = overlap
	}

	overlap = collision.OverlapAt(end)
	c.Expect(overlap, Equals, 0.0)
}

func overlapGrowsTo1(c gospec.Context, collision PathCollision) {
	start, end := collision.Start(), collision.End()

	overlap := collision.OverlapAt(start)
	c.Expect(overlap, Equals, 0.0)

	prevOverlap := overlap

	t := start + 1
	for ; t < end; t++ {
		overlap = collision.OverlapAt(t)
		c.Expect(overlap, Satisfies, overlap > prevOverlap)
		prevOverlap = overlap
	}

	overlap = collision.OverlapAt(end)
	c.Expect(overlap, Equals, 1.0)
}

func overlapShrinksTo0(c gospec.Context, collision PathCollision) {
	start, end := collision.Start(), collision.End()

	overlap := collision.OverlapAt(start)
	c.Expect(overlap, Equals, 1.0)

	prevOverlap := overlap

	t := start + 1
	for ; t < end; t++ {
		overlap = collision.OverlapAt(t)
		c.Expect(overlap, Satisfies, overlap < prevOverlap)
		prevOverlap = overlap
	}

	overlap = collision.OverlapAt(end)
	c.Expect(overlap, Equals, 0.0)
}

func DescribePathCollision(c gospec.Context) {
	var pathA, pathB PathAction
	var collision PathCollision

	c.Specify("when path A is following into path B's position from the side", func() {
		pathA.TimeSpan = NewTimeSpan(15, 35)
		pathB.TimeSpan = NewTimeSpan(10, 30)

		pathA.Orig = Cell{0, 1}
		pathA.Dest = Cell{0, 0}

		pathB.Orig = pathA.Dest
		pathB.Dest = Cell{1, 0}

		collision = pathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_A_INTO_B_FROM_SIDE)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = pathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_A_INTO_B_FROM_SIDE)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		c.Specify("the collision will begin when", func() {
			specs := [...]struct {
				A, B        TimeSpan
				description string
			}{{
				NewTimeSpan(15, 35),
				NewTimeSpan(10, 30),
				"when path A starts if path A starts after path B",
			}, {
				NewTimeSpan(10, 30),
				NewTimeSpan(15, 35),
				"when path A starts if path A starts before path B",
			}, {
				NewTimeSpan(10, 30),
				NewTimeSpan(10, 30),
				"when path A and path B start at the same time",
			}}

			for _, spec := range specs {
				c.Specify(spec.description, func() {
					pathA.TimeSpan = spec.A
					pathB.TimeSpan = spec.B

					c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathA.start)
					c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathA.start)
				})
			}
		})

		c.Specify("the collision will end when", func() {
			specs := [...]struct {
				A, B        TimeSpan
				description string
			}{{
				NewTimeSpan(15, 35),
				NewTimeSpan(10, 30),
				"when path B ends if path B ends before path A",
			}, {
				NewTimeSpan(10, 30),
				NewTimeSpan(15, 35),
				"when path B ends if path B ends after path A",
			}, {
				NewTimeSpan(10, 30),
				NewTimeSpan(10, 30),
				"when path A and path B end at the same time",
			}}

			for _, spec := range specs {
				c.Specify(spec.description, func() {
					pathA.TimeSpan = spec.A
					pathB.TimeSpan = spec.B

					c.Expect(pathCollision(pathA, pathB).End(), Equals, pathB.end)
					c.Expect(pathCollision(pathB, pathA).End(), Equals, pathB.end)
				})
			}
		})
		specs := [...]struct {
			A, B        TimeSpan
			description string
		}{{
			NewTimeSpan(15, 35),
			NewTimeSpan(10, 30),
			"when path B ends before path A",
		}, {
			NewTimeSpan(10, 30),
			NewTimeSpan(15, 35),
			"when path B ends after path A",
		}, {
			NewTimeSpan(10, 30),
			NewTimeSpan(10, 30),
			"when path A and path B end at the same time",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" the overlap will grow to a peak and decrease till path B ends", func() {
				pathA.TimeSpan = spec.A
				pathB.TimeSpan = spec.B

				overlapPeakAndDecrease(c, pathCollision(pathA, pathB))
				overlapPeakAndDecrease(c, pathCollision(pathB, pathA))
			})
		}

		c.Specify("when B starts as A ends the peak will be 1.0", func() {
			pathA.TimeSpan = NewTimeSpan(10, 30)
			pathB.TimeSpan = NewTimeSpan(30, 40)

			peak := overlapPeakAndDecrease(c, pathCollision(pathA, pathB))
			c.Expect(peak, Equals, 1.0)

			peak = overlapPeakAndDecrease(c, pathCollision(pathB, pathA))
			c.Expect(peak, Equals, 1.0)
		})
	})

	c.Specify("when path A is following into path B's position in the same direction", func() {
		pathA.TimeSpan = NewTimeSpan(5, 25)
		pathB.TimeSpan = NewTimeSpan(10, 30)

		pathA.Orig = Cell{-1, 0}
		pathA.Dest = Cell{0, 0}

		pathB.Orig = pathA.Dest
		pathB.Dest = Cell{1, 0}

		collision = pathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_A_INTO_B)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = pathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_A_INTO_B)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		c.Specify("will not collide if path A doesn't end before path B", func() {
			pathA.end = pathB.end

			expectations := func() {
				collision = pathCollision(pathA, pathB)
				c.Expect(collision.Type(), Equals, CT_NONE)

				collision = pathCollision(pathB, pathA)
				c.Expect(collision.Type(), Equals, CT_NONE)
			}

			c.Specify("and starts at the same time as path B", func() {
				pathA.start = pathB.start
				expectations()
			})

			c.Specify("and starts after path B", func() {
				pathA.start = pathB.start + 1
				expectations()
			})
		})

		c.Specify("the collision will begin when", func() {
			c.Specify("path A starts if path A starts before path B", func() {
				pathA.start = pathB.start - 1

				collision = pathCollision(pathA, pathB)
				c.Expect(collision.Start(), Equals, pathA.start)

				c.Specify("and ends when path B completes", func() {
					c.Expect(collision.End(), Equals, pathB.end)
				})
			})

			c.Specify("they both start if they both start at the same time", func() {
				pathA.TimeSpan = NewTimeSpan(10, 20)
				pathB.TimeSpan = NewTimeSpan(10, 21)

				collision = pathCollision(pathA, pathB)
				c.Expect(collision.Start(), Equals, pathA.start)
				c.Expect(collision.Start(), Equals, pathB.start)

				c.Specify("and ends when path B completes", func() {
					c.Expect(collision.End(), Equals, pathB.end)
				})
			})

			c.Specify("path A catchs path B if path B starts before path A", func() {
				for i := WorldTime(0); i <= 5*10; i += 5 {
					pathA.TimeSpan = NewTimeSpan(2+i, 12+i)
					pathB.TimeSpan = NewTimeSpan(0+i, 20+i)
					collision = pathCollision(pathA, pathB)
					c.Expect(collision.Start(), Equals, WorldTime(4+i))
				}

				pathA.TimeSpan = NewTimeSpan(3, 10)
				pathB.TimeSpan = NewTimeSpan(2, 13)
				collision = pathCollision(pathA, pathB)
				c.Expect(collision.Start(), Equals, WorldTime(4))

				c.Specify("and ends when path B completes", func() {
					c.Expect(collision.End(), Equals, pathB.end)
				})
			})
		})

		specs := [...]struct {
			A, B        TimeSpan
			description string
		}{{
			pathA.TimeSpan,
			pathB.TimeSpan,
			"when path A starts before path B",
		}, {
			NewTimeSpan(10, 30),
			NewTimeSpan(10, 40),
			"when path A and path B start together",
		}, {
			NewTimeSpan(10, 30),
			NewTimeSpan(5, 40),
			"when path A starts after path B",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" the overlap will grow to a peak and then decrease", func() {
				pathA.TimeSpan = spec.A
				pathB.TimeSpan = spec.B

				overlapPeakAndDecrease(c, pathCollision(pathA, pathB))
				overlapPeakAndDecrease(c, pathCollision(pathB, pathA))
			})
		}

		c.Specify("they have the same speed the overlap will grow to a peak and stay level until B completes", func() {
			pathA.TimeSpan = NewTimeSpan(10, 30)
			pathB.TimeSpan = NewTimeSpan(15, 35)

			overlapPeakLevelThenDecrease(c, pathCollision(pathA, pathB))
			overlapPeakLevelThenDecrease(c, pathCollision(pathB, pathA))
		})
	})

	c.Specify("when path A has the same destination as path B and the paths are opposing", func() {
		m, n, o := Cell{-1, 0}, Cell{0, 0}, Cell{1, 0}
		pathA = PathAction{NewTimeSpan(10, 30), m, n}
		pathB = PathAction{NewTimeSpan(10, 30), o, n}

		collision = pathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_HEAD_TO_HEAD)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = pathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_HEAD_TO_HEAD)
		c.Assume(collision.A, Equals, pathB)
		c.Assume(collision.B, Equals, pathA)

		c.Specify("the collision begins when", func() {
			c.Specify("path A starts if path A starts when path B ends", func() {
				pathA.TimeSpan = NewTimeSpan(30, 50)
				pathB.TimeSpan = NewTimeSpan(10, 30)

				c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathA.start)
				c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathA.start)
			})
			c.Specify("path B starts if path B starts when path A ends", func() {
				pathA.TimeSpan = NewTimeSpan(10, 30)
				pathB.TimeSpan = NewTimeSpan(30, 50)

				c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathB.start)
				c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathB.start)
			})
		})

		c.Specify("the collision begins when path A meets path B", func() {
			pathA.TimeSpan = NewTimeSpan(0, 10)
			pathB.TimeSpan = NewTimeSpan(0, 10)
			c.Expect(pathCollision(pathA, pathB).Start(), Equals, WorldTime(5))
			c.Expect(pathCollision(pathB, pathA).Start(), Equals, WorldTime(5))

			pathA.TimeSpan = NewTimeSpan(2, 12)
			pathB.TimeSpan = NewTimeSpan(0, 10)
			c.Expect(pathCollision(pathA, pathB).Start(), Equals, WorldTime(6))
			c.Expect(pathCollision(pathB, pathA).Start(), Equals, WorldTime(6))

			pathA.TimeSpan = NewTimeSpan(3, 13)
			pathB.TimeSpan = NewTimeSpan(4, 14)
			// Float answer is 8.5
			c.Expect(pathCollision(pathA, pathB).Start(), Equals, WorldTime(8))
			c.Expect(pathCollision(pathB, pathA).Start(), Equals, WorldTime(8))

			pathA.TimeSpan = NewTimeSpan(3, 17)
			pathB.TimeSpan = NewTimeSpan(5, 11)
			// Float answer is 8 * 3/19
			c.Expect(pathCollision(pathA, pathB).Start(), Equals, WorldTime(8))
			c.Expect(pathCollision(pathB, pathA).Start(), Equals, WorldTime(8))
		})

		c.Specify("the collision ends when", func() {
			c.Specify("path A ends if it ends after Path B", func() {
				pathA.TimeSpan = NewTimeSpan(10, 30)
				pathB.TimeSpan = NewTimeSpan(5, 25)
				c.Expect(pathCollision(pathA, pathB).End(), Equals, pathA.end)
				c.Expect(pathCollision(pathB, pathA).End(), Equals, pathA.end)
			})

			c.Specify("path B ends if it ends after Path A", func() {
				pathA.TimeSpan = NewTimeSpan(5, 25)
				pathB.TimeSpan = NewTimeSpan(10, 30)
				c.Expect(pathCollision(pathA, pathB).End(), Equals, pathB.end)
				c.Expect(pathCollision(pathB, pathA).End(), Equals, pathB.end)
			})
		})

		specs := [...]struct {
			A, B        TimeSpan
			description string
		}{{
			pathA.TimeSpan,
			pathB.TimeSpan,
			"and path A and B have the same time span",
		}, {
			NewTimeSpan(5, 25),
			NewTimeSpan(15, 35),
			"and path A starts and ends before path B",
		}, {
			NewTimeSpan(15, 35),
			NewTimeSpan(5, 25),
			"and path B starts and ends before path A",
		}, {
			NewTimeSpan(5, 35),
			NewTimeSpan(15, 25),
			"and path A starts before and ends after path B",
		}, {
			NewTimeSpan(15, 25),
			NewTimeSpan(5, 35),
			"and path B starts before and ends after path A",
		}, {
			NewTimeSpan(5, 25),
			NewTimeSpan(10, 25),
			"and path A starts before and ends with path B",
		}, {
			NewTimeSpan(10, 25),
			NewTimeSpan(5, 25),
			"and path B starts before and ends with path A",
		}, {
			NewTimeSpan(10, 25),
			NewTimeSpan(10, 30),
			"and path A starts with and ends before path B",
		}, {
			NewTimeSpan(10, 30),
			NewTimeSpan(10, 25),
			"and path B starts with and ends before path A",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" overlap will grow from 0.0 to 1.0", func() {
				pathA.TimeSpan = spec.A
				pathB.TimeSpan = spec.B

				overlapGrowsTo1(c, pathCollision(pathA, pathB))
				overlapGrowsTo1(c, pathCollision(pathB, pathA))
			})
		}
	})

	c.Specify("when path A has the same destination as path B and the paths are perpendicular", func() {
		// Side 1
		m, n, o := Cell{-1, 0}, Cell{0, 0}, Cell{0, 1}
		pathA = PathAction{NewTimeSpan(10, 30), m, n}
		pathB = PathAction{NewTimeSpan(10, 30), o, n}

		collision = pathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_FROM_SIDE)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = pathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_FROM_SIDE)
		c.Assume(collision.A, Equals, pathB)
		c.Assume(collision.B, Equals, pathA)

		// Side 2
		m, n, o = Cell{-1, 0}, Cell{0, 0}, Cell{0, -1}
		pathA = PathAction{NewTimeSpan(10, 30), m, n}
		pathB = PathAction{NewTimeSpan(10, 30), o, n}

		collision = pathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_FROM_SIDE)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = pathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_FROM_SIDE)
		c.Assume(collision.A, Equals, pathB)
		c.Assume(collision.B, Equals, pathA)

		c.Specify("the collision begins when", func() {
			c.Specify("path A starts if path B starts first", func() {
				pathA.TimeSpan = NewTimeSpan(10, 30)
				pathB.TimeSpan = NewTimeSpan(5, 25)
				c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathA.start)
				c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathA.start)
			})

			c.Specify("path B starts if path A starts first", func() {
				pathA.TimeSpan = NewTimeSpan(5, 25)
				pathB.TimeSpan = NewTimeSpan(10, 30)
				c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathB.start)
				c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathB.start)
			})
		})

		c.Specify("the collision ends when", func() {
			c.Specify("path A ends if it ends after Path B", func() {
				pathA.TimeSpan = NewTimeSpan(10, 30)
				pathB.TimeSpan = NewTimeSpan(5, 25)
				c.Expect(pathCollision(pathA, pathB).End(), Equals, pathA.end)
				c.Expect(pathCollision(pathB, pathA).End(), Equals, pathA.end)
			})

			c.Specify("path B ends if it ends after Path A", func() {
				pathA.TimeSpan = NewTimeSpan(5, 25)
				pathB.TimeSpan = NewTimeSpan(10, 30)
				c.Expect(pathCollision(pathA, pathB).End(), Equals, pathB.end)
				c.Expect(pathCollision(pathB, pathA).End(), Equals, pathB.end)
			})
		})

		specs := [...]struct {
			A, B        TimeSpan
			description string
		}{{
			pathA.TimeSpan,
			pathB.TimeSpan,
			"and path A and B have the same time span",
		}, {
			NewTimeSpan(5, 25),
			NewTimeSpan(15, 35),
			"and path A starts and ends before path B",
		}, {
			NewTimeSpan(15, 35),
			NewTimeSpan(5, 25),
			"and path B starts and ends before path A",
		}, {
			NewTimeSpan(5, 35),
			NewTimeSpan(15, 25),
			"and path A starts before and ends after path B",
		}, {
			NewTimeSpan(15, 25),
			NewTimeSpan(5, 35),
			"and path B starts before and ends after path A",
		}, {
			NewTimeSpan(5, 25),
			NewTimeSpan(10, 25),
			"and path A starts before and ends with path B",
		}, {
			NewTimeSpan(10, 25),
			NewTimeSpan(5, 25),
			"and path B starts before and ends with path A",
		}, {
			NewTimeSpan(10, 25),
			NewTimeSpan(10, 30),
			"and path A starts with and ends before path B",
		}, {
			NewTimeSpan(10, 30),
			NewTimeSpan(10, 25),
			"and path B starts with and ends before path A",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" overlap will grow from 0.0 to 1.0", func() {
				pathA.TimeSpan = spec.A
				pathB.TimeSpan = spec.B

				overlapGrowsTo1(c, pathCollision(pathA, pathB))
				overlapGrowsTo1(c, pathCollision(pathB, pathA))
			})
		}
	})

	c.Specify("when path A and path B are inverses of each other", func() {
		a, b := Cell{0, 0}, Cell{1, 0}
		pathA = PathAction{Orig: a, Dest: b}
		pathB = PathAction{Orig: b, Dest: a}

		pathA.TimeSpan = NewTimeSpan(10, 30)
		pathB.TimeSpan = NewTimeSpan(10, 30)

		collision = pathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_SWAP)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = pathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_SWAP)
		c.Assume(collision.A, Equals, pathB)
		c.Assume(collision.B, Equals, pathA)

		c.Specify("the collision begins when either start moving", func() {
			c.Specify("path A starts first", func() {
				pathB.TimeSpan.start = pathA.TimeSpan.start + 1

				collision = pathCollision(pathA, pathB)
				c.Expect(collision.Start(), Equals, pathA.Start())

				collision = pathCollision(pathB, pathA)
				c.Expect(collision.Start(), Equals, pathA.Start())
			})

			c.Specify("path B starts first", func() {
				pathA.TimeSpan.start = pathB.TimeSpan.start + 1

				collision = pathCollision(pathA, pathB)
				c.Expect(collision.Start(), Equals, pathB.Start())

				collision = pathCollision(pathB, pathA)
				c.Expect(collision.Start(), Equals, pathB.Start())
			})
		})

		c.Specify("the collision ends when the last one moving ends", func() {
			c.Specify("path A ends last", func() {
				pathA.TimeSpan.end = pathB.TimeSpan.end + 1

				collision = pathCollision(pathA, pathB)
				c.Expect(collision.End(), Equals, pathA.End())

				collision = pathCollision(pathB, pathA)
				c.Expect(collision.End(), Equals, pathA.End())
			})

			c.Specify("path B ends last", func() {
				pathB.TimeSpan.end = pathA.TimeSpan.end + 1

				collision = pathCollision(pathA, pathB)
				c.Expect(collision.End(), Equals, pathB.End())

				collision = pathCollision(pathB, pathA)
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

		for _, spec := range specs {
			c.Specify(spec.description+" the overlap will grow to a peak and then decreases", func() {
				pathA.TimeSpan = spec.A
				pathB.TimeSpan = spec.B

				overlapPeakAndDecrease(c, pathCollision(pathA, pathB))
				overlapPeakAndDecrease(c, pathCollision(pathB, pathA))
			})
		}
	})

	c.Specify("when path A and path B have the same origin and their facings are inverse", func() {
		o := Cell{0, 0}
		pathA = PathAction{NewTimeSpan(10, 30), o, o.Neighbor(South)}
		pathB = PathAction{NewTimeSpan(10, 30), o, o.Neighbor(North)}

		c.Assume(pathCollision(pathA, pathB).Type(), Equals, CT_SAME_ORIG)
		c.Assume(pathCollision(pathB, pathA).Type(), Equals, CT_SAME_ORIG)

		c.Specify("the collision will begin when either path starts", func() {
			c.Specify("path A starts first", func() {
				pathA.TimeSpan = NewTimeSpan(5, 25)

				c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathA.start)
				c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathA.start)
			})

			c.Specify("path B starts first", func() {
				pathB.TimeSpan = NewTimeSpan(5, 25)

				c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathB.start)
				c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathB.start)
			})
		})

		c.Specify("the collision will end if one finishs as the other starts", func() {
			c.Specify("path A ends when path B starts", func() {
				pathA.TimeSpan = NewTimeSpan(10, 30)
				pathB.TimeSpan = NewTimeSpan(30, 40)
				c.Expect(pathCollision(pathA, pathB).End(), Equals, pathA.end)
				c.Expect(pathCollision(pathB, pathA).End(), Equals, pathA.end)
			})

			c.Specify("path B ends when path B starts", func() {
				pathA.TimeSpan = NewTimeSpan(30, 40)
				pathB.TimeSpan = NewTimeSpan(10, 30)
				c.Expect(pathCollision(pathA, pathB).End(), Equals, pathB.end)
				c.Expect(pathCollision(pathB, pathA).End(), Equals, pathB.end)
			})
		})

		c.Specify("the collision will end when the paths no longer overlap", func() {
			pathA.TimeSpan = NewTimeSpan(0, 10)
			pathB.TimeSpan = NewTimeSpan(0, 10)
			c.Expect(pathCollision(pathA, pathB).End(), Equals, WorldTime(5))
			c.Expect(pathCollision(pathB, pathA).End(), Equals, WorldTime(5))

			pathA.TimeSpan = NewTimeSpan(2, 12)
			pathB.TimeSpan = NewTimeSpan(0, 10)
			c.Expect(pathCollision(pathA, pathB).End(), Equals, WorldTime(6))
			c.Expect(pathCollision(pathB, pathA).End(), Equals, WorldTime(6))

			pathA.TimeSpan = NewTimeSpan(3, 13)
			pathB.TimeSpan = NewTimeSpan(4, 14)
			// Float answer is 8.5
			c.Expect(pathCollision(pathA, pathB).End(), Equals, WorldTime(9))
			c.Expect(pathCollision(pathB, pathA).End(), Equals, WorldTime(9))

			pathA.TimeSpan = NewTimeSpan(3, 17)
			pathB.TimeSpan = NewTimeSpan(5, 11)
			// Float answer is 8 * 3/19
			c.Expect(pathCollision(pathA, pathB).End(), Equals, WorldTime(9))
			c.Expect(pathCollision(pathB, pathA).End(), Equals, WorldTime(9))
		})

		specs := [...]struct {
			A, B        TimeSpan
			description string
		}{{
			NewTimeSpan(10, 30),
			NewTimeSpan(10, 30),
			"when path A and B start and end at the same time",
		}, {
			NewTimeSpan(10, 30),
			NewTimeSpan(30, 50),
			"when path A ends as path B starts",
		}, {
			NewTimeSpan(10, 30),
			NewTimeSpan(20, 40),
			"when path A starts and ends before path B",
		}, {
			NewTimeSpan(10, 30),
			NewTimeSpan(20, 30),
			"when path A starts before and ends with path B",
		}, {
			NewTimeSpan(10, 30),
			NewTimeSpan(20, 25),
			"when path A starts before and ends after path B",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" the overlap will start at 1.0 and decrease to 0.0", func() {
				pathA.TimeSpan = spec.A
				pathB.TimeSpan = spec.B
				overlapShrinksTo0(c, pathCollision(pathA, pathB))
				overlapShrinksTo0(c, pathCollision(pathB, pathA))
			})
		}
	})

	c.Specify("when path A and path B share the same origin and destination", func() {
		o, d := Cell{0, 0}, Cell{1, 0}

		pathA = PathAction{NewTimeSpan(10, 30), o, d}
		pathB = PathAction{NewTimeSpan(10, 30), o, d}

		c.Assume(pathCollision(pathA, pathB).Type(), Equals, CT_SAME_ORIG_DEST)
		c.Assume(pathCollision(pathB, pathA).Type(), Equals, CT_SAME_ORIG_DEST)

		c.Specify("the collision begins when either path starts", func() {
			c.Specify("path A starts first", func() {
				pathA.TimeSpan = NewTimeSpan(5, 25)

				c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathA.start)
				c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathA.start)
			})

			c.Specify("path B starts first", func() {
				pathB.TimeSpan = NewTimeSpan(5, 25)

				c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathB.start)
				c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathB.start)
			})
		})

		c.Specify("the collision ends when both paths have completed", func() {
			c.Specify("path A finishes last", func() {
				pathA.TimeSpan = NewTimeSpan(15, 35)

				c.Expect(pathCollision(pathA, pathB).End(), Equals, pathA.end)
				c.Expect(pathCollision(pathB, pathA).End(), Equals, pathA.end)
			})

			c.Specify("path B finishes last", func() {
				pathB.TimeSpan = NewTimeSpan(15, 35)

				c.Expect(pathCollision(pathA, pathB).End(), Equals, pathB.end)
				c.Expect(pathCollision(pathB, pathA).End(), Equals, pathB.end)
			})
		})

		c.Specify("the overlap will decrease to a trough and the grow back to 1.0", nil)
		c.Specify("the overlap will be 1.0 for the duration of the collision", nil)
	})

	c.Specify("when path A and path B share the same origin and are perpendicular", func() {
		o := Cell{0, 0}
		pathA = PathAction{NewTimeSpan(10, 30), o, o.Neighbor(North)}
		pathB = PathAction{NewTimeSpan(10, 30), o, o.Neighbor(East)}

		c.Specify("a same origin collision is identified", func() {
			c.Expect(pathCollision(pathA, pathB).Type(), Equals, CT_SAME_ORIG_PERP)
			c.Expect(pathCollision(pathB, pathA).Type(), Equals, CT_SAME_ORIG_PERP)

			pathB = PathAction{NewTimeSpan(10, 30), o, o.Neighbor(West)}
			c.Expect(pathCollision(pathA, pathB).Type(), Equals, CT_SAME_ORIG_PERP)
			c.Expect(pathCollision(pathB, pathA).Type(), Equals, CT_SAME_ORIG_PERP)
		})

		c.Specify("the collision begins when either path starts first", func() {
			c.Specify("path A starts first", func() {
				pathA.TimeSpan = NewTimeSpan(5, 25)

				c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathA.start)
				c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathA.start)
			})

			c.Specify("path B starts first", func() {
				pathB.TimeSpan = NewTimeSpan(5, 25)

				c.Expect(pathCollision(pathA, pathB).Start(), Equals, pathB.start)
				c.Expect(pathCollision(pathB, pathA).Start(), Equals, pathB.start)
			})
		})

		c.Specify("the collision ends when the first one finishes", func() {
			c.Specify("path A finishes first", func() {
				pathA.TimeSpan = NewTimeSpan(5, 25)

				c.Expect(pathCollision(pathA, pathB).End(), Equals, pathA.end)
				c.Expect(pathCollision(pathB, pathA).End(), Equals, pathA.end)
			})

			c.Specify("path B finishes first", func() {
				pathB.TimeSpan = NewTimeSpan(5, 25)

				c.Expect(pathCollision(pathA, pathB).End(), Equals, pathB.end)
				c.Expect(pathCollision(pathB, pathA).End(), Equals, pathB.end)
			})
		})

		c.Specify("the overlap will start at 1.0 and decrease to 0.0", func() {
			overlapShrinksTo0(c, pathCollision(pathA, pathB))
			overlapShrinksTo0(c, pathCollision(pathB, pathA))
		})
	})
}

func DescribeCellCollision(c gospec.Context) {
	c.Specify("a cell-path collision can be calculated", func() {
		c.Specify("as not happening", func() {
			cell := Cell{0, 0}
			path := PathAction{
				TimeSpan: NewTimeSpan(10, 20),
				Orig:     Cell{1, 1},
				Dest:     Cell{1, 0},
			}
			collision := path.CollidesWith(cell)
			c.Expect(collision.Type(), Equals, CT_NONE)
		})

		c.Specify("if the path's origin is the cell", func() {
			cell := Cell{0, 0}
			path := PathAction{
				TimeSpan: NewTimeSpan(10, 20),
				Orig:     cell,
				Dest:     Cell{1, 0},
			}
			collision := path.CollidesWith(cell)
			c.Assume(collision.Type(), Equals, CT_CELL_ORIG)
			c.Assume(collision.(CellCollision).Cell, Equals, cell)
			c.Assume(collision.(CellCollision).Path, Equals, path)

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

		c.Specify("if the path's destination is the cell", func() {
			cell := Cell{0, 0}
			path := PathAction{
				TimeSpan: NewTimeSpan(10, 30),
				Orig:     Cell{1, 0},
				Dest:     cell,
			}
			collision := path.CollidesWith(cell)
			c.Assume(collision.Type(), Equals, CT_CELL_DEST)
			c.Assume(collision.(CellCollision).Cell, Equals, cell)
			c.Assume(collision.(CellCollision).Path, Equals, path)

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
