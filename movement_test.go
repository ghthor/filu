package engine

import (
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
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
		c.Expect(North.String(), Equals, "north")
		c.Expect(East.String(), Equals, "east")
		c.Expect(South.String(), Equals, "south")
		c.Expect(West.String(), Equals, "west")
	})
}

func DescribeWorldCoord(c gospec.Context) {
	worldCoord := WorldCoord{0, 0}

	c.Specify("neighbors", func() {
		c.Expect(worldCoord.Neighbor(North), Equals, WorldCoord{0, 1})
		c.Expect(worldCoord.Neighbor(East), Equals, WorldCoord{1, 0})
		c.Expect(worldCoord.Neighbor(South), Equals, WorldCoord{0, -1})
		c.Expect(worldCoord.Neighbor(West), Equals, WorldCoord{-1, 0})
	})

	c.Specify("determining directions between points", func() {
		c.Expect(worldCoord.DirectionTo(WorldCoord{0, 1}), Equals, North)
		c.Expect(worldCoord.DirectionTo(WorldCoord{1, 0}), Equals, East)
		c.Expect(worldCoord.DirectionTo(WorldCoord{0, -1}), Equals, South)
		c.Expect(worldCoord.DirectionTo(WorldCoord{-1, 0}), Equals, West)

		defer func() {
			x := recover()
			c.Expect(x, Not(IsNil))
			c.Expect(x, Equals, "unable to calculate Direction")
		}()

		worldCoord.DirectionTo(WorldCoord{1, 1})
	})
}

func DescribePathAction(c gospec.Context) {

	// TODO This test might not cover really short durations
	c.Specify("should calculate PartialWorldCoord percentages", func() {

		pa := PathAction{
			NewTimeSpan(10, 20),
			WorldCoord{0, 0},
			WorldCoord{0, 1},
		}

		c.Specify("where the sum of the origin% and destination% equals 1.0", func() {
			for t := pa.start; t <= pa.end; t++ {
				p := [...]PartialWorldCoord{
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
					orig := pa.OrigPartial(pa.start - 1)
					dest := pa.DestPartial(pa.start - 1)

					c.Expect(orig.Percentage, Equals, 1.0)
					c.Expect(dest.Percentage, Equals, 0.0)
				})

				c.Specify("is the start of the action", func() {
					orig := pa.OrigPartial(pa.start)
					dest := pa.DestPartial(pa.start)

					c.Expect(orig.Percentage, Equals, 1.0)
					c.Expect(dest.Percentage, Equals, 0.0)
				})
			})

			c.Specify("equal 0.0 and 1.0 if the time", func() {
				c.Specify("is the end of the action", func() {
					orig := pa.OrigPartial(pa.end)
					dest := pa.DestPartial(pa.end)

					c.Expect(orig.Percentage, Equals, 0.0)
					c.Expect(dest.Percentage, Equals, 1.0)
				})

				c.Specify("after the end of the action", func() {
					orig := pa.OrigPartial(pa.end + 1)
					dest := pa.DestPartial(pa.end + 1)

					c.Expect(orig.Percentage, Equals, 0.0)
					c.Expect(dest.Percentage, Equals, 1.0)
				})
			})
		})
	})

	c.Specify("must know which WorldCoord's that it traverses through", func() {
		pa := PathAction{
			Dest: WorldCoord{0, 0},
			Orig: WorldCoord{0, 1},
		}

		c.Expect(pa.Traverses(WorldCoord{0, 0}), IsTrue)
		c.Expect(pa.Traverses(WorldCoord{0, 1}), IsTrue)
	})

	c.Specify("must know if it is traversing a coordinate at an instant in time", func() {
		clk, duration := Clock(0), int64(25)

		Orig := WorldCoord{0, 0}
		Dest := WorldCoord{0, 1}

		pa := PathAction{
			NewTimeSpan(clk.Now(), clk.Future(duration)),
			Orig,
			Dest,
		}

		_, err := pa.TraversesAt(WorldCoord{-1, 0}, clk.Now())
		c.Expect(err, Not(IsNil))
		c.Expect(err.Error(), Equals, "wcOutOfRange")

		c.Specify("must traverse through the starting and ending coordinates", func() {
			clk = clk.Tick()

			pwc, err := pa.TraversesAt(Orig, clk.Now())

			c.Expect(err, IsNil)
			c.Expect(pwc.WorldCoord, Equals, Orig)

			pwc, err = pa.TraversesAt(Dest, clk.Now())

			c.Expect(err, IsNil)
			c.Expect(pwc.WorldCoord, Equals, Dest)
		})

		c.Specify("but not if the time is out of the scope of this action", func() {
			_, err := pa.TraversesAt(Dest, pa.start-1)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "timeOutOfRange")

			_, err = pa.TraversesAt(Dest, pa.end+1)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "timeOutOfRange")
		})

		c.Specify("shouldn't traverse end coord at the begining of the action", func() {
			_, err := pa.TraversesAt(Dest, pa.start)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "miss")
		})

		c.Specify("shouldn't traverse start coord at the end of the action", func() {
			_, err := pa.TraversesAt(Orig, pa.end)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "miss")
		})
	})

	pa1 := PathAction{
		Orig: WorldCoord{0, 0},
		Dest: WorldCoord{0, 1},
	}

	pa2 := PathAction{
		Orig: WorldCoord{0, 0},
		Dest: WorldCoord{0, -1},
	}

	c.Specify("must know it's orientation to other PathActions", func() {
		c.Expect(pa1.IsParallelTo(pa2), IsTrue)
		c.Expect(pa2.IsParallelTo(pa1), IsTrue)

		pa2.Orig = WorldCoord{1, -1}

		c.Expect(pa1.IsParallelTo(pa2), IsFalse)
		c.Expect(pa2.IsParallelTo(pa1), IsFalse)
	})

	c.Specify("must know if it crosses another PathAction", func() {
		c.Expect(pa1.Crosses(pa2), IsTrue)
		c.Expect(pa2.Crosses(pa1), IsTrue)

		pa2 = PathAction{
			Orig: WorldCoord{1, 1},
			Dest: WorldCoord{0, 1},
		}
		c.Expect(pa1.Crosses(pa2), IsTrue)
		c.Expect(pa2.Crosses(pa1), IsTrue)

		pa2 = PathAction{
			Orig: WorldCoord{1, 1},
			Dest: WorldCoord{1, 2},
		}
		c.Expect(pa1.Crosses(pa2), IsFalse)
		c.Expect(pa2.Crosses(pa1), IsFalse)
	})

	c.Specify("when checking for a collision", func() {
		pa1 = PathAction{
			NewTimeSpan(WorldTime(10), WorldTime(20)),
			WorldCoord{0, 0},
			WorldCoord{0, 1},
		}

		pa2 = PathAction{
			NewTimeSpan(WorldTime(15), WorldTime(25)),
			WorldCoord{0, 0},
			WorldCoord{0, 1},
		}

		c.Assume(pa1.Overlaps(pa2.TimeSpan), IsTrue)

		c.Specify("must check if there wasn't a chance for collision", func() {
			c.Specify("because times don't overlap", func() {
				pa2.TimeSpan = NewTimeSpan(pa1.end+1, pa1.end+11)
				c.Expect(pa1.Collides(pa2).Type, Equals, CT_NONE)
				c.Expect(pa2.Collides(pa1).Type, Equals, CT_NONE)
			})

			c.Specify("because they don't cross", func() {
				pa2.Orig = WorldCoord{10, 11}
				pa2.Dest = WorldCoord{10, 10}
				c.Expect(pa1.Collides(pa2).Type, Equals, CT_NONE)
				c.Expect(pa2.Collides(pa1).Type, Equals, CT_NONE)
			})
		})

		c.Specify("must know it has the same origin as the other PathAction", func() {
			pa2.Dest = WorldCoord{1, 0}
			c.Expect(pa1.Collides(pa2).Type, Equals, CT_SAME_ORIG)

			pa2.Dest = WorldCoord{-1, 0}
			c.Expect(pa1.Collides(pa2).Type, Equals, CT_SAME_ORIG)

			pa2.Dest = WorldCoord{0, -1}
			c.Expect(pa1.Collides(pa2).Type, Equals, CT_SAME_ORIG)
		})

		c.Specify("must know it has the same origin & destination as the other PathAction", func() {
			c.Expect(pa1.Collides(pa2).Type, Equals, CT_SAME_ORIG_DEST)
		})

		c.Specify("must know it has the same destination as the other PathAction", func() {
			c.Assume(pa1.Dest, Equals, pa2.Dest)

			pa2.Orig = WorldCoord{1, 1}
			c.Expect(pa1.Collides(pa2).Type, Equals, CT_FROM_SIDE)

			pa2.Orig = WorldCoord{-1, 1}
			c.Expect(pa1.Collides(pa2).Type, Equals, CT_FROM_SIDE)

			pa2.Orig = WorldCoord{0, 2}
			c.Expect(pa1.Collides(pa2).Type, Equals, CT_HEAD_TO_HEAD)
		})

		c.Specify("must set the time when the collision begins when the destination is the same", func() {
			pa1 = PathAction{
				NewTimeSpan(WorldTime(10), WorldTime(25)),
				WorldCoord{0, 0},
				WorldCoord{0, 1},
			}

			pa2 = PathAction{
				NewTimeSpan(WorldTime(9), WorldTime(19)),
				WorldCoord{1, 1},
				WorldCoord{0, 1},
			}

			c.Assume(pa1.Overlaps(pa2.TimeSpan), IsTrue)

			for pa1.start <= pa2.end && pa1.start < pa1.end {
				c.Expect(pa1.Collides(pa2).T, Equals, pa1.start)
				c.Expect(pa2.Collides(pa1).T, Equals, pa1.start)
				pa1.start += 1
			}

			pa1.start = pa2.start

			for pa2.start <= pa1.end && pa2.start < pa2.end {
				c.Expect(pa1.Collides(pa2).T, Equals, pa2.start)
				c.Expect(pa2.Collides(pa1).T, Equals, pa2.start)
				pa2.start += 1
			}
		})
	})
}

func DescribeCollision(c gospec.Context) {
	c.Specify("overlap for a collision that is", func() {

		// This is a good spec, I'm really happy with this one
		c.Specify("A following into B's Position", func() {
			A := PathAction{TimeSpan: NewTimeSpan(10, 20)}
			B := PathAction{TimeSpan: NewTimeSpan(12, 22)}

			c.Assume(A.Overlaps(B.TimeSpan), IsTrue)

			B.Orig = WorldCoord{0, 0}
			B.Dest = WorldCoord{1, 0}

			c.Specify("from the side", func() {
				A.Orig = WorldCoord{0, 1}
				A.Dest = B.Orig

				collision := B.Collides(A)
				c.Assume(collision.A, Equals, A)
				c.Assume(collision.B, Equals, B)
				c.Assume(collision.Type, Equals, CT_A_INTO_B_FROM_SIDE)
				c.Assume(collision.T, Equals, A.start)

				collision = A.Collides(B)
				c.Assume(collision.A, Equals, A)
				c.Assume(collision.B, Equals, B)
				c.Assume(collision.Type, Equals, CT_A_INTO_B_FROM_SIDE)
				c.Assume(collision.T, Equals, A.start)

				c.Specify("overlap will be 1.0 if B starts as A ends", func() {
					B.TimeSpan = NewTimeSpan(A.end, A.end+10)

					c.Assume(A.Overlaps(B.TimeSpan), IsTrue)

					collision = B.Collides(A)
					collision.T = B.start

					c.Expect(collision.Overlap(), Equals, 1.0)
				})

				c.Specify("overlap will grow to a peak and then diminish", func() {
					overlapStart := collision.Overlap()
					prevOverlap := overlapStart
					overlap := prevOverlap

					end := A.end
					if B.end > A.end {
						end = B.end
					}

					peak := 0.0

					for collision.T += 1; collision.T <= end; collision.T += 1 {
						overlap = collision.Overlap()

						if overlap <= prevOverlap {
							peak = prevOverlap
							prevOverlap = overlap
							break
						}
						prevOverlap = overlap
					}

					c.Expect(peak, Satisfies, peak > overlapStart)
					c.Expect(collision.T, Satisfies, collision.T < end)

					for collision.T += 1; collision.T <= end; collision.T += 1 {
						overlap = collision.Overlap()

						c.Expect(overlap, Satisfies, overlap < prevOverlap)
						c.Expect(overlap, Satisfies, overlap < peak)

						prevOverlap = overlap
					}

					c.Expect(overlap, Equals, 0.0)
				})
			})

			c.Specify("in the same direction", func() {
				A.Orig = WorldCoord{-1, 0}
				A.Dest = B.Orig

				collision := B.Collides(A)
				c.Assume(collision.A, Equals, A)
				c.Assume(collision.B, Equals, B)
				c.Assume(collision.Type, Equals, CT_A_INTO_B)
				c.Assume(collision.T, Equals, A.start)

				collision = A.Collides(B)
				c.Assume(collision.A, Equals, A)
				c.Assume(collision.B, Equals, B)
				c.Assume(collision.Type, Equals, CT_A_INTO_B)
				c.Assume(collision.T, Equals, A.start)

				c.Specify("will not occur if A starts at the same time as B and A finishes at the same time as B", func() {
					A.TimeSpan = NewTimeSpan(0, 100)
					B.TimeSpan = NewTimeSpan(0, 100)

					c.Assume(A.start, Satisfies, A.start == B.start)
					c.Assume(A.end, Satisfies, A.end == B.end)

					collision := A.Collides(B)

					c.Expect(collision.Type, Equals, CT_NONE)

				})

				c.Specify("will not occur if A starts after B and A finishes after B", func() {
					A.TimeSpan = NewTimeSpan(1, 101)
					B.TimeSpan = NewTimeSpan(0, 100)

					c.Assume(A.start, Satisfies, A.start > B.start)
					c.Assume(A.end, Satisfies, A.end > B.end)

					collision := A.Collides(B)
					c.Expect(collision.Type, Equals, CT_NONE)
				})

				c.Specify("overlap", func() {
					c.Specify("will be 1.0 if B starts as A ends", func() {
						B.TimeSpan = NewTimeSpan(A.end, A.end+10)

						c.Assume(A.Overlaps(B.TimeSpan), IsTrue)

						collision = B.Collides(A)
						collision.T = B.start

						c.Expect(collision.Overlap(), Equals, 1.0)
					})

					c.Specify("will grow to a peak", func() {
						peak := 0.0
						c.Specify("when A starts before B", func() {
							c.Assume(A.start, Satisfies, A.start < B.start)

							overlap := collision.Overlap()
							prevOverlap := overlap

							end := A.end
							if B.end > A.end {
								end = B.end
							}

							for collision.T += 1; collision.T <= end; collision.T += 1 {
								overlap = collision.Overlap()
								if overlap <= prevOverlap {
									peak = prevOverlap
									prevOverlap = overlap
									break
								}
								prevOverlap = overlap
							}

							c.Specify("and remain at that peak until B ends if A and B have the same speed", func() {
								c.Assume(B.duration, Equals, A.duration)
								c.Assume(collision.T, Satisfies, collision.T < A.end)

								for ; collision.T <= A.end; collision.T += 1 {
									overlap = collision.Overlap()
									c.Expect(overlap, IsWithin(0.00000000001), peak)
									prevOverlap = overlap
								}

								c.Specify("and then diminish till B ends", func() {
									c.Assume(collision.T, Satisfies, collision.T < B.end)

									for ; collision.T <= B.end; collision.T += 1 {
										overlap = collision.Overlap()
										c.Expect(overlap, Satisfies, overlap < peak)
										c.Expect(overlap, Satisfies, overlap < prevOverlap)
										prevOverlap = overlap
									}

									c.Expect(overlap, Equals, 0.0)
								})
							})

							c.Specify("then peak will diminish when B is faster then A", func() {
								B.TimeSpan = NewTimeSpan(B.start, WorldTime(int64(B.start)+A.duration-1))

								c.Assume(B.Overlaps(A.TimeSpan), IsTrue)
								c.Assume(B.duration, Satisfies, B.duration < A.duration)

								collision = A.Collides(B)
								overlap = collision.Overlap()
								prevOverlap = overlap
								peak = 0.0

								end := A.end
								if B.end > A.end {
									end = B.end
								}

								for collision.T += 1; collision.T <= end; collision.T += 1 {
									overlap = collision.Overlap()

									if overlap <= prevOverlap {
										peak = prevOverlap
										prevOverlap = overlap
										break
									}
									prevOverlap = overlap
								}

								for collision.T += 1; collision.T <= end; collision.T += 1 {
									overlap = collision.Overlap()

									c.Expect(overlap, Satisfies, overlap < peak)
									c.Expect(overlap, Satisfies, overlap < prevOverlap)

									prevOverlap = overlap
								}

								c.Expect(overlap, Equals, 0.0)
							})
						})
					})

					c.Specify("when B starts before A and A is faster then B", func() {
						A.TimeSpan = NewTimeSpan(1, 99)
						B.TimeSpan = NewTimeSpan(0, 100)

						// TODO Make Custom Matchers for these
						c.Assume(B.start, Satisfies, B.start < A.start)
						c.Assume(A.end, Satisfies, A.end < B.end)
						c.Assume(A.duration, Satisfies, A.duration < B.duration)

						collision := A.Collides(B)

						c.Assume(collision.Type, Equals, CT_A_INTO_B)

						c.Specify("the overlap will be 0.0 until A catches B", func() {

							overlap := collision.Overlap()
							prevOverlap := overlap

							for collision.T += 1; collision.T <= A.end; collision.T += 1 {
								overlap = collision.Overlap()
								if overlap > prevOverlap {
									prevOverlap = overlap
									break
								}
								prevOverlap = overlap
							}

							c.Assume(collision.T, Not(Equals), A.end)

							c.Specify("overlap will grow until A has ended", func() {
								for collision.T += 1; collision.T <= A.end; collision.T += 1 {
									overlap = collision.Overlap()
									c.Expect(overlap, Satisfies, overlap > prevOverlap)
									prevOverlap = overlap
								}

								c.Assume(collision.T, Equals, A.end+1)
								c.Assume(collision.T, Satisfies, collision.T <= B.end)

								c.Specify("then it will diminish until B has ended", func() {
									for ; collision.T <= B.end; collision.T += 1 {
										overlap = collision.Overlap()
										c.Expect(overlap, Satisfies, overlap < prevOverlap)
										prevOverlap = overlap
									}

									c.Expect(overlap, Equals, 0.0)
								})
							})
						})
					})
				})
			})
		})

		c.Specify("A contesting B for position", func() {
			c.Specify("head to head", func() {
				pa1 := PathAction{
					NewTimeSpan(10, 20),
					WorldCoord{0, 0},
					WorldCoord{1, 0},
				}

				pa2 := PathAction{
					NewTimeSpan(12, 22),
					WorldCoord{2, 0},
					WorldCoord{1, 0},
				}

				c.Assume(pa1.Overlaps(pa2.TimeSpan), IsTrue)

				collision := pa1.Collides(pa2)
				c.Assume(collision.Type, Equals, CT_HEAD_TO_HEAD)

				c.Specify("should always be >= 0.0", func() {
					overlap := collision.Overlap()
					c.Expect(overlap, Satisfies, overlap >= 0.0)
				})

				c.Specify("shouldn't be > 0.0 until sum of pertcentages is > 1.0", func() {
					// 2 moving objects meet exactly
					collision.T = 16
					overlap := collision.Overlap()

					c.Expect(overlap, Equals, 0.0)

					// Next step they begin to overlap
					collision.T += 1
					overlap = collision.Overlap()

					c.Expect(overlap, Satisfies, overlap > 0.0)
				})
			})

			c.Specify("from the side", func() {
				pa1 := PathAction{
					NewTimeSpan(10, 25),
					WorldCoord{0, 0},
					WorldCoord{1, 0},
				}

				pa2 := PathAction{
					NewTimeSpan(12, 22),
					WorldCoord{1, 1},
					WorldCoord{1, 0},
				}

				collision := pa1.Collides(pa2)

				c.Assume(collision.Type, Equals, CT_FROM_SIDE)
				c.Assume(collision.T, Equals, WorldTime(12))

				c.Specify("should equal 0.0 at the beginning of the collision", func() {
					overlap := collision.Overlap()

					//c.Expect(overlap, IsWithin(0.0001), 0.0)
					c.Expect(overlap, Equals, 0.0)
				})

				c.Specify("should grow from 0.0 to 1.0 over time", func() {
					overlap := collision.Overlap()
					prevOverlap := overlap

					// Go to the further action endpoint
					end := pa1.end
					if pa2.end > pa1.end {
						end = pa2.end
					}

					c.Assume(overlap, Equals, 0.0)

					for collision.T += 1; collision.T < end; collision.T += 1 {
						overlap = collision.Overlap()
						c.Expect(overlap, Satisfies, overlap > prevOverlap)
						prevOverlap = overlap
					}
					c.Expect(overlap, Not(Equals), 1.0)

					overlap = collision.Overlap()
					c.Expect(overlap, Equals, 1.0)
				})
			})
		})
	})
}
