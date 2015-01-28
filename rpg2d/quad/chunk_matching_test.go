package quad_test

import (
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/quad"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeChunkMatching(c gospec.Context) {
	c.Specify("a chunk", func() {
		c.Specify("will be equal to another chunk", func() {
			entities := []entity.Entity{
				&MockEntity{0, coord.Cell{0, 0}},
				&MockEntity{1, coord.Cell{1, 1}},
				&MockEntity{2, coord.Cell{2, 2}},
				&MockEntity{3, coord.Cell{3, 3}},
				&MockEntity{4, coord.Cell{4, 4}},
				&MockEntity{5, coord.Cell{5, 5}},
			}

			chunks := []quad.Chunk{{
				Entities: entities[0:],
			}, {
				Entities: entities[0:2],
			}, {
				Entities: entities[1:4],
			}}

			c.Specify("if it contains the same entities", func() {
				expectedChunks := []quad.Chunk{{
					Entities: entities[0:],
				}, {
					Entities: entities[0:2],
				}, {
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
