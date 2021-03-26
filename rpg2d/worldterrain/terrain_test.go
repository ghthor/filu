package worldterrain

import (
	"github.com/ghthor/filu/rpg2d/coord"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeTerrainMap(c gospec.Context) {
	C := func(x, y int) coord.Cell { return coord.Cell{x, y} }

	c.Specify("a terrain map", func() {
		terrainMap, err := NewMap(coord.Bounds{
			C(-2, 3),
			C(2, -2),
		}, "G")
		c.Expect(err, IsNil)
		c.Expect(len(terrainMap.Types), Equals, 6)
		c.Expect(len(terrainMap.Types[0]), Equals, 5)

		for _, row := range terrainMap.Types {
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

			terrainMap.Types[2][3] = TT_DIRT
			terrainMap.Types[3][2] = TT_ROCK
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
			terrainMap, err = NewMap(coord.Bounds{C(0, 0), C(5, -6)}, `
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
			terrainMap.Types[0][0] = TT_ROCK

			c.Specify("directly", func() {
				c.Expect(terrainMap.Types[0][0], Equals, TT_ROCK)
				c.Expect(terrainMap.Types[1][1], Equals, TT_GRASS)
			})
			c.Specify("by cell", func() {
				c.Expect(terrainMap.Cell(C(-2, 3)), Equals, TT_ROCK)
				c.Expect(terrainMap.Cell(C(-1, 2)), Equals, TT_GRASS)
			})
		})

		c.Specify("can be sliced into a smaller rectangle", func() {
			terrainMap.Types[1][2] = TT_DIRT
			terrainMap.Types[4][3] = TT_ROCK

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
				c.Assume(terrainMap.Types[1][1], Equals, TT_GRASS)
				slice.Types[0][0] = TT_ROCK
				c.Expect(terrainMap.Types[1][1], Equals, TT_ROCK)
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

			fullMap, err := NewMap(fullBounds, `
GDR
RGD
DRG
`)
			initialSlice := fullMap.Slice(initialBounds)
			resultSlice := fullMap.Slice(resultBounds)

			actualMap, err := initialSlice.Clone()
			c.Assume(err, IsNil)
			err = actualMap.MergeDiff(initialSlice.ToState().Diff(resultSlice.ToState()))
			c.Assume(err, IsNil)
			c.Expect(actualMap.String(), Equals, resultSlice.String())
		})
	})

	c.Specify("a terrain map state", func() {
		fullMap, err := NewMap(coord.Bounds{
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

			update := terrainState.Diff(terrainState)
			c.Expect(update, IsNil)
		})

		c.Specify("that is empty", func() {
			c.Specify("will calculate a diff that contains everything", func() {
				old := &MapState{}
				terrainState := fullMap.ToState()
				diff := old.Diff(terrainState)

				c.Expect(len(diff.Slices), Equals, 1)

				slice := diff.Slices[0]

				c.Expect(slice.IsEmpty(), IsFalse)
				c.Expect(slice.Bounds, Equals, terrainState.Map.Bounds)
				c.Expect(slice.Terrain, Equals, terrainState.String())
			})
		})

		c.Specify("can calculate differences with a map that overlaps to the", func() {
			ce := func(x, y int) coord.Cell { return coord.Cell{x, y} }
			b := func(tl, br coord.Cell) coord.Bounds { return coord.Bounds{tl, br} }
			strs := func(diffs []MapStateSlice) []string {
				strs := make([]string, 0, len(diffs))
				for _, d := range diffs {
					strs = append(strs, d.Terrain)
				}
				return strs
			}

			center := fullMap.Slice(b(ce(1, -1), ce(2, -2))).ToState()

			diff := func(with coord.Bounds) *MapStateSlices {
				return center.Diff(fullMap.Slice(with).ToState())
			}

			sliceStrs := func(rows ...string) []string {
				for i, row := range rows {
					rows[i] = "\n" + row + "\n"
				}
				return rows
			}

			c.Specify("north", func() {
				update := diff(b(
					ce(1, 0),
					ce(2, -1),
				))

				c.Expect(len(update.Slices), Equals, 1)
				c.Expect(strs(update.Slices), ContainsAll, sliceStrs("RG"))
			})

			c.Specify("north & east", func() {
				update := diff(b(
					ce(2, 0),
					ce(3, -1),
				))

				c.Expect(len(update.Slices), Equals, 3)
				c.Expect(strs(update.Slices), ContainsAll, sliceStrs("G", "G", "D"))
			})

			c.Specify("east", func() {
				update := diff(b(
					ce(2, -1),
					ce(3, -2),
				))

				c.Expect(len(update.Slices), Equals, 1)
				c.Expect(strs(update.Slices), ContainsAll, sliceStrs("D\nR"))
			})

			c.Specify("south & east", func() {
				update := diff(b(
					ce(2, -2),
					ce(3, -3),
				))

				c.Expect(len(update.Slices), Equals, 3)
				c.Expect(strs(update.Slices), ContainsAll, sliceStrs("R", "R", "G"))
			})

			c.Specify("south", func() {
				update := diff(b(
					ce(1, -2),
					ce(2, -3),
				))

				c.Expect(len(update.Slices), Equals, 1)
				c.Expect(strs(update.Slices), ContainsAll, sliceStrs("GG"))
			})

			c.Specify("south & west", func() {
				update := diff(b(
					ce(0, -2),
					ce(1, -3),
				))

				c.Expect(len(update.Slices), Equals, 3)
				c.Expect(strs(update.Slices), ContainsAll, sliceStrs("D", "D", "G"))
			})

			c.Specify("west", func() {
				update := diff(b(
					ce(0, -1),
					ce(1, -2),
				))

				c.Expect(len(update.Slices), Equals, 1)
				c.Expect(strs(update.Slices), ContainsAll, sliceStrs("D\nD"))
			})

			c.Specify("north & west", func() {
				update := diff(b(
					ce(0, 0),
					ce(1, -1),
				))

				c.Expect(len(update.Slices), Equals, 3)
				c.Expect(strs(update.Slices), ContainsAll, sliceStrs("D", "G", "R"))
			})
		})
	})
}
