package coord

import (
	"github.com/ghthor/filu/sim/stime"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func overlapPeakAndDecrease(c gospec.Context, collision PathCollision) float64 {
	start, end := collision.Start(), collision.End()

	// Going to fix this requirement when I implement floating point time
	// The collision needs > 3 steps and 2 of them have to be after the peak
	// This is the easiest way to enforce this
	c.Assume(collision.A.Duration, Satisfies, collision.A.Duration > 1)
	c.Assume(collision.B.Duration, Satisfies, collision.B.Duration > 1)

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
	c.Assume(collision.A.Duration, Satisfies, collision.A.Duration > 1)
	c.Assume(collision.B.Duration, Satisfies, collision.B.Duration > 1)

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

	for ; t <= collision.A.Span.End; t++ {
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
		pathA.Span = stime.NewSpan(15, 35)
		pathB.Span = stime.NewSpan(10, 30)

		pathA.Orig = Cell{0, 1}
		pathA.Dest = Cell{0, 0}

		pathB.Orig = pathA.Dest
		pathB.Dest = Cell{1, 0}

		collision = NewPathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_A_INTO_B_FROM_SIDE)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = NewPathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_A_INTO_B_FROM_SIDE)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		c.Specify("the collision will begin when", func() {
			specs := [...]struct {
				A, B        stime.Span
				description string
			}{{
				stime.NewSpan(15, 35),
				stime.NewSpan(10, 30),
				"when path A starts if path A starts after path B",
			}, {
				stime.NewSpan(10, 30),
				stime.NewSpan(15, 35),
				"when path A starts if path A starts before path B",
			}, {
				stime.NewSpan(10, 30),
				stime.NewSpan(10, 30),
				"when path A and path B start at the same time",
			}}

			for _, spec := range specs {
				c.Specify(spec.description, func() {
					pathA.Span = spec.A
					pathB.Span = spec.B

					c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathA.Span.Start)
					c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathA.Span.Start)
				})
			}
		})

		c.Specify("the collision will end when", func() {
			specs := [...]struct {
				A, B        stime.Span
				description string
			}{{
				stime.NewSpan(15, 35),
				stime.NewSpan(10, 30),
				"when path B ends if path B ends before path A",
			}, {
				stime.NewSpan(10, 30),
				stime.NewSpan(15, 35),
				"when path B ends if path B ends after path A",
			}, {
				stime.NewSpan(10, 30),
				stime.NewSpan(10, 30),
				"when path A and path B end at the same time",
			}}

			for _, spec := range specs {
				c.Specify(spec.description, func() {
					pathA.Span = spec.A
					pathB.Span = spec.B

					c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathB.Span.End)
					c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathB.Span.End)
				})
			}
		})
		specs := [...]struct {
			A, B        stime.Span
			description string
		}{{
			stime.NewSpan(15, 35),
			stime.NewSpan(10, 30),
			"when path B ends before path A",
		}, {
			stime.NewSpan(10, 30),
			stime.NewSpan(15, 35),
			"when path B ends after path A",
		}, {
			stime.NewSpan(10, 30),
			stime.NewSpan(10, 30),
			"when path A and path B end at the same time",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" the overlap will grow to a peak and decrease till path B ends", func() {
				pathA.Span = spec.A
				pathB.Span = spec.B

				overlapPeakAndDecrease(c, NewPathCollision(pathA, pathB))
				overlapPeakAndDecrease(c, NewPathCollision(pathB, pathA))
			})
		}

		c.Specify("when B starts as A ends the peak will be 1.0", func() {
			pathA.Span = stime.NewSpan(10, 30)
			pathB.Span = stime.NewSpan(30, 40)

			peak := overlapPeakAndDecrease(c, NewPathCollision(pathA, pathB))
			c.Expect(peak, Equals, 1.0)

			peak = overlapPeakAndDecrease(c, NewPathCollision(pathB, pathA))
			c.Expect(peak, Equals, 1.0)
		})
	})

	c.Specify("when path A is following into path B's position in the same direction", func() {
		pathA.Span = stime.NewSpan(5, 25)
		pathB.Span = stime.NewSpan(10, 30)

		pathA.Orig = Cell{-1, 0}
		pathA.Dest = Cell{0, 0}

		pathB.Orig = pathA.Dest
		pathB.Dest = Cell{1, 0}

		collision = NewPathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_A_INTO_B)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = NewPathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_A_INTO_B)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		c.Specify("will not collide if path A doesn't end before path B", func() {
			pathA.Span.End = pathB.Span.End

			expectations := func() {
				collision = NewPathCollision(pathA, pathB)
				c.Expect(collision.Type(), Equals, CT_NONE)

				collision = NewPathCollision(pathB, pathA)
				c.Expect(collision.Type(), Equals, CT_NONE)
			}

			c.Specify("and starts at the same time as path B", func() {
				pathA.Span.Start = pathB.Span.Start
				expectations()
			})

			c.Specify("and starts after path B", func() {
				pathA.Span.Start = pathB.Span.Start + 1
				expectations()
			})
		})

		c.Specify("the collision will begin when", func() {
			c.Specify("path A starts if path A starts before path B", func() {
				pathA.Span.Start = pathB.Span.Start - 1

				collision = NewPathCollision(pathA, pathB)
				c.Expect(collision.Start(), Equals, pathA.Span.Start)

				c.Specify("and ends when path B completes", func() {
					c.Expect(collision.End(), Equals, pathB.Span.End)
				})
			})

			c.Specify("they both start if they both start at the same time", func() {
				pathA.Span = stime.NewSpan(10, 20)
				pathB.Span = stime.NewSpan(10, 21)

				collision = NewPathCollision(pathA, pathB)
				c.Expect(collision.Start(), Equals, pathA.Span.Start)
				c.Expect(collision.Start(), Equals, pathB.Span.Start)

				c.Specify("and ends when path B completes", func() {
					c.Expect(collision.End(), Equals, pathB.Span.End)
				})
			})

			c.Specify("path A catchs path B if path B starts before path A", func() {
				for i := stime.Time(0); i <= 5*10; i += 5 {
					pathA.Span = stime.NewSpan(2+i, 12+i)
					pathB.Span = stime.NewSpan(0+i, 20+i)
					collision = NewPathCollision(pathA, pathB)
					c.Expect(collision.Start(), Equals, stime.Time(4+i))
				}

				pathA.Span = stime.NewSpan(3, 10)
				pathB.Span = stime.NewSpan(2, 13)
				collision = NewPathCollision(pathA, pathB)
				c.Expect(collision.Start(), Equals, stime.Time(4))

				c.Specify("and ends when path B completes", func() {
					c.Expect(collision.End(), Equals, pathB.Span.End)
				})
			})
		})

		specs := [...]struct {
			A, B        stime.Span
			description string
		}{{
			pathA.Span,
			pathB.Span,
			"when path A starts before path B",
		}, {
			stime.NewSpan(10, 30),
			stime.NewSpan(10, 40),
			"when path A and path B start together",
		}, {
			stime.NewSpan(10, 30),
			stime.NewSpan(5, 40),
			"when path A starts after path B",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" the overlap will grow to a peak and then decrease", func() {
				pathA.Span = spec.A
				pathB.Span = spec.B

				overlapPeakAndDecrease(c, NewPathCollision(pathA, pathB))
				overlapPeakAndDecrease(c, NewPathCollision(pathB, pathA))
			})
		}

		c.Specify("they have the same speed the overlap will grow to a peak and stay level until B completes", func() {
			pathA.Span = stime.NewSpan(10, 30)
			pathB.Span = stime.NewSpan(15, 35)

			overlapPeakLevelThenDecrease(c, NewPathCollision(pathA, pathB))
			overlapPeakLevelThenDecrease(c, NewPathCollision(pathB, pathA))
		})
	})

	c.Specify("when path A has the same destination as path B and the paths are opposing", func() {
		m, n, o := Cell{-1, 0}, Cell{0, 0}, Cell{1, 0}
		pathA = PathAction{stime.NewSpan(10, 30), m, n}
		pathB = PathAction{stime.NewSpan(10, 30), o, n}

		collision = NewPathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_HEAD_TO_HEAD)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = NewPathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_HEAD_TO_HEAD)
		c.Assume(collision.A, Equals, pathB)
		c.Assume(collision.B, Equals, pathA)

		c.Specify("the collision begins when", func() {
			c.Specify("path A starts if path A starts when path B ends", func() {
				pathA.Span = stime.NewSpan(30, 50)
				pathB.Span = stime.NewSpan(10, 30)

				c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathA.Span.Start)
				c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathA.Span.Start)
			})
			c.Specify("path B starts if path B starts when path A ends", func() {
				pathA.Span = stime.NewSpan(10, 30)
				pathB.Span = stime.NewSpan(30, 50)

				c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathB.Span.Start)
				c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathB.Span.Start)
			})
		})

		c.Specify("the collision begins when path A meets path B", func() {
			pathA.Span = stime.NewSpan(0, 10)
			pathB.Span = stime.NewSpan(0, 10)
			c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, stime.Time(5))
			c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, stime.Time(5))

			pathA.Span = stime.NewSpan(2, 12)
			pathB.Span = stime.NewSpan(0, 10)
			c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, stime.Time(6))
			c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, stime.Time(6))

			pathA.Span = stime.NewSpan(3, 13)
			pathB.Span = stime.NewSpan(4, 14)
			// Float answer is 8.5
			c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, stime.Time(8))
			c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, stime.Time(8))

			pathA.Span = stime.NewSpan(3, 17)
			pathB.Span = stime.NewSpan(5, 11)
			// Float answer is 8 * 3/19
			c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, stime.Time(8))
			c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, stime.Time(8))
		})

		c.Specify("the collision ends when", func() {
			c.Specify("path A ends if it ends after Path B", func() {
				pathA.Span = stime.NewSpan(10, 30)
				pathB.Span = stime.NewSpan(5, 25)
				c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathA.Span.End)
				c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathA.Span.End)
			})

			c.Specify("path B ends if it ends after Path A", func() {
				pathA.Span = stime.NewSpan(5, 25)
				pathB.Span = stime.NewSpan(10, 30)
				c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathB.Span.End)
				c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathB.Span.End)
			})
		})

		specs := [...]struct {
			A, B        stime.Span
			description string
		}{{
			pathA.Span,
			pathB.Span,
			"and path A and B have the same time span",
		}, {
			stime.NewSpan(5, 25),
			stime.NewSpan(15, 35),
			"and path A starts and ends before path B",
		}, {
			stime.NewSpan(15, 35),
			stime.NewSpan(5, 25),
			"and path B starts and ends before path A",
		}, {
			stime.NewSpan(5, 35),
			stime.NewSpan(15, 25),
			"and path A starts before and ends after path B",
		}, {
			stime.NewSpan(15, 25),
			stime.NewSpan(5, 35),
			"and path B starts before and ends after path A",
		}, {
			stime.NewSpan(5, 25),
			stime.NewSpan(10, 25),
			"and path A starts before and ends with path B",
		}, {
			stime.NewSpan(10, 25),
			stime.NewSpan(5, 25),
			"and path B starts before and ends with path A",
		}, {
			stime.NewSpan(10, 25),
			stime.NewSpan(10, 30),
			"and path A starts with and ends before path B",
		}, {
			stime.NewSpan(10, 30),
			stime.NewSpan(10, 25),
			"and path B starts with and ends before path A",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" overlap will grow from 0.0 to 1.0", func() {
				pathA.Span = spec.A
				pathB.Span = spec.B

				overlapGrowsTo1(c, NewPathCollision(pathA, pathB))
				overlapGrowsTo1(c, NewPathCollision(pathB, pathA))
			})
		}
	})

	c.Specify("when path A has the same destination as path B and the paths are perpendicular", func() {
		// Side 1
		m, n, o := Cell{-1, 0}, Cell{0, 0}, Cell{0, 1}
		pathA = PathAction{stime.NewSpan(10, 30), m, n}
		pathB = PathAction{stime.NewSpan(10, 30), o, n}

		collision = NewPathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_FROM_SIDE)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = NewPathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_FROM_SIDE)
		c.Assume(collision.A, Equals, pathB)
		c.Assume(collision.B, Equals, pathA)

		// Side 2
		m, n, o = Cell{-1, 0}, Cell{0, 0}, Cell{0, -1}
		pathA = PathAction{stime.NewSpan(10, 30), m, n}
		pathB = PathAction{stime.NewSpan(10, 30), o, n}

		collision = NewPathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_FROM_SIDE)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = NewPathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_FROM_SIDE)
		c.Assume(collision.A, Equals, pathB)
		c.Assume(collision.B, Equals, pathA)

		c.Specify("the collision begins when", func() {
			c.Specify("path A starts if path B starts first", func() {
				pathA.Span = stime.NewSpan(10, 30)
				pathB.Span = stime.NewSpan(5, 25)
				c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathA.Span.Start)
				c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathA.Span.Start)
			})

			c.Specify("path B starts if path A starts first", func() {
				pathA.Span = stime.NewSpan(5, 25)
				pathB.Span = stime.NewSpan(10, 30)
				c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathB.Span.Start)
				c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathB.Span.Start)
			})
		})

		c.Specify("the collision ends when", func() {
			c.Specify("path A ends if it ends after Path B", func() {
				pathA.Span = stime.NewSpan(10, 30)
				pathB.Span = stime.NewSpan(5, 25)
				c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathA.Span.End)
				c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathA.Span.End)
			})

			c.Specify("path B ends if it ends after Path A", func() {
				pathA.Span = stime.NewSpan(5, 25)
				pathB.Span = stime.NewSpan(10, 30)
				c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathB.Span.End)
				c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathB.Span.End)
			})
		})

		specs := [...]struct {
			A, B        stime.Span
			description string
		}{{
			pathA.Span,
			pathB.Span,
			"and path A and B have the same time span",
		}, {
			stime.NewSpan(5, 25),
			stime.NewSpan(15, 35),
			"and path A starts and ends before path B",
		}, {
			stime.NewSpan(15, 35),
			stime.NewSpan(5, 25),
			"and path B starts and ends before path A",
		}, {
			stime.NewSpan(5, 35),
			stime.NewSpan(15, 25),
			"and path A starts before and ends after path B",
		}, {
			stime.NewSpan(15, 25),
			stime.NewSpan(5, 35),
			"and path B starts before and ends after path A",
		}, {
			stime.NewSpan(5, 25),
			stime.NewSpan(10, 25),
			"and path A starts before and ends with path B",
		}, {
			stime.NewSpan(10, 25),
			stime.NewSpan(5, 25),
			"and path B starts before and ends with path A",
		}, {
			stime.NewSpan(10, 25),
			stime.NewSpan(10, 30),
			"and path A starts with and ends before path B",
		}, {
			stime.NewSpan(10, 30),
			stime.NewSpan(10, 25),
			"and path B starts with and ends before path A",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" overlap will grow from 0.0 to 1.0", func() {
				pathA.Span = spec.A
				pathB.Span = spec.B

				overlapGrowsTo1(c, NewPathCollision(pathA, pathB))
				overlapGrowsTo1(c, NewPathCollision(pathB, pathA))
			})
		}
	})

	c.Specify("when path A and path B are inverses of each other", func() {
		a, b := Cell{0, 0}, Cell{1, 0}
		pathA = PathAction{Orig: a, Dest: b}
		pathB = PathAction{Orig: b, Dest: a}

		pathA.Span = stime.NewSpan(10, 30)
		pathB.Span = stime.NewSpan(10, 30)

		collision = NewPathCollision(pathA, pathB)
		c.Assume(collision.Type(), Equals, CT_SWAP)
		c.Assume(collision.A, Equals, pathA)
		c.Assume(collision.B, Equals, pathB)

		collision = NewPathCollision(pathB, pathA)
		c.Assume(collision.Type(), Equals, CT_SWAP)
		c.Assume(collision.A, Equals, pathB)
		c.Assume(collision.B, Equals, pathA)

		c.Specify("the collision begins when either start moving", func() {
			c.Specify("path A starts first", func() {
				pathB.Span.Start = pathA.Span.Start + 1

				collision = NewPathCollision(pathA, pathB)
				c.Expect(collision.Start(), Equals, pathA.Start())

				collision = NewPathCollision(pathB, pathA)
				c.Expect(collision.Start(), Equals, pathA.Start())
			})

			c.Specify("path B starts first", func() {
				pathA.Span.Start = pathB.Span.Start + 1

				collision = NewPathCollision(pathA, pathB)
				c.Expect(collision.Start(), Equals, pathB.Start())

				collision = NewPathCollision(pathB, pathA)
				c.Expect(collision.Start(), Equals, pathB.Start())
			})
		})

		c.Specify("the collision ends when the last one moving ends", func() {
			c.Specify("path A ends last", func() {
				pathA.Span.End = pathB.Span.End + 1

				collision = NewPathCollision(pathA, pathB)
				c.Expect(collision.End(), Equals, pathA.End())

				collision = NewPathCollision(pathB, pathA)
				c.Expect(collision.End(), Equals, pathA.End())
			})

			c.Specify("path B ends last", func() {
				pathB.Span.End = pathA.Span.End + 1

				collision = NewPathCollision(pathA, pathB)
				c.Expect(collision.End(), Equals, pathB.End())

				collision = NewPathCollision(pathB, pathA)
				c.Expect(collision.End(), Equals, pathB.End())
			})
		})

		specs := [...]struct {
			A, B        stime.Span
			description string
		}{{
			pathA.Span,
			pathB.Span,
			"and the time span's are equal",
		}, {
			stime.NewSpan(10, 90),
			stime.NewSpan(31, 69),
			"and one starts before and ends after the other",
		}, {
			stime.NewSpan(28, 30),
			stime.NewSpan(30, 32),
			"and one starts as the other ends",
		}, {
			stime.NewSpan(20, 30),
			stime.NewSpan(40, 50),
			"and one starts after the other ends",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" the overlap will grow to a peak and then decreases", func() {
				pathA.Span = spec.A
				pathB.Span = spec.B

				overlapPeakAndDecrease(c, NewPathCollision(pathA, pathB))
				overlapPeakAndDecrease(c, NewPathCollision(pathB, pathA))
			})
		}
	})

	c.Specify("when path A and path B have the same origin and their facings are inverse", func() {
		o := Cell{0, 0}
		pathA = PathAction{stime.NewSpan(10, 30), o, o.Neighbor(South)}
		pathB = PathAction{stime.NewSpan(10, 30), o, o.Neighbor(North)}

		c.Assume(NewPathCollision(pathA, pathB).Type(), Equals, CT_SAME_ORIG)
		c.Assume(NewPathCollision(pathB, pathA).Type(), Equals, CT_SAME_ORIG)

		c.Specify("the collision will begin when either path starts", func() {
			c.Specify("path A starts first", func() {
				pathA.Span = stime.NewSpan(5, 25)

				c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathA.Span.Start)
				c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathA.Span.Start)
			})

			c.Specify("path B starts first", func() {
				pathB.Span = stime.NewSpan(5, 25)

				c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathB.Span.Start)
				c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathB.Span.Start)
			})
		})

		c.Specify("the collision will end if one finishs as the other starts", func() {
			c.Specify("path A ends when path B starts", func() {
				pathA.Span = stime.NewSpan(10, 30)
				pathB.Span = stime.NewSpan(30, 40)
				c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathA.Span.End)
				c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathA.Span.End)
			})

			c.Specify("path B ends when path B starts", func() {
				pathA.Span = stime.NewSpan(30, 40)
				pathB.Span = stime.NewSpan(10, 30)
				c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathB.Span.End)
				c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathB.Span.End)
			})
		})

		c.Specify("the collision will end when the paths no longer overlap", func() {
			pathA.Span = stime.NewSpan(0, 10)
			pathB.Span = stime.NewSpan(0, 10)
			c.Expect(NewPathCollision(pathA, pathB).End(), Equals, stime.Time(5))
			c.Expect(NewPathCollision(pathB, pathA).End(), Equals, stime.Time(5))

			pathA.Span = stime.NewSpan(2, 12)
			pathB.Span = stime.NewSpan(0, 10)
			c.Expect(NewPathCollision(pathA, pathB).End(), Equals, stime.Time(6))
			c.Expect(NewPathCollision(pathB, pathA).End(), Equals, stime.Time(6))

			pathA.Span = stime.NewSpan(3, 13)
			pathB.Span = stime.NewSpan(4, 14)
			// Float answer is 8.5
			c.Expect(NewPathCollision(pathA, pathB).End(), Equals, stime.Time(9))
			c.Expect(NewPathCollision(pathB, pathA).End(), Equals, stime.Time(9))

			pathA.Span = stime.NewSpan(3, 17)
			pathB.Span = stime.NewSpan(5, 11)
			// Float answer is 8 * 3/19
			c.Expect(NewPathCollision(pathA, pathB).End(), Equals, stime.Time(9))
			c.Expect(NewPathCollision(pathB, pathA).End(), Equals, stime.Time(9))
		})

		specs := [...]struct {
			A, B        stime.Span
			description string
		}{{
			stime.NewSpan(10, 30),
			stime.NewSpan(10, 30),
			"when path A and B start and end at the same time",
		}, {
			stime.NewSpan(10, 30),
			stime.NewSpan(30, 50),
			"when path A ends as path B starts",
		}, {
			stime.NewSpan(10, 30),
			stime.NewSpan(20, 40),
			"when path A starts and ends before path B",
		}, {
			stime.NewSpan(10, 30),
			stime.NewSpan(20, 30),
			"when path A starts before and ends with path B",
		}, {
			stime.NewSpan(10, 30),
			stime.NewSpan(20, 25),
			"when path A starts before and ends after path B",
		}}

		for _, spec := range specs {
			c.Specify(spec.description+" the overlap will start at 1.0 and decrease to 0.0", func() {
				pathA.Span = spec.A
				pathB.Span = spec.B
				overlapShrinksTo0(c, NewPathCollision(pathA, pathB))
				overlapShrinksTo0(c, NewPathCollision(pathB, pathA))
			})
		}
	})

	c.Specify("when path A and path B share the same origin and destination", func() {
		o, d := Cell{0, 0}, Cell{1, 0}

		pathA = PathAction{stime.NewSpan(10, 30), o, d}
		pathB = PathAction{stime.NewSpan(10, 30), o, d}

		c.Assume(NewPathCollision(pathA, pathB).Type(), Equals, CT_SAME_ORIG_DEST)
		c.Assume(NewPathCollision(pathB, pathA).Type(), Equals, CT_SAME_ORIG_DEST)

		c.Specify("the collision begins when either path starts", func() {
			c.Specify("path A starts first", func() {
				pathA.Span = stime.NewSpan(5, 25)

				c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathA.Span.Start)
				c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathA.Span.Start)
			})

			c.Specify("path B starts first", func() {
				pathB.Span = stime.NewSpan(5, 25)

				c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathB.Span.Start)
				c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathB.Span.Start)
			})
		})

		c.Specify("the collision ends when both paths have completed", func() {
			c.Specify("path A finishes last", func() {
				pathA.Span = stime.NewSpan(15, 35)

				c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathA.Span.End)
				c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathA.Span.End)
			})

			c.Specify("path B finishes last", func() {
				pathB.Span = stime.NewSpan(15, 35)

				c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathB.Span.End)
				c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathB.Span.End)
			})
		})

		c.Specify("the overlap will decrease to a trough and the grow back to 1.0", func() {})
		c.Specify("the overlap will be 1.0 for the duration of the collision", func() {})
	})

	c.Specify("when path A and path B share the same origin and are perpendicular", func() {
		o := Cell{0, 0}
		pathA = PathAction{stime.NewSpan(10, 30), o, o.Neighbor(North)}
		pathB = PathAction{stime.NewSpan(10, 30), o, o.Neighbor(East)}

		c.Specify("a same origin collision is identified", func() {
			c.Expect(NewPathCollision(pathA, pathB).Type(), Equals, CT_SAME_ORIG_PERP)
			c.Expect(NewPathCollision(pathB, pathA).Type(), Equals, CT_SAME_ORIG_PERP)

			pathB = PathAction{stime.NewSpan(10, 30), o, o.Neighbor(West)}
			c.Expect(NewPathCollision(pathA, pathB).Type(), Equals, CT_SAME_ORIG_PERP)
			c.Expect(NewPathCollision(pathB, pathA).Type(), Equals, CT_SAME_ORIG_PERP)
		})

		c.Specify("the collision begins when either path starts first", func() {
			c.Specify("path A starts first", func() {
				pathA.Span = stime.NewSpan(5, 25)

				c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathA.Span.Start)
				c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathA.Span.Start)
			})

			c.Specify("path B starts first", func() {
				pathB.Span = stime.NewSpan(5, 25)

				c.Expect(NewPathCollision(pathA, pathB).Start(), Equals, pathB.Span.Start)
				c.Expect(NewPathCollision(pathB, pathA).Start(), Equals, pathB.Span.Start)
			})
		})

		c.Specify("the collision ends when the first one finishes", func() {
			c.Specify("path A finishes first", func() {
				pathA.Span = stime.NewSpan(5, 25)

				c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathA.Span.End)
				c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathA.Span.End)
			})

			c.Specify("path B finishes first", func() {
				pathB.Span = stime.NewSpan(5, 25)

				c.Expect(NewPathCollision(pathA, pathB).End(), Equals, pathB.Span.End)
				c.Expect(NewPathCollision(pathB, pathA).End(), Equals, pathB.Span.End)
			})
		})

		c.Specify("the overlap will start at 1.0 and decrease to 0.0", func() {
			overlapShrinksTo0(c, NewPathCollision(pathA, pathB))
			overlapShrinksTo0(c, NewPathCollision(pathB, pathA))
		})
	})
}

func DescribeCellCollision(c gospec.Context) {
	c.Specify("a cell-path collision can be calculated", func() {
		c.Specify("as not happening", func() {
			cell := Cell{0, 0}
			path := PathAction{
				Span: stime.NewSpan(10, 20),
				Orig: Cell{1, 1},
				Dest: Cell{1, 0},
			}
			collision := path.CollidesWith(cell)
			c.Expect(collision.Type(), Equals, CT_NONE)
		})

		c.Specify("if the path's origin is the cell", func() {
			cell := Cell{0, 0}
			path := PathAction{
				Span: stime.NewSpan(10, 20),
				Orig: cell,
				Dest: Cell{1, 0},
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
				Span: stime.NewSpan(10, 30),
				Orig: Cell{1, 0},
				Dest: cell,
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
