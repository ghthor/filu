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
	q, err := quad.New(coord.Bounds{
		TopL: coord.Cell{-10, 9},
		BotR: coord.Cell{9, -10},
	}, 4, nil)
	c.Assume(err, IsNil)

	c.Specify("the input phase", func() {
		c.Specify("will remove any entites that move out of bounds", func() {

			// A Single Entity
			q = q.Insert(MockEntity{0, coord.Cell{-10, 9}})
			c.Assume(len(q.QueryBounds(q.Bounds())), Equals, 1)

			q, outOfBounds := quad.RunInputPhaseOn(q, quad.InputPhaseHandlerFn(func(chunk quad.Chunk, now stime.Time) quad.Chunk {
				c.Assume(len(chunk.Entities), Equals, 1)

				// Move the entity out of bounds
				chunk.Entities[0] = MockEntity{0, coord.Cell{-11, 9}}

				return chunk
			}), stime.Time(0))

			c.Expect(len(outOfBounds), Equals, 1)
			c.Expect(len(q.QueryBounds(q.Bounds())), Equals, 0)

			// Multiple entities
			q = q.Insert(MockEntity{0, coord.Cell{-10, 9}})
			q = q.Insert(MockEntity{1, coord.Cell{9, -10}})
			q = q.Insert(MockEntity{2, coord.Cell{5, -1}})

			q, outOfBounds = quad.RunInputPhaseOn(q, quad.InputPhaseHandlerFn(func(chunk quad.Chunk, now stime.Time) quad.Chunk {
				// Move the entity out of bounds
				for i, e := range chunk.Entities {
					if e.Id() == 1 {
						chunk.Entities[i] = MockEntity{1, coord.Cell{10, -10}}
					}
				}

				return chunk
			}), stime.Time(0))

			c.Expect(len(q.QueryBounds(q.Bounds())), Equals, 2)
			c.Expect(q.QueryCell(coord.Cell{-10, 9})[0].Id(), Equals, int64(0))
			c.Expect(q.QueryCell(coord.Cell{5, -1})[0].Id(), Equals, int64(2))

			c.Expect(len(outOfBounds), Equals, 1)
			c.Expect(outOfBounds[0].Id(), Equals, int64(1))

		})
	})

	c.Specify("the broad phase", func() {
		c.Specify("will create chunks of interest", func() {
			type testCase struct {
				entities       []entity.Entity
				expectedChunks []quad.Chunk
			}

			cell := func(x, y int) coord.Cell { return coord.Cell{x, y} }
			bounds := func(tl, br coord.Cell) coord.Bounds { return coord.Bounds{tl, br} }
			chunk := func(e []entity.Entity) quad.Chunk { return quad.Chunk{Entities: e} }

			entities := []entity.Entity{
				&MockEntityWithBounds{
					0,
					cell(-5, 5),
					bounds(cell(-5, 5), cell(-4, 5)),
				},
				&MockEntityWithBounds{
					1,
					cell(-5, 6),
					bounds(cell(-4, 6), cell(-4, 5)),
				},
				&MockEntityWithBounds{
					2,
					cell(5, 5),
					bounds(cell(5, 5), cell(6, 5)),
				},
				&MockEntityWithBounds{
					3,
					cell(6, 5),
					bounds(cell(6, 5), cell(6, 5)),
				},
				&MockEntityWithBounds{
					4,
					cell(6, 4),
					bounds(cell(5, 5), cell(6, 4)),
				}}

			chunks := []quad.Chunk{
				chunk(entities[0:2]),
				chunk(entities[2:4]),
			}

			testCases := []testCase{{
				entities[0:2],
				chunks[0:1],
			}, {
				entities[0:4],
				chunks[0:2],
			}}

			for _, testCase := range testCases {
				for _, e := range testCase.entities {
					q = q.Insert(e)
				}

				var actualChunks []quad.Chunk

				q, actualChunks = quad.RunBroadPhaseOn(q, stime.Time(0))
				c.Expect(len(actualChunks), Equals, len(testCase.expectedChunks))

				c.Expect(actualChunks, ContainsAll, testCase.expectedChunks)
			}
		})
	})

	c.Specify("the narrow phase", func() {
		c.Specify("will realize all future potentials", func() {
		})
	})
}
