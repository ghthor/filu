package quad_test

import (
	"github.com/ghthor/engine/coord"
	"github.com/ghthor/engine/rpg2d/quad"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribePhase(c gospec.Context) {
	q, err := quad.New(coord.Bounds{
		TopL: coord.Cell{-10, 10},
		BotR: coord.Cell{10, -10},
	}, 1, nil)
	c.Assume(err, IsNil)

	c.Specify("the input phase", func() {
		c.Specify("will remove any entites that move out of bounds", func() {

			// A Single Entity
			q = q.Insert(MockEntity{0, coord.Cell{-10, 10}})
			c.Assume(len(q.QueryBounds(q.Bounds())), Equals, 1)

			q, outOfBounds := q.RunInputPhase(quad.InputPhaseHandlerFn(func(chunk quad.Chunk) quad.Chunk {
				c.Assume(len(chunk.Entities), Equals, 1)

				// Move the entity out of bounds
				chunk.Entities[0] = MockEntity{0, coord.Cell{-11, 10}}

				return chunk
			}))

			c.Expect(len(outOfBounds), Equals, 1)
			c.Expect(len(q.QueryBounds(q.Bounds())), Equals, 0)

			// Multiple entities
			q = q.Insert(MockEntity{0, coord.Cell{-10, 10}})
			q = q.Insert(MockEntity{1, coord.Cell{10, -10}})
			q = q.Insert(MockEntity{2, coord.Cell{5, -1}})

			q, outOfBounds = q.RunInputPhase(quad.InputPhaseHandlerFn(func(chunk quad.Chunk) quad.Chunk {
				// Move the entity out of bounds
				for i, e := range chunk.Entities {
					if e.Id() == 1 {
						chunk.Entities[i] = MockEntity{1, coord.Cell{11, -10}}
					}
				}

				return chunk
			}))

			c.Expect(len(q.QueryBounds(q.Bounds())), Equals, 2)
			c.Expect(q.QueryCell(coord.Cell{-10, 10})[0].Id(), Equals, int64(0))
			c.Expect(q.QueryCell(coord.Cell{5, -1})[0].Id(), Equals, int64(2))

			c.Expect(len(outOfBounds), Equals, 1)
			c.Expect(outOfBounds[0].Id(), Equals, int64(1))

		})
	})

	c.Specify("the broad phase", func() {
		c.Specify("will create chunks of interest", func() {
		})
	})

	c.Specify("the narrow phase", func() {
		c.Specify("will realize all future potentials", func() {
		})
	})
}
