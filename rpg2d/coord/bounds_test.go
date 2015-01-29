package coord

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeBounds(c gospec.Context) {
	b := Bounds{
		Cell{0, 0},
		Cell{0, 0},
	}

	c.Specify("the width, height and area of an bounds", func() {
		c.Expect(b.Width(), Equals, 1)
		c.Expect(b.Height(), Equals, 1)
		c.Expect(b.Area(), Equals, 1)

		b = Bounds{
			Cell{0, 0},
			Cell{1, -1},
		}
		c.Expect(b.Width(), Equals, 2)
		c.Expect(b.Height(), Equals, 2)
		c.Expect(b.Area(), Equals, 4)
	})

	c.Specify("bounds contains a cell inside of itself", func() {
		c.Expect(b.Contains(Cell{0, 0}), IsTrue)
		containsCheck := func(b Bounds) {
			for i := b.TopL.X; i <= b.BotR.X; i++ {
				for j := b.TopL.Y; j >= b.BotR.Y; j-- {
					c.Expect(b.Contains(Cell{i, j}), IsTrue)
				}
			}
		}

		containsCheck(Bounds{
			Cell{0, 0},
			Cell{1, -1},
		})
		containsCheck(Bounds{
			Cell{1, 1},
			Cell{2, -10},
		})
	})

	c.Specify("can identify cells that lay on it's edges", func() {
		edgeCheck := func(b Bounds) {
			c.Assume(b.IsInverted(), IsFalse)

			// Horizontal Edges
			for _, y := range [...]int{b.TopL.Y, b.BotR.Y} {
				for x := b.TopL.X; x <= b.BotR.X; x++ {
					c.Expect(b.HasOnEdge(Cell{x, y}), IsTrue)
				}
			}

			// Vertical Edges
			for _, x := range [...]int{b.TopL.X, b.BotR.X} {
				for y := b.TopL.Y - 1; y > b.BotR.Y; y-- {
					c.Expect(b.HasOnEdge(Cell{x, y}), IsTrue)
				}
			}

			outside := Bounds{
				b.TopL.Add(-1, 1),
				b.BotR.Add(1, -1),
			}

			// Horizontal Edges
			for _, y := range [...]int{outside.TopL.Y, outside.BotR.Y} {
				for x := outside.TopL.X; x <= outside.BotR.X; x++ {
					c.Expect(b.HasOnEdge(Cell{x, y}), IsFalse)
				}
			}

			// Vertical Edges
			for _, x := range [...]int{outside.TopL.X, outside.BotR.X} {
				for y := outside.TopL.Y - 1; y > outside.BotR.Y; y-- {
					c.Expect(b.HasOnEdge(Cell{x, y}), IsFalse)
				}
			}
		}

		edgeCheck(b)

		edgeCheck(Bounds{
			Cell{0, 0},
			Cell{1, -1},
		})

		edgeCheck(Bounds{
			Cell{1, 1},
			Cell{1, -1},
		})

		edgeCheck(Bounds{
			Cell{-10, 10},
			Cell{10, -10},
		})

		edgeCheck(Bounds{
			Cell{-10, 10},
			Cell{-10, -10},
		})

		edgeCheck(Bounds{
			Cell{-10, -10},
			Cell{10, -10},
		})
	})

	c.Specify("can calulate the intersection of 2 bounds", func() {
		b := Bounds{
			Cell{0, 0},
			Cell{10, -10},
		}

		c.Specify("when they overlap", func() {
			other := Bounds{
				Cell{5, -5},
				Cell{15, -15},
			}

			intersection := Bounds{
				Cell{5, -5},
				Cell{10, -10},
			}

			intersectionResult, err := b.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)

			intersectionResult, err = other.Intersection(b)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)

			other = Bounds{
				Cell{-5, 5},
				Cell{5, -5},
			}

			intersection = Bounds{
				Cell{0, 0},
				Cell{5, -5},
			}

			intersectionResult, err = b.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)

			intersectionResult, err = other.Intersection(b)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)
		})

		c.Specify("when one is contained inside the other", func() {
			// bounds contain other
			other := Bounds{
				Cell{5, -5},
				Cell{6, -6},
			}

			intersection, err := b.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, other)

			intersection, err = other.Intersection(b)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, other)

			// other contains bounds
			other = Bounds{
				Cell{-1, 1},
				Cell{11, -11},
			}

			intersection, err = b.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, b)

			intersection, err = other.Intersection(b)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, b)
		})

		c.Specify("and an error is returned if the rectangles do not overlap", func() {
			other := Bounds{
				Cell{11, -11},
				Cell{11, -11},
			}

			_, err := b.Intersection(other)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "no overlap")
		})
	})

	c.Specify("can be joined together", func() {
		type testCase struct {
			spec   string
			bounds []Bounds
			joined Bounds
		}

		testCases := func() []testCase {
			c := func(x, y int) Cell { return Cell{x, y} }
			b := func(tl, br Cell) Bounds { return Bounds{tl, br} }

			return []testCase{{
				"when 2 bounds are overlaping",
				[]Bounds{{
					c(5, 5), c(6, 4),
				}, {
					c(4, 6), c(5, 5),
				}},
				b(c(4, 6), c(6, 4)),
			}, {
				"when 3 bounds are overlapping",
				[]Bounds{{
					c(-3, 5), c(-2, -5),
				}, {
					c(-5, 2), c(5, -2),
				}, {
					c(1, -1), c(2, -3),
				}},
				b(c(-5, 5), c(5, -5)),
			}, {
				"when 3 bounds are not overlapping",
				[]Bounds{{
					c(-3, 5), c(-2, 0),
				}, {
					c(-5, -2), c(5, -2),
				}, {
					c(1, -4), c(2, -5),
				}},
				b(c(-5, 5), c(5, -5)),
			}}
		}()

		for _, testCase := range testCases {
			c.Specify(testCase.spec, func() {
				b := testCase.bounds
				c.Expect(b[0].JoinAll(b[1:]...), Equals, testCase.joined)
			})
		}
	})

	c.Specify("flip topleft and bottomright if they are inverted", func() {
		b = Bounds{
			Cell{0, 0},
			Cell{-1, 1},
		}

		c.Expect(b.IsInverted(), IsTrue)
		c.Expect(b.Invert().IsInverted(), IsFalse)
	})

	c.Specify("expand bounds by a magnitude", func() {
		c.Expect(b.Expand(1), Equals, Bounds{
			Cell{-1, 1},
			Cell{1, -1},
		})

		b = Bounds{
			Cell{5, 6},
			Cell{5, -6},
		}

		c.Expect(b.Expand(2), Equals, Bounds{
			Cell{3, 8},
			Cell{7, -8},
		})
	})

	c.Specify("bounds can be split into 4 quads", func() {
		b := Bounds{
			Cell{0, 0},
			Cell{10, -9},
		}

		quads := [4]Bounds{{
			Cell{0, 0},
			Cell{4, -4},
		}, {
			Cell{5, 0},
			Cell{10, -4},
		}, {
			Cell{5, -5},
			Cell{10, -9},
		}, {
			Cell{0, -5},
			Cell{4, -9},
		}}

		quadsResult, err := b.Quads()
		c.Assume(err, IsNil)

		for i, quad := range quadsResult {
			c.Expect(quad, Equals, quads[i])
		}

		// Width == Height == 2
		b = Bounds{
			Cell{2, -2},
			Cell{3, -3},
		}

		quads = [4]Bounds{{
			Cell{2, -2},
			Cell{2, -2},
		}, {
			Cell{3, -2},
			Cell{3, -2},
		}, {
			Cell{3, -3},
			Cell{3, -3},
		}, {
			Cell{2, -3},
			Cell{2, -3},
		}}

		quadsResult, err = b.Quads()
		c.Assume(err, IsNil)

		for i, quad := range quadsResult {
			c.Expect(quad, Equals, quads[i])
		}

		c.Specify("if the bounds contains at least 4 cells", func() {
			type testCase struct {
				spec   string
				bounds Bounds
			}

			cases := []testCase{{
				"in the NW quadrant",
				Bounds{
					Cell{-2, 2},
					Cell{-1, 1},
				},
			}, {
				"in the NE quadrant",
				Bounds{
					Cell{1, 2},
					Cell{2, 1},
				},
			}, {
				"in the SE quadrant",
				Bounds{
					Cell{1, -1},
					Cell{2, -2},
				},
			}, {
				"in the SE quadrant",
				Bounds{
					Cell{-2, -1},
					Cell{-1, -2},
				},
			}, {
				"spanning the NW and NE quadrants",
				Bounds{
					Cell{-1, 2},
					Cell{0, 1},
				},
			}, {
				"spanning the NE and NW quadrants",
				Bounds{
					Cell{0, 2},
					Cell{1, 1},
				},
			}, {
				"spanning the NE and SE quadrants",
				Bounds{
					Cell{1, 1},
					Cell{2, 0},
				},
			}, {
				"spanning the SE and NE quadrants",
				Bounds{
					Cell{1, 0},
					Cell{2, -1},
				},
			}, {
				"spanning the SE and SW quadrants",
				Bounds{
					Cell{0, -1},
					Cell{1, -2},
				},
			}, {
				"spanning the SW and SE quadrants",
				Bounds{
					Cell{-1, -1},
					Cell{0, -2},
				},
			}, {
				"spanning the SW and NW quadrants",
				Bounds{
					Cell{-2, 0},
					Cell{-1, -1},
				},
			}, {
				"spanning the NW and SW quadrants",
				Bounds{
					Cell{-2, 1},
					Cell{-1, 0},
				},
			}}

			// Verify that all the test cases split single cell quads
			for _, testCase := range cases {
				c.Specify(testCase.spec, func() {
					b := testCase.bounds
					c.Assume(b.Area(), Equals, 4)

					quads, err := b.Quads()
					c.Assume(err, IsNil)

					for quadrant := NW; quadrant <= SW; quadrant++ {
						q := quads[quadrant]
						c.Expect(q.Area(), Equals, 1)
						c.Expect(q.TopL, Equals, q.BotR)
						c.Expect(q.Width(), Equals, 1)
						c.Expect(q.Height(), Equals, 1)

						// Verify the bounds of this quad is 1 cell
						c.Expect(q.TopL, Equals, q.TopR())
						c.Expect(q.TopL, Equals, q.BotR)
						c.Expect(q.TopL, Equals, q.BotL())

						c.Expect(q.TopR(), Equals, q.BotR)
						c.Expect(q.TopR(), Equals, q.BotL())

						c.Expect(q.BotR, Equals, q.BotR)

						// We only compare 1 of the corner values of q.
						// This is because the expectations above verify
						// That TopR == TopL == BotR == BotL
						switch quadrant {
						case NW:
							c.Expect(q.TopL, Equals, b.TopL)
						case NE:
							c.Expect(q.TopR(), Equals, b.TopR())
						case SE:
							c.Expect(q.BotR, Equals, b.BotR)
						case SW:
							c.Expect(q.BotL(), Equals, b.BotL())
						}
					}
				})
			}

			// Verify that all the test cases cannot be split if they
			// have an area of 4, but only a height of 1

			// Take a 2x2 quad
			//     X X
			//     X X
			// and transform it into
			// X X X X
			for _, testCase := range cases {
				b := testCase.bounds
				b.TopL.X -= 2
				b.BotR = Cell{b.BotR.X, b.TopL.Y}
				c.Assume(b.Height(), Equals, 1)

				_, err := b.Quads()
				c.Assume(err, Not(IsNil))

				c.Expect(err, Equals, ErrBoundsAreTooSmall)
			}

			// Take a 2x2 quad
			//     X X
			//     X X
			// and transform it into
			//     X X X X
			for _, testCase := range cases {
				b := testCase.bounds
				b.BotR = Cell{b.BotR.X + 2, b.TopL.Y}
				c.Assume(b.Height(), Equals, 1)

				_, err := b.Quads()
				c.Assume(err, Not(IsNil))

				c.Expect(err, Equals, ErrBoundsAreTooSmall)
			}

			// Take a 2x2 quad
			//     X X
			//     X X
			// and transform it into
			//   X X X X
			for _, testCase := range cases {
				b := testCase.bounds
				b.TopL.X -= 1
				b.BotR = Cell{b.BotR.X + 1, b.TopL.Y}

				_, err := b.Quads()
				c.Assume(err, Not(IsNil))

				c.Expect(err, Equals, ErrBoundsAreTooSmall)
			}

			// TODO
			// There are more transform cases,
			// But for now that should be fine

			// Verify that all the test cases cannot be split if they
			// have an area of 4, but only a width of 1

			// Take a 2x2 quad
			//
			//
			//     X X
			//     X X
			//
			// and transform it into
			//     X
			//     X
			//     X
			//     X
			for _, testCase := range cases {
				b := testCase.bounds
				b.TopL.Y += 2
				b.BotR.X = b.TopL.X
				c.Assume(b.Width(), Equals, 1)

				_, err := b.Quads()
				c.Assume(err, Not(IsNil))

				c.Expect(err, Equals, ErrBoundsAreTooSmall)
			}
		})

		c.Specify("only if the height is greater than 1", func() {
			b = Bounds{
				Cell{1, 1},
				Cell{2, 1},
			}

			_, err := b.Quads()
			c.Expect(err, Not(IsNil))
			c.Expect(err, Equals, ErrBoundsAreTooSmall)
		})

		c.Specify("only if the width is greater than 1", func() {
			b = Bounds{
				Cell{1, 1},
				Cell{1, 0},
			}

			_, err := b.Quads()
			c.Expect(err, Not(IsNil))
			c.Expect(err, Equals, ErrBoundsAreTooSmall)
		})

		c.Specify("only if it isn't inverted", func() {
			b = Bounds{
				Cell{0, 0},
				Cell{-1, 1},
			}

			_, err := b.Quads()
			c.Expect(err, Not(IsNil))
			c.Expect(err, Equals, ErrBoundsAreInverted)
		})
	})
}
