package engine

import (
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/time"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeDiffConn(c gospec.Context) {
	terrainMap, err := NewTerrainMap(coord.Bounds{
		coord.Cell{-10, 10},
		coord.Cell{10, -10},
	}, string(TT_GRASS))
	c.Assume(err, IsNil)

	packets := make(chan string, 1)

	conn := &DiffConn{JsonOutputConn: spyConn{packets}}
	conn.lastState = WorldStateJson{
		0,
		[]EntityJson{MockEntity{}.Json()},
		nil,
		terrainMap.Slice(coord.Bounds{
			coord.Cell{-2, 2},
			coord.Cell{2, -2},
		}).Json(),
	}

	c.Specify("stores the next state as the last state", func() {
		c.Assume(conn.lastState.Time, Equals, time.Time(0))
		conn.SendJson("update", WorldStateJson{
			Time:       1,
			TerrainMap: conn.lastState.TerrainMap,
		})
		c.Expect(conn.lastState.Time, Equals, time.Time(1))
	})

	c.Specify("the connection sends", func() {
		c.Specify("nothing if nothing has changed", func() {
			conn.SendJson("update", conn.lastState)
			c.Expect(len(packets), Equals, 0)
		})

		c.Specify("the diff from the last state and the next one", func() {
			nextState := conn.lastState.Clone()
			nextState.Time++

			c.Specify("when an entity has changed", func() {
				entity := MockEntity{}.Json().(MockEntityJson)
				entity.Name = "state has been changed"
				nextState.Entities[0] = entity

				conn.SendJson("update", nextState)
				c.Expect(len(packets), Equals, 1)
			})

			c.Specify("when an entity has been added", func() {
				nextState.Entities = append(nextState.Entities, MockEntity{id: 1}.Json())

				conn.SendJson("update", nextState)
				c.Expect(len(packets), Equals, 1)
			})

			c.Specify("when an entity has been removed", func() {
				nextState.Entities = nextState.Entities[:0]

				conn.SendJson("update", nextState)
				c.Expect(len(packets), Equals, 1)
			})

			// The viewport changing is sensed by DiffConn when the AABB of the TerrainMap has changed
			// and a Diff is run between the new terrain map and the old one
			c.Specify("when the viewport has changed", func() {
				nextState.TerrainMap = terrainMap.Slice(coord.Bounds{
					coord.Cell{-3, 2},
					coord.Cell{1, -2},
				}).Json()

				conn.SendJson("update", nextState)
				c.Expect(len(packets), Equals, 1)
			})

			// TODO
			c.Specify("when the terrain has changed", func() {})
		})
	})
}
