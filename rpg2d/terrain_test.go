package rpg2d

import (
	"github.com/ghthor/filu/rpg2d/coord"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeTerrainMap(c gospec.Context) {
	C := func(x, y int) coord.Cell { return coord.Cell{x, y} }

	c.Specify("a terrain map", func() {
		terrainMap, err := NewTerrainMap(coord.Bounds{
			C(-2, 3),
			C(2, -2),
		}, "G")
		c.Expect(err, IsNil)
		c.Expect(len(terrainMap.TerrainTypes), Equals, 6)
		c.Expect(len(terrainMap.TerrainTypes[0]), Equals, 5)

		for _, row := range terrainMap.TerrainTypes {
			for x, _ := range row {
				c.Expect(row[x], Equals, TT_GRASS)
			}
		}
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

		c.Specify("can be created with a string", func() {
			terrainMap, err = NewTerrainMap(coord.Bounds{C(0, 0), C(5, -6)}, `
RRRRRD
RRRRRD
RRRRRD
RRRRRD
RRRRRD
RRRRRD
DDDDDD
`)
			c.Expect(err, IsNil)

			c.Expect(terrainMap.String(), Equals, `
RRRRRD
RRRRRD
RRRRRD
RRRRRD
RRRRRD
RRRRRD
DDDDDD
`)
		})

		c.Specify("can be accessed", func() {
			terrainMap.TerrainTypes[0][0] = TT_ROCK

			c.Specify("directly", func() {
				c.Expect(terrainMap.TerrainTypes[0][0], Equals, TT_ROCK)
				c.Expect(terrainMap.TerrainTypes[1][1], Equals, TT_GRASS)
			})
			c.Specify("by cell", func() {
				c.Expect(terrainMap.Cell(C(-2, 3)), Equals, TT_ROCK)
				c.Expect(terrainMap.Cell(C(-1, 2)), Equals, TT_GRASS)
			})
		})

		c.Specify("can be sliced into a smaller rectangle", func() {
			terrainMap.TerrainTypes[1][2] = TT_DIRT
			terrainMap.TerrainTypes[4][3] = TT_ROCK

			slice := terrainMap.Slice(coord.Bounds{
				C(-1, 2),
				C(1, -1),
			})
			c.Expect(slice.String(), Equals, `
GDG
GGG
GGG
GGR
`)

			c.Specify("that can be sliced again", func() {
				slice = slice.Slice(coord.Bounds{
					C(-1, 2),
					C(0, 2),
				})
				c.Expect(slice.String(), Equals, "\nGD\n")
			})

			c.Specify("that shares memory with the original slice", func() {
				c.Assume(terrainMap.TerrainTypes[1][1], Equals, TT_GRASS)
				slice.TerrainTypes[0][0] = TT_ROCK
				c.Expect(terrainMap.TerrainTypes[1][1], Equals, TT_ROCK)
			})
		})

		c.Specify("can be sliced by an overlapping rectangle", func() {
			slice := terrainMap.Slice(coord.Bounds{
				C(-5, 2),
				C(2, -1),
			})
			c.Expect(slice.Bounds, Equals, coord.Bounds{
				C(-2, 2),
				C(2, -1),
			})
		})

		c.Specify("cannot be sliced by a non overlapping rectangle", func() {
			defer func() {
				e := recover()
				c.Expect(e, Equals, "invalid terrain map slicing operation: no overlap")
			}()

			terrainMap.Slice(coord.Bounds{
				C(-3000, -3000),
				C(-3000, -3001),
			})
		})

		c.Specify("can be joined", func() {
			fullBounds := coord.Bounds{C(-2, 2), C(0, 0)}
			initialBounds := coord.Bounds{C(-1, 1), C(0, 0)}
			resultBounds := coord.Bounds{C(-2, 2), C(-1, 1)}

			fullMap, err := NewTerrainMap(fullBounds, `
GDR
RGD
DRG
`)
			initialSlice := fullMap.Slice(initialBounds)
			resultSlice := fullMap.Slice(resultBounds)

			actualMap, err := initialSlice.Clone()
			c.Assume(err, IsNil)
			err = actualMap.MergeDiff(resultBounds, initialSlice.ToState().Diff(resultSlice.ToState())...)
			c.Assume(err, IsNil)
			c.Expect(actualMap.String(), Equals, resultSlice.String())
		})
	})

	c.Specify("a terrain map state", func() {
		fullMap, err := NewTerrainMap(coord.Bounds{
			C(0, 0),
			C(3, -3),
		}, `
GRGG
DDDD
DRRR
DGGR
`)
		c.Assume(err, IsNil)

		c.Specify("can calculate there are no differences", func() {
			terrainState := fullMap.ToState()

			diffs := terrainState.Diff(terrainState)
			c.Expect(diffs, ContainsExactly, []*TerrainMapState{})
		})

		c.Specify("that is empty", func() {
			c.Specify("will calculate a diff that contains everything", func() {
				old := &TerrainMapState{}
				terrainState := fullMap.ToState()
				diffs := old.Diff(terrainState)

				c.Expect(len(diffs), Equals, 1)

				diff := diffs[0]

				c.Expect(diff.IsEmpty(), IsFalse)
				c.Expect(diff.Bounds, Equals, terrainState.TerrainMap.Bounds)
				c.Expect(diff.Terrain, Equals, terrainState.String())
			})
		})

		c.Specify("can calculate differences with a map that overlaps to the", func() {
			ce := func(x, y int) coord.Cell { return coord.Cell{x, y} }
			b := func(tl, br coord.Cell) coord.Bounds { return coord.Bounds{tl, br} }
			strs := func(diffs []TerrainMapStateSlice) []string {
				strs := make([]string, 0, len(diffs))
				for _, d := range diffs {
					strs = append(strs, d.Terrain)
				}
				return strs
			}

			center := fullMap.Slice(b(ce(1, -1), ce(2, -2))).ToState()

			diffs := func(with coord.Bounds) []TerrainMapStateSlice {
				return center.Diff(fullMap.Slice(with).ToState())
			}

			sliceStrs := func(rows ...string) []string {
				for i, row := range rows {
					rows[i] = "\n" + row + "\n"
				}
				return rows
			}

			c.Specify("north", func() {
				diffs := diffs(b(
					ce(1, 0),
					ce(2, -1),
				))

				c.Expect(len(diffs), Equals, 1)
				c.Expect(strs(diffs), ContainsAll, sliceStrs("RG"))
			})

			c.Specify("north & east", func() {
				diffs := diffs(b(
					ce(2, 0),
					ce(3, -1),
				))

				c.Expect(len(diffs), Equals, 3)
				c.Expect(strs(diffs), ContainsAll, sliceStrs("G", "G", "D"))
			})

			c.Specify("east", func() {
				diffs := diffs(b(
					ce(2, -1),
					ce(3, -2),
				))

				c.Expect(len(diffs), Equals, 1)
				c.Expect(strs(diffs), ContainsAll, sliceStrs("D\nR"))
			})

			c.Specify("south & east", func() {
				diffs := diffs(b(
					ce(2, -2),
					ce(3, -3),
				))

				c.Expect(len(diffs), Equals, 3)
				c.Expect(strs(diffs), ContainsAll, sliceStrs("R", "R", "G"))
			})

			c.Specify("south", func() {
				diffs := diffs(b(
					ce(1, -2),
					ce(2, -3),
				))

				c.Expect(len(diffs), Equals, 1)
				c.Expect(strs(diffs), ContainsAll, sliceStrs("GG"))
			})

			c.Specify("south & west", func() {
				diffs := diffs(b(
					ce(0, -2),
					ce(1, -3),
				))

				c.Expect(len(diffs), Equals, 3)
				c.Expect(strs(diffs), ContainsAll, sliceStrs("D", "D", "G"))
			})

			c.Specify("west", func() {
				diffs := diffs(b(
					ce(0, -1),
					ce(1, -2),
				))

				c.Expect(len(diffs), Equals, 1)
				c.Expect(strs(diffs), ContainsAll, sliceStrs("D\nD"))
			})

			c.Specify("north & west", func() {
				diffs := diffs(b(
					ce(0, 0),
					ce(1, -1),
				))

				c.Expect(len(diffs), Equals, 3)
				c.Expect(strs(diffs), ContainsAll, sliceStrs("D", "G", "R"))
			})
		})
	})
}
