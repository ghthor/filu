package coord

import (
	"github.com/ghthor/filu/sim/stime"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeDirection(c gospec.Context) {
	c.Specify("North is parallel to South", func() {
		c.Expect(North.IsParallelTo(North), IsTrue)
		c.Expect(North.IsParallelTo(South), IsTrue)
		c.Expect(North.IsParallelTo(East), IsFalse)
		c.Expect(North.IsParallelTo(West), IsFalse)

		c.Expect(South.IsParallelTo(North), IsTrue)
		c.Expect(South.IsParallelTo(South), IsTrue)
		c.Expect(South.IsParallelTo(East), IsFalse)
		c.Expect(South.IsParallelTo(West), IsFalse)
	})

	c.Specify("East is parallel to West", func() {
		c.Expect(East.IsParallelTo(North), IsFalse)
		c.Expect(East.IsParallelTo(South), IsFalse)
		c.Expect(East.IsParallelTo(East), IsTrue)
		c.Expect(East.IsParallelTo(West), IsTrue)

		c.Expect(West.IsParallelTo(North), IsFalse)
		c.Expect(West.IsParallelTo(South), IsFalse)
		c.Expect(West.IsParallelTo(East), IsTrue)
		c.Expect(West.IsParallelTo(West), IsTrue)
	})

	c.Specify("converts to a string", func() {
		c.Expect(North.String(), Equals, "North")
		c.Expect(East.String(), Equals, "East")
		c.Expect(South.String(), Equals, "South")
		c.Expect(West.String(), Equals, "West")
	})

	c.Specify("string converts to Direction", func() {
		d, err := NewDirectionWithString("North")
		c.Expect(err, IsNil)
		c.Expect(d, Equals, North)

		d, err = NewDirectionWithString("East")
		c.Expect(err, IsNil)
		c.Expect(d, Equals, East)

		d, err = NewDirectionWithString("South")
		c.Expect(err, IsNil)
		c.Expect(d, Equals, South)

		d, err = NewDirectionWithString("West")
		c.Expect(err, IsNil)
		c.Expect(d, Equals, West)

		_, err = NewDirectionWithString("notadirection")
		c.Expect(err, Equals, ErrInvalidDirection)
	})
}

func DescribeCell(c gospec.Context) {
	cell := Cell{0, 0}

	c.Specify("neighbors", func() {
		c.Expect(cell.Neighbor(North), Equals, Cell{0, 1})
		c.Expect(cell.Neighbor(East), Equals, Cell{1, 0})
		c.Expect(cell.Neighbor(South), Equals, Cell{0, -1})
		c.Expect(cell.Neighbor(West), Equals, Cell{-1, 0})
	})

	c.Specify("determining directions between points", func() {
		c.Expect(cell.DirectionTo(Cell{0, 1}), Equals, North)
		c.Expect(cell.DirectionTo(Cell{1, 0}), Equals, East)
		c.Expect(cell.DirectionTo(Cell{0, -1}), Equals, South)
		c.Expect(cell.DirectionTo(Cell{-1, 0}), Equals, West)

		defer func() {
			x := recover()
			c.Expect(x, Not(IsNil))
			c.Expect(x, Equals, "unable to calculate Direction")
		}()

		cell.DirectionTo(Cell{1, 1})
	})
}

func DescribePathAction(c gospec.Context) {

	// TODO This test might not cover really short durations
	c.Specify("should calculate partial cell percentages", func() {

		pa := PathAction{
			stime.NewSpan(10, 20),
			Cell{0, 0},
			Cell{0, 1},
		}

		c.Specify("where the sum of the origin% and destination% equals 1.0", func() {
			for t := pa.Span.Start; t <= pa.Span.End; t++ {
				p := [...]PartialCell{
					pa.OrigPartial(t),
					pa.DestPartial(t),
				}

				sum := p[0].Percentage + p[1].Percentage
				/// I think the ϵ can be smaller, but it doesn't really matter at this point
				c.Expect(sum, IsWithin(0.000000000000000000000000000000000000000000000000000000000001), 1.0)
			}
		})

		c.Specify("where the origin% and destination%", func() {
			c.Specify("equal 1.0 and 0.0 if the time", func() {
				c.Specify("before the start of the action", func() {
					orig := pa.OrigPartial(pa.Span.Start - 1)
					dest := pa.DestPartial(pa.Span.Start - 1)

					c.Expect(orig.Percentage, Equals, 1.0)
					c.Expect(dest.Percentage, Equals, 0.0)
				})

				c.Specify("is the start of the action", func() {
					orig := pa.OrigPartial(pa.Span.Start)
					dest := pa.DestPartial(pa.Span.Start)

					c.Expect(orig.Percentage, Equals, 1.0)
					c.Expect(dest.Percentage, Equals, 0.0)
				})
			})

			c.Specify("equal 0.0 and 1.0 if the time", func() {
				c.Specify("is the end of the action", func() {
					orig := pa.OrigPartial(pa.Span.End)
					dest := pa.DestPartial(pa.Span.End)

					c.Expect(orig.Percentage, Equals, 0.0)
					c.Expect(dest.Percentage, Equals, 1.0)
				})

				c.Specify("after the end of the action", func() {
					orig := pa.OrigPartial(pa.Span.End + 1)
					dest := pa.DestPartial(pa.Span.End + 1)

					c.Expect(orig.Percentage, Equals, 0.0)
					c.Expect(dest.Percentage, Equals, 1.0)
				})
			})
		})
	})

	c.Specify("must know which cell's that it traverses through", func() {
		pa := PathAction{
			Dest: Cell{0, 0},
			Orig: Cell{0, 1},
		}

		c.Expect(pa.Traverses(Cell{0, 0}), IsTrue)
		c.Expect(pa.Traverses(Cell{0, 1}), IsTrue)
		c.Expect(pa.Traverses(Cell{0, 2}), IsFalse)
	})

	c.Specify("must know if it is traversing a cell at an instant in time", func() {
		clk, duration := stime.Clock(0), int64(25)

		Orig := Cell{0, 0}
		Dest := Cell{0, 1}

		pa := PathAction{
			stime.NewSpan(clk.Now(), clk.Future(duration)),
			Orig,
			Dest,
		}

		_, err := pa.TraversesAt(Cell{-1, 0}, clk.Now())
		c.Expect(err, Not(IsNil))
		c.Expect(err.Error(), Equals, "wcOutOfRange")

		c.Specify("must traverse through the origin and destination cells", func() {
			clk = clk.Tick()

			pc, err := pa.TraversesAt(Orig, clk.Now())

			c.Expect(err, IsNil)
			c.Expect(pc.Cell, Equals, Orig)

			pc, err = pa.TraversesAt(Dest, clk.Now())

			c.Expect(err, IsNil)
			c.Expect(pc.Cell, Equals, Dest)
		})

		c.Specify("but not if the time is out of the scope of this action", func() {
			_, err := pa.TraversesAt(Dest, pa.Span.Start-1)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "timeOutOfRange")

			_, err = pa.TraversesAt(Dest, pa.Span.End+1)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "timeOutOfRange")
		})

		c.Specify("shouldn't traverse destination cell at the begining of the action", func() {
			_, err := pa.TraversesAt(Dest, pa.Span.Start)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "miss")
		})

		c.Specify("shouldn't traverse origin cell at the end of the action", func() {
			_, err := pa.TraversesAt(Orig, pa.Span.End)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "miss")
		})
	})

	pa1 := PathAction{
		Orig: Cell{0, 0},
		Dest: Cell{0, 1},
	}

	pa2 := PathAction{
		Orig: Cell{0, 0},
		Dest: Cell{0, -1},
	}

	c.Specify("must know it's orientation to other PathActions", func() {
		c.Expect(pa1.IsParallelTo(pa2), IsTrue)
		c.Expect(pa2.IsParallelTo(pa1), IsTrue)

		pa2.Orig = Cell{1, -1}

		c.Expect(pa1.IsParallelTo(pa2), IsFalse)
		c.Expect(pa2.IsParallelTo(pa1), IsFalse)
	})

	c.Specify("must know if it crosses another PathAction", func() {
		c.Expect(pa1.Crosses(pa2), IsTrue)
		c.Expect(pa2.Crosses(pa1), IsTrue)

		pa2 = PathAction{
			Orig: Cell{1, 1},
			Dest: Cell{0, 1},
		}
		c.Expect(pa1.Crosses(pa2), IsTrue)
		c.Expect(pa2.Crosses(pa1), IsTrue)

		pa2 = PathAction{
			Orig: Cell{1, 1},
			Dest: Cell{1, 2},
		}
		c.Expect(pa1.Crosses(pa2), IsFalse)
		c.Expect(pa2.Crosses(pa1), IsFalse)
	})

	c.Specify("must know it's bounds", func() {
		actions := func() (pas [4]PathAction) {
			c := func(x, y int) Cell { return Cell{x, y} }
			pa := func(orig, dest Cell) PathAction {
				return PathAction{Orig: orig, Dest: dest}
			}

			pas[N] = pa(c(0, 0), c(0, 1))
			pas[E] = pa(c(0, 0), c(1, 0))
			pas[S] = pa(c(0, 0), c(0, -1))
			pas[W] = pa(c(0, 0), c(-1, 0))

			return
		}()

		for dir, a := range actions {
			c.Assume(a.Direction(), Equals, Direction(dir))
		}

		bounds := func() (bounds [4]Bounds) {
			c := func(x, y int) Cell { return Cell{x, y} }
			b := func(tl, br Cell) Bounds { return Bounds{tl, br} }

			bounds[N] = b(c(0, 1), c(0, 0))
			bounds[E] = b(c(0, 0), c(1, 0))
			bounds[S] = b(c(0, 0), c(0, -1))
			bounds[W] = b(c(-1, 0), c(0, 0))

			return
		}()

		for _, b := range bounds {
			c.Assume(b.IsInverted(), IsFalse)
		}

		for dir, a := range actions {
			c.Expect(a.Bounds(), Equals, bounds[dir])
		}
	})
}

func DescribeMoveAction(c gospec.Context) {
	c.Specify("an entity can move in any direction immediately after moving", func() {
		pathAction1 := &PathAction{
			stime.NewSpan(stime.Time(0), stime.Time(20)),
			Cell{0, 0},
			Cell{0, 1},
		}

		pathAction2 := &PathAction{
			stime.NewSpan(stime.Time(20), stime.Time(40)),
			Cell{0, 1},
			Cell{1, 1},
		}

		c.Expect(pathAction2.CanHappenAfter(pathAction1), IsTrue)

		pathAction2.Span.Start = pathAction1.Span.End + 1
		c.Expect(pathAction2.CanHappenAfter(pathAction1), IsFalse)
	})

	c.Specify("an entity can't move before turning", func() {
		pathAction := &PathAction{
			stime.NewSpan(stime.Time(21), stime.Time(41)),
			Cell{0, 1},
			Cell{1, 1},
		}

		turnAction := TurnAction{
			To:   pathAction.Direction().Reverse(),
			Time: stime.Time(pathAction.Start()),
		}

		c.Expect(pathAction.CanHappenAfter(turnAction), IsFalse)
	})

	c.Specify("An entity can't move immediatly after turning", func() {
		pathAction := &PathAction{
			stime.NewSpan(stime.Time(21), stime.Time(41)),
			Cell{0, 1},
			Cell{1, 1},
		}

		turnAction := TurnAction{
			To:   pathAction.Direction(),
			Time: stime.Time(pathAction.Start() - TurnActionDelay),
		}

		c.Expect(pathAction.CanHappenAfter(turnAction), IsFalse)

		turnAction.Time = turnAction.Time - 1
		c.Expect(pathAction.CanHappenAfter(turnAction), IsTrue)
	})

	c.Specify("an entity can't immediatly turn after turning", func() {
		turnAction1 := TurnAction{
			South, North,
			stime.Time(0),
		}

		turnAction2 := TurnAction{
			North, South,
			stime.Time(TurnActionDelay),
		}

		c.Expect(turnAction2.CanHappenAfter(turnAction1), IsFalse)

		turnAction2.Time = stime.Time(TurnActionDelay + 1)
		c.Expect(turnAction2.CanHappenAfter(turnAction1), IsTrue)
	})
}
