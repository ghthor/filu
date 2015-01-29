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
		TopL: coord.Cell{-16, 16},
		BotR: coord.Cell{15, -15},
	}, 4, nil)
	c.Assume(err, IsNil)

	c.Specify("the input phase", func() {
		c.Specify("will remove any entites that move out of bounds", func() {

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

			q, outOfBounds = quad.RunInputPhaseOn(q, quad.InputPhaseHandlerFn(func(chunk quad.Chunk, now stime.Time) quad.Chunk {
				// Move the entity out of bounds
				for i, e := range chunk.Entities {
					if e.Id() == 1 {
						chunk.Entities[i] = MockEntity{1, coord.Cell{16, -15}}
					}
				}

				return chunk
			}), stime.Time(0))

			c.Expect(len(q.QueryBounds(q.Bounds())), Equals, 3)
			c.Expect(len(outOfBounds), Equals, 1)

			c.Expect(q.QueryCell(coord.Cell{-16, 16})[0].Id(), Equals, int64(0))
			c.Expect(outOfBounds[0].Id(), Equals, int64(1))
			c.Expect(q.QueryCell(coord.Cell{-1, 1})[0].Id(), Equals, int64(2))
			c.Expect(q.QueryCell(coord.Cell{0, 0})[0].Id(), Equals, int64(3))

		})
	})

	c.Specify("the broad phase", func() {
		c.Specify("will create chunks of interest", func() {
			type testCase struct {
				entities       []entity.Entity
				expectedChunks []quad.Chunk
			}

			newTestCase := func(chunks []quad.Chunk) testCase {
				var entities []entity.Entity

				for _, chunk := range chunks {
					entities = append(entities, chunk.Entities...)
				}

				return testCase{
					entities,
					chunks,
				}
			}

			testCases := func() []testCase {
				c := func(x, y int) coord.Cell { return coord.Cell{x, y} }
				b := func(tl, br coord.Cell) coord.Bounds { return coord.Bounds{tl, br} }

				chunks := []quad.Chunk{{
					Entities: []entity.Entity{
						&MockEntityWithBounds{
							0,
							c(-5, 5),
							b(c(-5, 5), c(-4, 5)),
						},
						&MockEntityWithBounds{
							1,
							c(-5, 6),
							b(c(-4, 6), c(-4, 5)),
						},
					},
				}, {
					Entities: []entity.Entity{
						&MockEntityWithBounds{
							2,
							c(5, 5),
							b(c(5, 5), c(6, 5)),
						},
						&MockEntityWithBounds{
							3,
							c(6, 5),
							b(c(6, 5), c(6, 5)),
						},
						&MockEntityWithBounds{
							4,
							c(6, 4),
							b(c(5, 5), c(6, 4)),
						},
					},
				}}

				return []testCase{
					newTestCase(chunks[0:1]),
					newTestCase(chunks[1:2]),
				}
			}()

			makeQuad := func(entities []entity.Entity) quad.Quad {
				q, err := quad.New(coord.Bounds{
					coord.Cell{-16, 16},
					coord.Cell{15, -15},
				}, 4, nil)
				c.Assume(err, IsNil)

				for _, e := range entities {
					q = q.Insert(e)
				}

				return q
			}

			for _, testCase := range testCases {
				q := makeQuad(testCase.entities)
				var actualChunks []quad.Chunk

				q, actualChunks = quad.RunBroadPhaseOn(q, stime.Time(0))
				c.Expect(len(actualChunks), Equals, len(testCase.expectedChunks))
				c.Expect(actualChunks, ContainsExactly, testCase.expectedChunks)
			}
		})
	})

	c.Specify("the narrow phase", func() {
		c.Specify("will realize all future potentials", func() {
		})
	})
}
