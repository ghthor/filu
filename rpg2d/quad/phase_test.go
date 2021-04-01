package quad_test

import (
	"fmt"
	"sort"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/entity/entitytest"
	"github.com/ghthor/filu/rpg2d/quad"
	"github.com/ghthor/filu/sim/stime"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

var quadBounds = coord.Bounds{
	coord.Cell{-8, 8},
	coord.Cell{7, -7},
}

// Creates a set of entities in collision groups
// used for testing the broad phase.
func cgEntitiesDataSet() ([]entitytest.MockEntityWithBounds, []quad.Collision, []*quad.CollisionGroup) {
	entities := func() []entitytest.MockEntityWithBounds {
		c := func(x, y int) coord.Cell { return coord.Cell{x, y} }
		b := func(tl, br coord.Cell) coord.Bounds { return coord.Bounds{tl, br} }
		e := func(id entity.Id, cell coord.Cell, bounds coord.Bounds) entitytest.MockEntityWithBounds {
			return entitytest.MockEntityWithBounds{
				id, cell, bounds, 0,
			}
		}

		return []entitytest.MockEntityWithBounds{
			// CollisionGroup 0
			e(0, c(0, 0), b(c(0, 0), c(1, 0))),
			e(1, c(1, 0), b(c(1, 0), c(2, 0))),

			// CollisionGroup 1
			e(2, c(1, 1), b(c(1, 2), c(1, 1))),
			e(3, c(1, 3), b(c(1, 3), c(1, 2))),
			e(4, c(2, 2), b(c(1, 2), c(2, 2))),

			// CollisionGroup 2
			e(5, c(-1, 0), b(c(-2, 0), c(-2, 0))),
			e(6, c(-2, 0), b(c(-2, 0), c(-2, -1))),
			e(7, c(-2, -1), b(c(-2, -1), c(-1, -1))),
			e(8, c(-1, -1), b(c(-1, -1), c(0, -1))),
			e(9, c(0, -1), b(c(0, -1), c(1, -1))),
			e(10, c(1, -1), b(c(1, -1), c(1, -2))),
			e(11, c(1, -2), b(c(-2, -2), c(1, -2))),

			// CollisionGroup 3
			e(12, c(0, 5), b(c(0, 5), c(1, 5))),
			e(13, c(1, 5), b(c(1, 5), c(2, 5))),
			e(14, c(2, 5), b(c(2, 5), c(3, 5))),
			e(15, c(0, 6), b(c(0, 6), c(1, 6))),
			e(16, c(1, 6), b(c(1, 6), c(2, 6))),
			e(17, c(2, 6), b(c(2, 6), c(3, 6))),
			e(18, c(3, 6), b(c(3, 6), c(3, 5))),

			// CollisionGroup 4
			e(19, c(4, 1), b(c(4, 1), c(5, 1))),
			e(20, c(4, 2), b(c(4, 2), c(5, 2))),
			e(21, c(5, 1), b(c(5, 2), c(5, 1))),

			// CollisionGroup 5
			e(22, c(0, -3), b(c(-1, -3), c(0, -3))),
			e(23, c(-1, -3), b(c(-1, -3), c(0, -3))),

			// Non Collision Group Entities
			e(24, quadBounds.TopL, b(c(quadBounds.TopL.X-1, quadBounds.TopL.Y), quadBounds.TopL)),
			e(25, c(-5, -6), b(c(-5, -6), c(-5, -6))),
			e(26, c(-5, -7), b(c(-5, -7), c(5, -7))),
		}
	}()

	collisions := func(e []entitytest.MockEntityWithBounds) []quad.Collision {
		c := func(a, b entity.Entity) quad.Collision { return quad.NewCollision(a, b) }

		return []quad.Collision{
			// Group 0
			c(e[0], e[1]),

			// Group 1
			c(e[2], e[3]),
			c(e[2], e[4]),
			c(e[3], e[4]),

			// Group 2
			c(e[5], e[6]),
			c(e[6], e[7]),
			c(e[7], e[8]),
			c(e[8], e[9]),
			c(e[9], e[10]),
			c(e[10], e[11]),

			// Group 3,
			c(e[12], e[13]),
			c(e[13], e[14]),
			c(e[14], e[18]),
			c(e[15], e[16]),
			c(e[16], e[17]),
			c(e[17], e[18]),

			// Group 4
			c(e[19], e[21]),
			c(e[20], e[21]),

			// Group 5
			c(e[22], e[23]),
		}
	}(entities)

	cgroups := func(c []quad.Collision) []*quad.CollisionGroup {
		cg := func(collisions ...quad.Collision) (cg *quad.CollisionGroup) {
			cg = quad.NewCollisionGroup(len(collisions))
			for _, c := range collisions {
				cg.AddCollisionFromMerge(c)
			}
			return
		}

		return []*quad.CollisionGroup{
			cg(c[0]),
			cg(c[1:4]...),
			cg(c[4:10]...),
			cg(c[10:16]...),
			cg(c[16:18]...),
			cg(c[18:19]...),
		}
	}(collisions)

	return entities, collisions, cgroups
}

type byId []entity.Entity

func (e byId) Len() int      { return len(e) }
func (e byId) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e byId) Less(i, j int) bool {
	return e[i].Id() < e[j].Id()
}

func DescribePhase(c gospec.Context) {
	e := func(id entity.Id, x, y int) entitytest.MockEntity {
		return entitytest.MockEntity{
			id,
			coord.Cell{x, y},
			0,
		}
	}
	cell := func(x, y int) coord.Cell { return coord.Cell{x, y} }

	c.Specify("the update phase", func() {
		q, err := quad.New(coord.Bounds{
			TopL: coord.Cell{-16, 16},
			BotR: coord.Cell{15, -15},
		}, 3, nil)
		c.Assume(err, IsNil)

		q = q.Insert(e(0, 0, 0))
		q = q.Insert(e(1, 1, 0))
		q = q.Insert(e(2, 2, 0))
		q = q.Insert(e(3, 3, 0))

		c.Specify("will insert the updated entity", func() {
			q = q.RunUpdatePhase(quad.UpdatePhaseHandlerFn(
				func(entity entity.Entity, now stime.Time) entity.Entity {
					c := entity.Cell()
					return e(entity.Id(), c.X, c.Y+1)
				},
			), stime.Time(0))

			c.Expect(len(q.QueryBounds(q.Bounds(), nil)), Equals, 4)
			c.Expect(q.QueryCell(cell(0, 1), nil)[0].Id(), Equals, entity.Id(0))
			c.Expect(q.QueryCell(cell(1, 1), nil)[0].Id(), Equals, entity.Id(1))
			c.Expect(q.QueryCell(cell(2, 1), nil)[0].Id(), Equals, entity.Id(2))
			c.Expect(q.QueryCell(cell(3, 1), nil)[0].Id(), Equals, entity.Id(3))
		})

		c.Specify("will remove an entity", func() {
			c.Assume(len(q.QueryCell(cell(2, 0), nil)), Equals, 1)

			q = q.RunUpdatePhase(quad.UpdatePhaseHandlerFn(
				func(e entity.Entity, now stime.Time) entity.Entity {
					if e.Id() == entity.Id(2) {
						return nil
					}
					return e
				},
			), stime.Time(0))

			c.Expect(len(q.QueryBounds(q.Bounds(), nil)), Equals, 3)
			c.Expect(len(q.QueryCell(cell(2, 0), nil)), Equals, 0)
		})
	})

	c.Specify("the input phase", func() {
		c.Specify("will insert new entities", func() {
			q, err := quad.New(coord.Bounds{
				TopL: coord.Cell{-16, 16},
				BotR: coord.Cell{15, -15},
			}, 3, nil)
			c.Assume(err, IsNil)

			// A Single Entity
			q = q.Insert(entitytest.MockEntity{0, coord.Cell{-16, 16}, 0})
			c.Assume(len(q.QueryBounds(q.Bounds(), nil)), Equals, 1)

			q = q.RunInputPhase(quad.InputPhaseHandlerFn(
				func(e entity.Entity, now stime.Time, changes quad.InputPhaseChanges) entity.Entity {
					changes.New(entitytest.MockEntity{1, coord.Cell{-15, 16}, 0})
					return nil
				},
			), stime.Time(0))

			c.Assume(len(q.QueryBounds(q.Bounds(), nil)), Equals, 2)
		})
	})

	c.Specify("the broad phase", func() {
		cgEntities, _, cgroups := cgEntitiesDataSet()
		// Sanity check the data set is what I expect it
		// to be. Have these check because this is the first
		// time I've used slice operations extensively and I
		// want to make sure I'm using the right indices in
		// range expressions.
		c.Assume(len(cgroups[0].Entities()), Equals, 2)
		c.Assume(len(cgroups[0].CollisionsById), Equals, 1)
		c.Assume(cgroups[0].Entities(), ContainsAll, cgEntities[0:2])
		c.Assume(cgroups[0].Entities(), Not(ContainsAny), cgEntities[2:])

		c.Assume(len(cgroups[1].Entities()), Equals, 3)
		c.Assume(len(cgroups[1].CollisionsById), Equals, 3)
		c.Assume(cgroups[1].Entities(), Not(ContainsAny), cgEntities[0:2])
		c.Assume(cgroups[1].Entities(), ContainsAll, cgEntities[2:5])
		c.Assume(cgroups[1].Entities(), Not(ContainsAny), cgEntities[5:])

		c.Assume(len(cgroups[2].Entities()), Equals, 7)
		c.Assume(len(cgroups[2].CollisionsById), Equals, 6)
		c.Assume(cgroups[2].Entities(), Not(ContainsAny), cgEntities[0:5])
		c.Assume(cgroups[2].Entities(), ContainsAll, cgEntities[5:12])
		c.Assume(cgroups[2].Entities(), Not(ContainsAny), cgEntities[12:])

		c.Assume(len(cgroups[3].Entities()), Equals, 7)
		c.Assume(len(cgroups[3].CollisionsById), Equals, 6)
		c.Assume(cgroups[3].Entities(), Not(ContainsAny), cgEntities[0:12])
		c.Assume(cgroups[3].Entities(), ContainsAll, cgEntities[12:19])
		c.Assume(cgroups[3].Entities(), Not(ContainsAny), cgEntities[19:])

		c.Assume(len(cgroups[4].Entities()), Equals, 3)
		c.Assume(len(cgroups[4].CollisionsById), Equals, 2)
		c.Assume(cgroups[4].Entities(), Not(ContainsAny), cgEntities[0:19])
		c.Assume(cgroups[4].Entities(), ContainsAll, cgEntities[19:22])
		c.Assume(cgroups[4].Entities(), Not(ContainsAny), cgEntities[22:])

		c.Assume(len(cgroups[5].Entities()), Equals, 2)
		c.Assume(len(cgroups[5].CollisionsById), Equals, 1)
		c.Assume(cgroups[5].Entities(), Not(ContainsAny), cgEntities[0:22])
		c.Assume(cgroups[5].Entities(), ContainsAll, cgEntities[22:24])
		c.Assume(cgroups[5].Entities(), Not(ContainsAny), cgEntities[24:])

		makeQuad := func(entities []entity.Entity, quadMaxSize int) quad.QuadRoot {
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
				cgroups  []*quad.CollisionGroup
				unsolved quad.CollisionGroupIndex
			}

			testCases := func(cg []*quad.CollisionGroup) []testCase {
				tc := func(cgroups ...*quad.CollisionGroup) testCase {
					var entities []entity.Entity
					for _, cg := range cgroups {
						entities = append(entities, cg.Entities()...)
					}
					sort.Sort(byId(entities))
					return testCase{entities, cgroups, nil}
				}

				return []testCase{
					// Make test cases that only contain collision groups
					tc(cg[0]),
					tc(cg[0:2]...),
					tc(cg[0:3]...),
					tc(cg[0:4]...),
					tc(cg[0:5]...),
					tc(cg[0:6]...),
					tc(cg[1]),
					tc(cg[1:3]...),
					tc(cg[1:4]...),
					tc(cg[1:5]...),
					tc(cg[1:6]...),
					tc(cg[2]),
					tc(cg[2:4]...),
					tc(cg[2:5]...),
					tc(cg[2:6]...),
					tc(cg[3]),
					tc(cg[3:5]...),
					tc(cg[3:6]...),
					tc(cg[4]),
					tc(cg[4:6]...),
					tc(cg[5]),
					tc(cg...),
				}
			}(cgroups)

			for _, testCase := range testCases {
				// I begins as 4 because that is the min
				// quad max size that is specified to work
				for i := 4; i < len(testCase.entities)+1; i++ {
					q := makeQuad(testCase.entities, i)

					var cgroups []*quad.CollisionGroup
					var unsolved quad.CollisionGroupIndex
					cgroups = q.RunBroadPhase(stime.Time(0))

					c.Expect(len(cgroups), Equals, len(testCase.cgroups))
					c.Expect(cgroups, ContainsAll, testCase.cgroups)

					// Lets break early so the output is more useful
					// in debugging why the test is failing.
					if matches, _, _, _ := ContainsAll(cgroups, testCase.cgroups); !matches &&
						!unsolved.Equals(testCase.unsolved) {
						fmt.Println("maxSize: ", i)
						return
					}
				}
			}
		})

		c.Specify("will not create a collision group", func() {
			q := makeQuad(func() []entity.Entity {
				var entities []entity.Entity
				for _, e := range cgEntities {
					entities = append(entities, e)
				}
				return entities
			}(), 10)

			cgroups := q.RunBroadPhase(stime.Time(0))

			cgroupedEntities := func() []entity.Entity {
				var entities []entity.Entity
				for _, cg := range cgroups {
					entities = append(entities, cg.Entities()...)
				}
				return entities
			}()

			c.Expect(cgroupedEntities, Not(Contains), cgEntities[25])
			c.Expect(cgroupedEntities, Not(Contains), cgEntities[26])
		})
	})

	c.Specify("the narrow phase", func() {
		q, err := quad.New(coord.Bounds{
			TopL: coord.Cell{-16, 16},
			BotR: coord.Cell{15, -15},
		}, 3, nil)
		c.Assume(err, IsNil)

		q = q.Insert(e(0, 0, 0))
		q = q.Insert(e(1, 1, 0))

		c.Specify("will insert all entities returned", func() {
			q = q.RunNarrowPhase(quad.NarrowPhaseHandlerFn(
				func(cgrps []*quad.CollisionGroup, now stime.Time) quad.NarrowPhaseChanges {
					return quad.SliceNarrowPhaseChanges{
						e(0, -1, 0),
						e(1, 2, 0),
						e(2, 2, 1),
					}
				},
			), []*quad.CollisionGroup{nil}, stime.Time(0))

			c.Expect(len(q.QueryBounds(q.Bounds(), nil)), Equals, 3)
			c.Expect(q.QueryCell(cell(-1, 0), nil)[0].Id(), Equals, entity.Id(0))
			c.Expect(q.QueryCell(cell(2, 0), nil)[0].Id(), Equals, entity.Id(1))
			c.Expect(q.QueryCell(cell(2, 1), nil)[0].Id(), Equals, entity.Id(2))
		})
	})
}
