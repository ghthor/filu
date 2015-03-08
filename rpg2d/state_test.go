// +build ignore
package rpg2d

import (
	"encoding/json"

	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/entity/entitytest"
	"github.com/ghthor/engine/rpg2d/quad"
	"github.com/ghthor/engine/sim/stime"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeWorldState(c gospec.Context) {
	c.Specify("generates json compatitable state object", func() {
		quadTree, err := quad.New(coord.Bounds{
			coord.Cell{-4, 4},
			coord.Cell{3, -3},
		}, 20, nil)
		c.Assume(err, IsNil)

		terrain, err := NewTerrainMap(quadTree.Bounds(), string(TT_GRASS))
		c.Assume(err, IsNil)

		world := newWorld(stime.Time(0), quadTree, terrain)

		mockEntity := entitytest.MockEntity{EntityId: 0}
		world.Insert(mockEntity)

		worldState := world.ToState()

		c.Assume(worldState.Time, Equals, stime.Time(0))
		c.Assume(len(worldState.Entities), Equals, 1)

		jsonBytes, err := json.Marshal(worldState)
		c.Expect(err, IsNil)
		c.Expect(string(jsonBytes), Equals, `{"time":0,"bounds":{"tl":{"x":-4,"y":4},"br":{"x":3,"y":-3}},"entities":[{"id":0,"name":"MockEntity0","cell":{"x":0,"y":0}}],"terrainMap":{"bounds":{"tl":{"x":-4,"y":4},"br":{"x":3,"y":-3}},"terrain":"\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\n"}}`)

		c.Specify("that can be cloned and modified", func() {
			world.Insert(entitytest.MockEntity{EntityId: 1})
			world.Insert(entitytest.MockEntity{EntityId: 2})
			world.Insert(entitytest.MockEntity{EntityId: 3})
			world.Insert(entitytest.MockEntity{EntityId: 4})

			worldState = world.ToState()
			clone := worldState.Clone()

			// Modify the clone
			clone.Entities = append(clone.Entities[:2], clone.Entities[3:]...)

			// Check that the modification didn't effect the original
			for i, e := range worldState.Entities {
				e, isMockEntity := e.(entitytest.MockEntityState)
				c.Assume(isMockEntity, IsTrue)
				c.Expect(e.Id(), Equals, entity.Id(i))
			}
		})

		c.Specify("that can be culled by a bounding rectangle", func() {
			toBeCulled := []entity.State{
				entitytest.MockEntity{EntityCell: coord.Cell{-3, 3}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{3, 3}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{3, -3}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{-3, -3}}.ToState(),
			}

			wontBeCulled := []entity.State{
				entitytest.MockEntity{EntityCell: coord.Cell{-2, 2}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{2, 2}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{2, -2}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{-2, -2}}.ToState(),
			}

			worldState.Entities = append(worldState.Entities[:0], wontBeCulled...)
			worldState.Entities = append(worldState.Entities, toBeCulled...)

			worldState = worldState.Cull(coord.Bounds{
				coord.Cell{-2, 2},
				coord.Cell{2, -2},
			})

			c.Expect(worldState.Entities, Not(ContainsAll), toBeCulled)
			c.Expect(worldState.Entities, ContainsAll, wontBeCulled)
			c.Expect(worldState.TerrainMap.String(), Equals, `
GGGGG
GGGGG
GGGGG
GGGGG
GGGGG
`)
		})

		c.Specify("that can calculate the differences with a previous worldState state", func() {
			c.Specify("when there are no differences", func() {
				c.Expect(len(worldState.Diff(worldState).Entities), Equals, 0)
			})

			c.Specify("when an entity has changed state", func() {
				clone := worldState.Clone()
				entity := clone.Entities[0].(entitytest.MockEntityState)

				// This is a state change
				entity.EntityCell = coord.Cell{-1, 0}
				clone.Entities[0] = entity

				c.Expect(len(worldState.Diff(clone).Entities), Equals, 1)
			})

			c.Specify("when there is a new entity", func() {
				clone := worldState.Clone()
				clone.Entities = append(clone.Entities, entitytest.MockEntity{EntityId: 1}.ToState())
				c.Expect(len(worldState.Diff(clone).Entities), Equals, 1)
			})

			c.Specify("when an entity doesn't exist anymore", func() {
				clone := worldState.Clone()
				clone.Entities = clone.Entities[:0]
				c.Expect(len(worldState.Diff(clone).Removed), Equals, 1)
			})

			c.Specify("when the viewport has changed", func() {
				clone := worldState.Clone()
				worldState = worldState.Cull(coord.Bounds{
					coord.Cell{-2, 2},
					coord.Cell{2, -2},
				})

				// TODO Specify all 4 directions and all 4 corners
				clone = clone.Cull(coord.Bounds{
					coord.Cell{-3, 2},
					coord.Cell{1, -2},
				})
				c.Expect(worldState.Diff(clone).TerrainMapSlices, Not(IsNil))
			})

			c.Specify("when the viewport hasn't changed", func() {
				clone := worldState.Clone()
				c.Expect(worldState.Diff(clone).TerrainMapSlices, ContainsExactly, []*TerrainMapState{})
			})
		})
	})
}
