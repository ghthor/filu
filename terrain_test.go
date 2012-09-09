package engine

import (
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
)

func DescribeTerrainMap(c gospec.Context) {
	c.Specify("a terrain map", func() {
		terrainMap := NewTerrainMap(AABB{
			Cell{-2, 3},
			Cell{2, -2},
		})

		c.Expect(len(terrainMap.TerrainTypes), Equals, 6)
		c.Expect(len(terrainMap.TerrainTypes[0]), Equals, 5)

		for _, row := range terrainMap.TerrainTypes {
			for x, _ := range row {
				c.Expect(row[x], Equals, TT_GRASS)
			}
		}

		c.Specify("can be accessed", func() {
			terrainMap.TerrainTypes[0][0] = TT_ROCK

			c.Specify("directly", func() {
				c.Expect(terrainMap.TerrainTypes[0][0], Equals, TT_ROCK)
				c.Expect(terrainMap.TerrainTypes[1][1], Equals, TT_GRASS)
			})
			c.Specify("by cell", func() {
				c.Expect(terrainMap.Cell(Cell{-2, 3}), Equals, TT_ROCK)
				c.Expect(terrainMap.Cell(Cell{-1, 2}), Equals, TT_GRASS)
			})
		})

		// TODO I wonder if I can check the memory allocations during this spec
		// to see if my byte slice is large enough
		c.Specify("can be converted to a string", func() {
			c.Expect(terrainMap.String(), Equals, `
GGGGG
GGGGG
GGGGG
GGGGG
GGGGG
GGGGG
`)

			terrainMap.TerrainTypes[2][3] = TT_DIRT
			terrainMap.TerrainTypes[3][2] = TT_ROCK
			c.Expect(terrainMap.String(), Equals, `
GGGGG
GGGGG
GGGDG
GGRGG
GGGGG
GGGGG
`)
		})

		c.Specify("can be sliced into a smaller rectangle", func() {
			terrainMap.TerrainTypes[1][2] = TT_DIRT
			terrainMap.TerrainTypes[4][3] = TT_ROCK

			slice, err := terrainMap.Slice(AABB{
				Cell{-1, 2},
				Cell{1, -1},
			})
			c.Expect(err, IsNil)
			c.Expect(slice.String(), Equals, `
GDG
GGG
GGG
GGR
`)

			c.Specify("that can be sliced again", func() {
				slice, err = slice.Slice(AABB{
					Cell{-1, 2},
					Cell{0, 2},
				})

				c.Expect(err, IsNil)
				c.Expect(slice.String(), Equals, "\nGD\n")
			})

			c.Specify("that shares memory with the original slice", func() {
				c.Assume(terrainMap.TerrainTypes[1][1], Equals, TT_GRASS)
				slice.TerrainTypes[0][0] = TT_ROCK
				c.Expect(terrainMap.TerrainTypes[1][1], Equals, TT_ROCK)
			})
		})
	})
}
