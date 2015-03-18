package rpg2d_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/ghthor/engine/rpg2d"
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/entity/entitytest"
	"github.com/ghthor/engine/rpg2d/quad"
	"github.com/ghthor/engine/rpg2d/rpg2dtest"
	"github.com/ghthor/engine/sim/stime"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func init() {
	gob.Register(entitytest.MockEntityState{})
}

func DescribeWorldState(c gospec.Context) {
	quadTree, err := quad.New(coord.Bounds{
		coord.Cell{-4, 4},
		coord.Cell{3, -3},
	}, 20, nil)
	c.Assume(err, IsNil)

	terrain, err := rpg2d.NewTerrainMap(quadTree.Bounds(), string(rpg2d.TT_GRASS))
	c.Assume(err, IsNil)

	world := rpg2d.NewWorld(stime.Time(0), quadTree, terrain)

	mockEntity := entitytest.MockEntity{EntityId: 0}
	world.Insert(mockEntity)

	worldState := world.ToState()

	c.Assume(worldState.Time, Equals, stime.Time(0))
	c.Assume(len(worldState.Entities), Equals, 1)

	c.Specify("a world state", func() {
		c.Specify("can be encoded as json", func() {
			jsonBytes, err := json.Marshal(worldState)
			c.Expect(err, IsNil)
			c.Expect(string(jsonBytes), Equals, `{"time":0,"bounds":{"tl":{"x":-4,"y":4},"br":{"x":3,"y":-3}},"entities":[{"id":0,"name":"MockEntity0","cell":{"x":0,"y":0}}],"terrainMap":{"bounds":{"tl":{"x":-4,"y":4},"br":{"x":3,"y":-3}},"terrain":"\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\nGGGGGGGG\n"}}`)
		})

		func() {
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			enc := gob.NewEncoder(buf)

			c.Specify("can be encoded as a gob object", func() {
				c.Expect(enc.Encode(worldState), IsNil)
			})

			c.Specify("can be decoded from a gob object", func() {
				dec := gob.NewDecoder(buf)
				c.Assume(enc.Encode(worldState), IsNil)

				state := rpg2d.WorldState{}
				c.Expect(dec.Decode(&state), IsNil)
				c.Expect(state, rpg2dtest.StateEquals, worldState)
			})
		}()

		c.Specify("can be cloned and modified", func() {
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
				c.Expect(e.EntityId(), Equals, entity.Id(i))
			}
		})

		c.Specify("can be culled by a bounding rectangle", func() {
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

			c.Expect(worldState.Bounds, Equals, coord.Bounds{
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

		c.Specify("can calculate the differences with a previous worldState state", func() {
			c.Specify("when there are no differences", func() {
				c.Expect(len(worldState.Diff(worldState).Entities), Equals, 0)
			})

			c.Specify("when an entity has changed state", func() {
				clone := worldState.Clone()
				entity := clone.Entities[0].(entitytest.MockEntityState)

				// This is a state change
				entity.Cell = coord.Cell{-1, 0}
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
				c.Expect(worldState.Diff(clone).TerrainMapSlices, ContainsExactly, []*rpg2d.TerrainMapState{})
			})
		})

		c.Specify("can be updated with a world state diff", func() {
			c.Specify("that contains a new entity", func() {
				world.Insert(entitytest.MockEntity{EntityId: 1})
				nextState := world.ToState()
				diff := worldState.Diff(nextState)

				worldState.Apply(diff)
				c.Expect(worldState, rpg2dtest.StateEquals, nextState)
			})

			c.Specify("that removes an entity", func() {
				world.Remove(mockEntity)
				nextState := world.ToState()
				diff := worldState.Diff(nextState)

				worldState.Apply(diff)
				c.Expect(worldState, rpg2dtest.StateEquals, nextState)
			})
		})
	})
}
