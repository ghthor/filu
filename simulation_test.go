package engine

import (
	"encoding/json"

	. "github.com/ghthor/engine/rpg2d/coord"
	. "github.com/ghthor/engine/sim/stime"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type noopConn int

func (c noopConn) SendJson(msg string, obj interface{}) error {
	c++
	return nil
}

func DescribeWorldState(c gospec.Context) {
	c.Specify("generates json compatitable state object", func() {
		worldState := newWorldState(Clock(0), Bounds{
			Cell{-3, 3},
			Cell{3, -3},
		})

		entity := MockEntity{id: 0}
		worldState.quadTree.Insert(entity)

		jsonState := worldState.Json()
		jsonState.TerrainMap.Prepare()

		c.Assume(jsonState.Time, Equals, Time(0))
		c.Assume(len(jsonState.Entities), Equals, 1)

		jsonBytes, err := json.Marshal(jsonState)
		c.Expect(err, IsNil)
		c.Expect(string(jsonBytes), Equals, `{"time":0,"entities":[{"id":0,"name":"MockEntity0","cell":{"x":0,"y":0}}],"removed":null,"terrainMap":{"bounds":{"tl":{"x":-3,"y":3},"br":{"x":3,"y":-3}},"terrain":"\nGGGGGGG\nGGGGGGG\nGGGGGGG\nGGGGGGG\nGGGGGGG\nGGGGGGG\nGGGGGGG\n"}}`)

		c.Specify("that can be cloned and modified", func() {
			worldState.quadTree.Insert(MockEntity{id: 1})
			worldState.quadTree.Insert(MockEntity{id: 2})
			worldState.quadTree.Insert(MockEntity{id: 3})
			worldState.quadTree.Insert(MockEntity{id: 4})

			jsonState = worldState.Json()
			clone := jsonState.Clone()

			// Modify the clone
			clone.Entities = append(clone.Entities[:2], clone.Entities[3:]...)

			// Check that the modification didn't effect the original
			for i, entity := range jsonState.Entities {
				e, isMockEntity := entity.(MockEntityJson)
				c.Assume(isMockEntity, IsTrue)
				c.Expect(e.Id(), Equals, EntityId(i))
			}
		})

		c.Specify("that can be culled by a bounding rectangle", func() {
			toBeCulled := []EntityJson{
				MockEntity{cell: Cell{-3, 3}}.Json(),
				MockEntity{cell: Cell{3, 3}}.Json(),
				MockEntity{cell: Cell{3, -3}}.Json(),
				MockEntity{cell: Cell{-3, -3}}.Json(),
			}

			wontBeCulled := []EntityJson{
				MockEntity{cell: Cell{-2, 2}}.Json(),
				MockEntity{cell: Cell{2, 2}}.Json(),
				MockEntity{cell: Cell{2, -2}}.Json(),
				MockEntity{cell: Cell{-2, -2}}.Json(),
			}

			jsonState.Entities = append(jsonState.Entities[:0], wontBeCulled...)
			jsonState.Entities = append(jsonState.Entities, toBeCulled...)

			jsonState = jsonState.Cull(Bounds{
				Cell{-2, 2},
				Cell{2, -2},
			})

			c.Expect(jsonState.Entities, Not(ContainsAll), toBeCulled)
			c.Expect(jsonState.Entities, ContainsAll, wontBeCulled)
			c.Expect(jsonState.TerrainMap.String(), Equals, `
GGGGG
GGGGG
GGGGG
GGGGG
GGGGG
`)
		})

		c.Specify("that can calculate the differences with a previous worldState state", func() {
			c.Specify("when there are no differences", func() {
				c.Expect(len(jsonState.Diff(jsonState).Entities), Equals, 0)
			})

			c.Specify("when an entity has changed state", func() {
				clone := jsonState.Clone()
				entity := clone.Entities[0].(MockEntityJson)
				entity.Name = "this is a state change"
				clone.Entities[0] = entity
				c.Expect(len(jsonState.Diff(clone).Entities), Equals, 1)
			})

			c.Specify("when there is a new entity", func() {
				clone := jsonState.Clone()
				clone.Entities = append(clone.Entities, MockEntity{id: 1}.Json())
				c.Expect(len(jsonState.Diff(clone).Entities), Equals, 1)
			})

			c.Specify("when an entity doesn't exist anymore", func() {
				clone := jsonState.Clone()
				clone.Entities = clone.Entities[:0]
				c.Expect(len(jsonState.Diff(clone).Removed), Equals, 1)
			})

			c.Specify("when the viewport has changed", func() {
				clone := jsonState.Clone()
				jsonState = jsonState.Cull(Bounds{
					Cell{-2, 2},
					Cell{2, -2},
				})

				clone = clone.Cull(Bounds{
					Cell{-3, 2},
					Cell{1, -2},
				})
				c.Expect(jsonState.Diff(clone).TerrainMap, Not(IsNil))
			})

			c.Specify("when the viewport hasn't changed", func() {
				clone := jsonState.Clone()
				c.Expect(jsonState.Diff(clone).TerrainMap, IsNil)
			})
		})
	})
}
