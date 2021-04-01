package quad_test

import (
	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/entity/entitytest"
	"github.com/ghthor/filu/rpg2d/quad"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeCollision(c gospec.Context) {
	c.Specify("a collision", func() {
		c.Specify("is the same", func() {
			aptr, bptr := &entitytest.MockEntity{EntityId: 0}, &entitytest.MockEntity{EntityId: 1}
			a, b := entitytest.MockEntity{EntityId: 0}, entitytest.MockEntity{EntityId: 1}

			c.Specify("if a and b are the same", func() {
				c1, c2 := quad.NewCollision(aptr, bptr), quad.NewCollision(aptr, bptr)
				c.Expect(c1, Equals, c2)
				c.Expect(c2, Equals, c1)

				c1, c2 = quad.NewCollision(a, b), quad.NewCollision(a, b)
				c.Expect(c1, Equals, c2)
				c.Expect(c2, Equals, c1)
			})

			c.Specify("if a and b are swapped", func() {
				c1, c2 := quad.NewCollision(aptr, bptr), quad.NewCollision(bptr, aptr)
				c.Expect(c1, Equals, c2)
				c.Expect(c2, Equals, c1)

				c1, c2 = quad.NewCollision(a, b), quad.NewCollision(b, a)
				c.Expect(c1, Equals, c2)
				c.Expect(c2, Equals, c1)
			})
		})
	})
}

func DescribeCollisionIndex(c gospec.Context) {
	c.Specify("a collision index", func() {
		c.Specify("can be compared for equality", func() {
			c.Expect(quad.CollisionIndex(nil), Equals, quad.CollisionIndex(nil))

			e := []entitytest.MockEntity{
				{EntityId: 0},
				{EntityId: 1},
				{EntityId: 2},
				{EntityId: 3},
				{EntityId: 4},
			}

			col := func(a, b entity.Entity) quad.Collision { return quad.NewCollision(a, b) }

			cindex := quad.CollisionIndex{
				e[0]: []quad.Collision{
					col(e[0], e[1]),
					col(e[0], e[2]),
				},

				e[1]: []quad.Collision{
					col(e[1], e[0]),
				},

				e[2]: []quad.Collision{
					col(e[2], e[0]),
				},
			}

			c.Expect(cindex, Equals, cindex)

			cindex1 := cindex
			cindex2 := quad.CollisionIndex{
				e[0]: []quad.Collision{
					col(e[2], e[0]),
					col(e[0], e[1]),
				},

				e[1]: []quad.Collision{
					col(e[1], e[0]),
				},

				e[2]: []quad.Collision{
					col(e[0], e[2]),
				},
			}

			c.Expect(cindex1, Equals, cindex2)
		})
	})
}

func DescribeCollisionGroup(c gospec.Context) {
	entities := func() []entitytest.MockEntityWithBounds {
		c := func(x, y int) coord.Cell { return coord.Cell{x, y} }
		b := func(tl, br coord.Cell) coord.Bounds { return coord.Bounds{tl, br} }
		e := func(id entity.Id, cell coord.Cell, bounds coord.Bounds) entitytest.MockEntityWithBounds {
			return entitytest.MockEntityWithBounds{
				id, cell, bounds, 0,
			}
		}

		return []entitytest.MockEntityWithBounds{
			e(0, c(0, 0), b(c(0, 1), c(0, 0))),
			e(1, c(0, 1), b(c(0, 1), c(0, 0))),
			e(2, c(1, 1), b(c(0, 1), c(1, 1))),
			e(3, c(5, 5), b(c(5, 5), c(6, 5))),
			e(4, c(7, 5), b(c(6, 5), c(7, 5))),
		}
	}()

	collisions := func(e []entitytest.MockEntityWithBounds) []quad.Collision {
		c := func(a, b entity.Entity) quad.Collision { return quad.NewCollision(a, b) }

		return []quad.Collision{
			c(e[0], e[1]),
			c(e[0], e[2]),
			c(e[1], e[2]),

			c(e[3], e[4]),
		}
	}(entities)

	cindexs := func(e []entitytest.MockEntityWithBounds, c []quad.Collision) []quad.CollisionIndex {
		return []quad.CollisionIndex{
			{ // Collision Group 0 Index
				e[0]: []quad.Collision{
					c[0], c[1],
				},

				e[1]: []quad.Collision{
					c[0], c[2],
				},

				e[2]: []quad.Collision{
					c[1], c[2],
				},
			},

			{ // Collision Group 1 Index
				e[3]: []quad.Collision{c[3]},
				e[4]: []quad.Collision{c[3]},
			},
		}
	}(entities, collisions)

	newcg := func(collisions ...quad.Collision) *quad.CollisionGroup {
		cg := quad.NewCollisionGroup(1)

		for _, c := range collisions {
			cg.AddCollisionFromMerge(c)
		}

		return cg
	}

	cgroups := func(c []quad.Collision) []*quad.CollisionGroup {
		return []*quad.CollisionGroup{
			newcg(c[:3]...),
			newcg(c[3]),
		}
	}(collisions)

	c.Assume(len(cgroups[0].Entities()), Equals, 3)
	c.Assume(len(cgroups[0].CollisionsById), Equals, 3)

	c.Assume(len(cgroups[1].Entities()), Equals, 2)
	c.Assume(len(cgroups[1].CollisionsById), Equals, 1)

	c.Specify("a collision group", func() {
		c.Specify("can be compared for equality", func() {
			cg0equals := func(c []quad.Collision) []*quad.CollisionGroup {
				return []*quad.CollisionGroup{
					newcg(c[2], c[0], c[1]),
					newcg(c[1], c[2], c[0]),
					newcg(quad.NewCollision(
						c[1].B, c[1].A,
					), c[2], c[0]),
				}
			}(collisions)

			cg0notequals := func(c []quad.Collision) []*quad.CollisionGroup {
				return []*quad.CollisionGroup{
					newcg(c[0], c[1]),
				}
			}(collisions)

			for _, cg0 := range cg0equals {
				c.Expect(cg0, Equals, cgroups[0])
			}

			for _, cg0 := range cg0notequals {
				c.Expect(cg0, Not(Equals), cgroups[0])
			}

		})

		c.Specify("is a group of unique collisions", func() {
			for _, cg := range cgroups {
				for id1, c1 := range cg.CollisionsById {
					for id2, c2 := range cg.CollisionsById {
						// ignore the same index
						if id1 == id2 {
							continue
						}

						c.Expect(c1, Not(Equals), c2)
					}
				}
			}
		})

		c.Specify("will not add a collision it already has", func() {
			cg := cgroups[0]

			cg.AddCollisionFromMerge(collisions[0])
			c.Expect(len(cg.Entities()), Equals, 3)
			c.Expect(len(cg.CollisionsById), Equals, 3)

			cg.AddCollision(collisions[0].B, collisions[0].A)
			c.Expect(len(cg.Entities()), Equals, 3)
			c.Expect(len(cg.CollisionsById), Equals, 3)

			cg.AddCollision(
				entitytest.MockEntityWithBounds{EntityId: 10},
				entitytest.MockEntityWithBounds{EntityId: 20},
			)

			c.Expect(len(cg.Entities()), Equals, 5)
			c.Expect(len(cg.CollisionsById), Equals, 4)
		})

		c.Specify("has a list of the entities involved in the group", func() {
			for _, cg := range cgroups {
				for _, collision := range cg.CollisionsById {
					c.Expect(cg.Entities(), Contains, collision.A)
					c.Expect(cg.Entities(), Contains, collision.B)
				}
			}
		})

		c.Specify("contains only entities that are involved with collisions in the group", func() {
			for _, cg := range cgroups {
				for _, e := range cg.Entities() {
					var collisionsEntityExistsIn []quad.Collision

					for _, collision := range cg.CollisionsById {
						if collision.A == e || collision.B == e {
							collisionsEntityExistsIn = append(collisionsEntityExistsIn, collision)
						}
					}

					c.Expect(len(collisionsEntityExistsIn), Satisfies, len(collisionsEntityExistsIn) > 0)
				}
			}
		})

		c.Specify("can be used to create a collision index", func() {
			for i, cg := range cgroups {
				c.Expect(cg.CollisionIndex, Equals, cindexs[i])
			}
		})

		c.Specify("has a bounds that includes all of the entities", func() {
			for _, cg := range cgroups {
				bounds := cg.Bounds()

				for _, collision := range cg.CollisionsById {
					c.Expect(bounds.Join(collision.Bounds()), Equals, bounds)
					c.Expect(bounds.Join(collision.A.Bounds()), Equals, bounds)
					c.Expect(bounds.Join(collision.B.Bounds()), Equals, bounds)
				}

				for _, e := range cg.Entities() {
					c.Expect(bounds.Join(e.Bounds()), Equals, bounds)
				}
			}
		})
	})
}

func DescribeCollisionGroupIndex(c gospec.Context) {
	c.Specify("a collision group index", func() {
		c.Specify("can be compared for equality", func() {
			// nil == nil
			c.Expect(quad.CollisionGroupIndex(nil), Equals, quad.CollisionGroupIndex(nil))

			c.Expect(quad.CollisionGroupIndex{
				entitytest.MockEntity{EntityId: 0}: nil,
			}, Equals, quad.CollisionGroupIndex{
				entitytest.MockEntity{EntityId: 0}: nil,
			})

			cg := quad.NewCollisionGroup(1)

			c.Expect(quad.CollisionGroupIndex{
				entitytest.MockEntity{EntityId: 0}: cg,
			}, Equals, quad.CollisionGroupIndex{
				entitytest.MockEntity{EntityId: 0}: cg,
			})

			cg2 := quad.NewCollisionGroup(1)
			cg2.AddCollision(entitytest.MockEntity{EntityId: 0}, entitytest.MockEntity{EntityId: 1})

			c.Expect(quad.CollisionGroupIndex{
				entitytest.MockEntity{EntityId: 0}: cg,
			}, Not(Equals), quad.CollisionGroupIndex{
				entitytest.MockEntity{EntityId: 0}: cg2,
			})
		})
	})
}
