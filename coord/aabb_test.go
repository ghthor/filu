package coord

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeAABB(c gospec.Context) {
	aabb := Bounds{
		Cell{0, 0},
		Cell{0, 0},
	}

	c.Specify("the width, height and area of an aabb", func() {
		c.Expect(aabb.Width(), Equals, 1)
		c.Expect(aabb.Height(), Equals, 1)
		c.Expect(aabb.Area(), Equals, 1)

		aabb = Bounds{
			Cell{0, 0},
			Cell{1, -1},
		}
		c.Expect(aabb.Width(), Equals, 2)
		c.Expect(aabb.Height(), Equals, 2)
		c.Expect(aabb.Area(), Equals, 4)
	})

	c.Specify("aabb contains a cell inside of itself", func() {
		c.Expect(aabb.Contains(Cell{0, 0}), IsTrue)
		containsCheck := func(aabb Bounds) {
			for i := aabb.TopL.X; i <= aabb.BotR.X; i++ {
				for j := aabb.TopL.Y; j >= aabb.BotR.Y; j-- {
					c.Expect(aabb.Contains(Cell{i, j}), IsTrue)
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
		edgeCheck := func(aabb Bounds) {
			c.Assume(aabb.IsInverted(), IsFalse)

			// Horizontal Edges
			for _, y := range [...]int{aabb.TopL.Y, aabb.BotR.Y} {
				for x := aabb.TopL.X; x <= aabb.BotR.X; x++ {
					c.Expect(aabb.HasOnEdge(Cell{x, y}), IsTrue)
				}
			}

			// Vertical Edges
			for _, x := range [...]int{aabb.TopL.X, aabb.BotR.X} {
				for y := aabb.TopL.Y - 1; y > aabb.BotR.Y; y-- {
					c.Expect(aabb.HasOnEdge(Cell{x, y}), IsTrue)
				}
			}

			outside := Bounds{
				aabb.TopL.Add(-1, 1),
				aabb.BotR.Add(1, -1),
			}

			// Horizontal Edges
			for _, y := range [...]int{outside.TopL.Y, outside.BotR.Y} {
				for x := outside.TopL.X; x <= outside.BotR.X; x++ {
					c.Expect(aabb.HasOnEdge(Cell{x, y}), IsFalse)
				}
			}

			// Vertical Edges
			for _, x := range [...]int{outside.TopL.X, outside.BotR.X} {
				for y := outside.TopL.Y - 1; y > outside.BotR.Y; y-- {
					c.Expect(aabb.HasOnEdge(Cell{x, y}), IsFalse)
				}
			}
		}

		edgeCheck(aabb)

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

	c.Specify("can calulate the intersection of 2 AABBs", func() {
		aabb := Bounds{
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

			intersectionResult, err := aabb.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)

			intersectionResult, err = other.Intersection(aabb)
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

			intersectionResult, err = aabb.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)

			intersectionResult, err = other.Intersection(aabb)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)
		})

		c.Specify("when one is contained inside the other", func() {
			// aabb Contains other
			other := Bounds{
				Cell{5, -5},
				Cell{6, -6},
			}

			intersection, err := aabb.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, other)

			intersection, err = other.Intersection(aabb)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, other)

			// other Contains aabb
			other = Bounds{
				Cell{-1, 1},
				Cell{11, -11},
			}

			intersection, err = aabb.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, aabb)

			intersection, err = other.Intersection(aabb)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, aabb)
		})

		c.Specify("and an error is returned if the rectangles do not overlap", func() {
			other := Bounds{
				Cell{11, -11},
				Cell{11, -11},
			}

			_, err := aabb.Intersection(other)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "no overlap")
		})
	})

	c.Specify("flip topleft and bottomright if they are inverted", func() {
		aabb = Bounds{
			Cell{0, 0},
			Cell{-1, 1},
		}

		c.Expect(aabb.IsInverted(), IsTrue)
		c.Expect(aabb.Invert().IsInverted(), IsFalse)
	})

	c.Specify("expand AABB by a magnitude", func() {
		c.Expect(aabb.Expand(1), Equals, Bounds{
			Cell{-1, 1},
			Cell{1, -1},
		})

		aabb = Bounds{
			Cell{5, 6},
			Cell{5, -6},
		}

		c.Expect(aabb.Expand(2), Equals, Bounds{
			Cell{3, 8},
			Cell{7, -8},
		})
	})

	c.Specify("AABB can be split into 4 quads", func() {
		aabb := Bounds{
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

		quadsResult, err := splitAABBToQuads(aabb)
		c.Assume(err, IsNil)

		for i, quad := range quadsResult {
			c.Expect(quad, Equals, quads[i])
		}

		// Width == Height == 2
		aabb = Bounds{
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

		quadsResult, err = splitAABBToQuads(aabb)
		c.Assume(err, IsNil)

		for i, quad := range quadsResult {
			c.Expect(quad, Equals, quads[i])
		}

		c.Specify("only if the height is greater than 1", func() {
			aabb = Bounds{
				Cell{1, 1},
				Cell{2, 1},
			}

			_, err := splitAABBToQuads(aabb)
			c.Expect(err, Not(IsNil))
			c.Expect(err, Equals, ErrBoundsAreTooSmall)
		})

		c.Specify("only if the width is greater than 1", func() {
			aabb = Bounds{
				Cell{1, 1},
				Cell{1, 0},
			}

			_, err := splitAABBToQuads(aabb)
			c.Expect(err, Not(IsNil))
			c.Expect(err, Equals, ErrBoundsAreTooSmall)
		})

		c.Specify("only if it isn't inverted", func() {
			aabb = Bounds{
				Cell{0, 0},
				Cell{-1, 1},
			}

			_, err := splitAABBToQuads(aabb)
			c.Expect(err, Not(IsNil))
			c.Expect(err, Equals, ErrBoundsAreInverted)
		})
	})
}
