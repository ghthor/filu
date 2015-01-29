package quad_test

import (
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/quad"
	"github.com/ghthor/engine/sim/stime"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

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
		c.Specify("will create collision groups", func() {
			type testCase struct {
				entities []entity.Entity
				cgroups  []quad.CollisionGroup
			}
		})
	})

	c.Specify("the narrow phase", func() {
		c.Specify("will realize all future potentials", func() {
		})
	})
}
