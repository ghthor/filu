package quad_test

import (
	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/entity/entitytest"
	"github.com/ghthor/filu/rpg2d/quad"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeChunkMatching(c gospec.Context) {
	c.Specify("a chunk", func() {
		c.Specify("will be equal to another chunk", func() {
			cell := func(x, y int) coord.Cell { return coord.Cell{x, y} }
			bnds := func(tl, br coord.Cell) coord.Bounds { return coord.Bounds{tl, br} }

			entities := []entity.Entity{
				&entitytest.MockEntity{0, cell(0, 0)},
				&entitytest.MockEntity{1, cell(1, 1)},
				&entitytest.MockEntity{2, cell(2, 2)},
				&entitytest.MockEntity{3, cell(3, 3)},
				&entitytest.MockEntity{4, cell(4, 4)},
				&entitytest.MockEntity{5, cell(5, 5)},
			}

			chunks := []quad.Chunk{{
				Bounds:   bnds(cell(0, 5), cell(5, 0)),
				Entities: entities[0:],
			}, {
				Bounds:   bnds(cell(0, 1), cell(1, 0)),
				Entities: entities[0:2],
			}, {
				Bounds:   bnds(cell(1, 3), cell(3, 1)),
				Entities: entities[1:4],
			}}

			c.Specify("if it contains the same entities", func() {
				expectedChunks := []quad.Chunk{{
					Bounds:   bnds(cell(0, 5), cell(5, 0)),
					Entities: entities[0:],
				}, {
					Bounds:   bnds(cell(0, 1), cell(1, 0)),
					Entities: entities[0:2],
				}, {
					Bounds:   bnds(cell(1, 3), cell(3, 1)),
					Entities: entities[1:4],
				}}

				c.Specify("in the same order", func() {
					// Copy the slices to make sure the compare
					// is working for different underlying arrays.
					for i, c := range expectedChunks {
						e := make([]entity.Entity, len(c.Entities))
						copy(e, c.Entities)
						expectedChunks[i].Entities = e
					}

					for i, _ := range chunks {
						c.Expect(chunks[i], Equals, expectedChunks[i])
					}
				})

				c.Specify("in any order", func() {
					// Copy all the slices so I can rearrange them
					// without bothering the underlying array in
					// the var `chunks`.
					e := expectedChunks[0].Entities
					jumbled := make([]entity.Entity, 0, len(e))
					jumbled = append(jumbled, e[3], e[0], e[5])
					jumbled = append(jumbled, e[1], e[4], e[2])
					c.Assume(len(jumbled), Equals, len(e))
					expectedChunks[0].Entities = jumbled

					e = expectedChunks[1].Entities
					jumbled = make([]entity.Entity, 0, len(e))
					jumbled = append(jumbled, e[1], e[0])
					c.Assume(len(jumbled), Equals, len(e))
					expectedChunks[1].Entities = jumbled

					e = expectedChunks[2].Entities
					jumbled = make([]entity.Entity, 0, len(e))
					jumbled = append(jumbled, e[0], e[2], e[1])
					c.Assume(len(jumbled), Equals, len(e))
					expectedChunks[2].Entities = jumbled

					for i, _ := range chunks {
						c.Expect(chunks[i], Equals, expectedChunks[i])
					}
				})
			})
			// TODO if the bounds are the same
			// Not really necessary at this very moment
		})
	})
}
