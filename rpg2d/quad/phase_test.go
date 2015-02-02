package quad_test

import (
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/quad"
	"github.com/ghthor/engine/sim/stime"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

var quadBounds = coord.Bounds{
	coord.Cell{-8, 8},
	coord.Cell{7, -7},
}

const quadMaxSize = 4

// Creates a set of entities in collision groups
// used for testing the broad phase.
func cgEntitiesDataSet() ([]MockEntityWithBounds, []quad.Collision, []quad.CollisionGroup) {
	entities := func() []MockEntityWithBounds {
		c := func(x, y int) coord.Cell { return coord.Cell{x, y} }
		b := func(tl, br coord.Cell) coord.Bounds { return coord.Bounds{tl, br} }

		return []MockEntityWithBounds{
			{ // CollisionGroup 0
				0, c(0, 0),
				b(c(0, 0), c(1, 0)),
			}, {
				1, c(1, 0),
				b(c(1, 0), c(2, 0)),
			},

			{ // CollisionGroup 1
				2, c(1, 1),
				b(c(1, 2), c(1, 1)),
			}, {
				3, c(1, 3),
				b(c(1, 3), c(1, 2)),
			}, {
				4, c(2, 2),
				b(c(1, 2), c(2, 2)),
			},

			{ // CollisionGroup 2
				5, c(-1, 0),
				b(c(-2, 0), c(-2, 0)),
			}, {
				6, c(-2, 0),
				b(c(-2, 0), c(-2, -1)),
			}, {
				7, c(-2, -1),
				b(c(-2, -1), c(-1, -1)),
			}, {
				8, c(-1, -1),
				b(c(-1, -1), c(0, -1)),
			}, {
				9, c(0, -1),
				b(c(0, -1), c(1, -1)),
			}, {
				10, c(1, -1),
				b(c(1, -1), c(1, -2)),
			}, {
				11, c(1, -2),
				b(c(-2, -2), c(1, -2)),
			},

			{ // CollisionGroup 3
				12, c(0, 5),
				b(c(0, 5), c(1, 5)),
			}, {
				13, c(1, 5),
				b(c(1, 5), c(2, 5)),
			}, {
				14, c(0, 6),
				b(c(0, 6), c(1, 6)),
			}, {
				15, c(1, 6),
				b(c(1, 6), c(2, 6)),
			}, {
				16, c(2, 6),
				b(c(2, 6), c(2, 5)),
			},
		}
	}()

	collisions := func(e []MockEntityWithBounds) []quad.Collision {
		c := func(a, b entity.Entity) quad.Collision { return quad.Collision{a, b} }

		return []quad.Collision{
			// Group 0
			c(e[0], e[1]),

			// Group 1
			c(e[2], e[3]),
			c(e[2], e[4]),
			c(e[3], e[4]),

			// Group 3
			c(e[5], e[6]),
			c(e[6], e[7]),
			c(e[7], e[8]),
			c(e[8], e[9]),
			c(e[9], e[10]),
			c(e[10], e[11]),

			// Group 3,
			c(e[12], e[13]),
			c(e[13], e[16]),
			c(e[14], e[15]),
			c(e[15], e[16]),
		}
	}(entities)

	cgroups := func(c []quad.Collision) []quad.CollisionGroup {
		cg := func(collisions ...quad.Collision) (cg quad.CollisionGroup) {
			for _, c := range collisions {
				cg = cg.AddCollision(c)
			}
			return
		}

		return []quad.CollisionGroup{
			cg(c[0]),
			cg(c[1:4]...),
			cg(c[4:10]...),
			cg(c[10:14]...),
		}
	}(collisions)

	return entities, collisions, cgroups
}

func DescribePhase(c gospec.Context) {
	c.Specify("the input phase", func() {
		c.Specify("will remove any entites that move out of bounds", func() {
			q, err := quad.New(coord.Bounds{
				TopL: coord.Cell{-16, 16},
				BotR: coord.Cell{15, -15},
			}, 3, nil)
			c.Assume(err, IsNil)

			// A Single Entity
			q = q.Insert(MockEntity{0, coord.Cell{-16, 16}})
			c.Assume(len(q.QueryBounds(q.Bounds())), Equals, 1)

			q, outOfBounds := quad.RunInputPhaseOn(q, quad.InputPhaseHandlerFn(func(chunk quad.Chunk, now stime.Time) quad.Chunk {
				c.Assume(len(chunk.Entities), Equals, 1)

				// Move the entity out of bounds
				chunk.Entities[0] = MockEntity{0, coord.Cell{-17, 16}}

				return chunk
			}), stime.Time(0))

			c.Expect(len(outOfBounds), Equals, 1)
			c.Expect(len(q.QueryBounds(q.Bounds())), Equals, 0)

			// Multiple entities
			q = q.Insert(MockEntity{0, coord.Cell{-16, 16}})
			q = q.Insert(MockEntity{1, coord.Cell{15, -15}})
			q = q.Insert(MockEntity{2, coord.Cell{-1, 1}})
			q = q.Insert(MockEntity{3, coord.Cell{0, 0}})
			q = q.Insert(MockEntity{4, coord.Cell{5, -2}})

			q, outOfBounds = quad.RunInputPhaseOn(q, quad.InputPhaseHandlerFn(func(chunk quad.Chunk, now stime.Time) quad.Chunk {
				// Move the entity out of bounds
				for i, e := range chunk.Entities {
					switch e.Id() {
					case 1:
						// Move out of quadtree's bounds
						chunk.Entities[i] = MockEntity{1, coord.Cell{16, -15}}
					case 4:
						// Move from SE to NE quadrant
						chunk.Entities[i] = MockEntity{4, coord.Cell{5, 5}}
					}
				}

				return chunk
			}), stime.Time(0))

			c.Expect(len(q.QueryBounds(q.Bounds())), Equals, 4)
			c.Expect(len(outOfBounds), Equals, 1)

			c.Expect(q.QueryCell(coord.Cell{-16, 16})[0].Id(), Equals, int64(0))
			c.Expect(outOfBounds[0].Id(), Equals, int64(1))
			c.Expect(q.QueryCell(coord.Cell{-1, 1})[0].Id(), Equals, int64(2))
			c.Expect(q.QueryCell(coord.Cell{0, 0})[0].Id(), Equals, int64(3))
			c.Expect(q.QueryCell(coord.Cell{5, 5})[0].Id(), Equals, int64(4))

		})
	})

	c.Specify("the broad phase", func() {
		cgEntities, _, cgroups := cgEntitiesDataSet()
		// Sanity check the data set is what I expect it
		// to be. Have these check because this is the first
		// time I've used slice operations extensively and I
		// want to make sure I'm using the right indices in
		// range expressions.
		c.Assume(len(cgroups[0].Entities), Equals, 2)
		c.Assume(len(cgroups[0].Collisions), Equals, 1)
		c.Assume(cgroups[0].Entities, ContainsAll, cgEntities[0:2])
		c.Assume(cgroups[0].Entities, Not(ContainsAny), cgEntities[2:])

		c.Assume(len(cgroups[1].Entities), Equals, 3)
		c.Assume(len(cgroups[1].Collisions), Equals, 3)
		c.Assume(cgroups[1].Entities, Not(ContainsAny), cgEntities[0:2])
		c.Assume(cgroups[1].Entities, ContainsAll, cgEntities[2:5])
		c.Assume(cgroups[1].Entities, Not(ContainsAny), cgEntities[5:])

		c.Assume(len(cgroups[2].Entities), Equals, 7)
		c.Assume(len(cgroups[2].Collisions), Equals, 6)
		c.Assume(cgroups[2].Entities, Not(ContainsAny), cgEntities[0:5])
		c.Assume(cgroups[2].Entities, ContainsAll, cgEntities[5:12])
		c.Assume(cgroups[2].Entities, Not(ContainsAny), cgEntities[12:])

		c.Assume(len(cgroups[3].Entities), Equals, 5)
		c.Assume(len(cgroups[3].Collisions), Equals, 4)
		c.Assume(cgroups[3].Entities, Not(ContainsAny), cgEntities[0:12])
		c.Assume(cgroups[3].Entities, ContainsAll, cgEntities[12:17])
		c.Assume(cgroups[3].Entities, Not(ContainsAny), cgEntities[17:])

		makeQuad := func(entities []entity.Entity) quad.Quad {
			q, err := quad.New(quadBounds, quadMaxSize, nil)
			c.Assume(err, IsNil)

			for _, e := range entities {
				q = q.Insert(e)
			}

			return q
		}

		c.Specify("will create collision groups", func() {
			type testCase struct {
				entities []entity.Entity
				cgroups  []quad.CollisionGroup
			}

			testCases := func(cg []quad.CollisionGroup) []testCase {
				tc := func(cgroups ...quad.CollisionGroup) testCase {
					var entities []entity.Entity
					for _, cg := range cgroups {
						entities = append(entities, cg.Entities...)
					}
					return testCase{entities, cgroups}
				}

				return []testCase{
					// Make test cases that only contain collision groups
					tc(cg[0]),
					tc(cg[0:2]...),
					tc(cg[0:3]...),
					tc(cg[0:4]...),
					tc(cg[1]),
					tc(cg[1:3]...),
					tc(cg[1:4]...),
					tc(cg[2]),
					tc(cg[2:4]...),
					tc(cg[3]),
					tc(cg...),
				}
			}(cgroups)

			for _, testCase := range testCases {
				q := makeQuad(testCase.entities)

				var cgroups []*quad.CollisionGroup
				q, cgroups, _, _ = quad.RunBroadPhaseOn(q, stime.Time(0))

				c.Expect(len(cgroups), Equals, len(testCase.cgroups))
				c.Expect(cgroups, ContainsAll, testCase.cgroups)
			}
		})
	})

	c.Specify("the narrow phase", func() {
		c.Specify("will realize all future potentials", func() {
		})
	})
}
