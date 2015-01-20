package engine

import (
	"encoding/json"

	. "github.com/ghthor/engine/coord"
	. "github.com/ghthor/engine/time"

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
		worldState := newWorldState(Clock(0), AABB{
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

			jsonState = jsonState.Cull(AABB{
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
				jsonState = jsonState.Cull(AABB{
					Cell{-2, 2},
					Cell{2, -2},
				})

				clone = clone.Cull(AABB{
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

func DescribeDiffConn(c gospec.Context) {
	terrainMap, err := NewTerrainMap(AABB{
		Cell{-10, 10},
		Cell{10, -10},
	}, string(TT_GRASS))
	c.Assume(err, IsNil)

	packets := make(chan string, 1)

	conn := &DiffConn{JsonOutputConn: spyConn{packets}}
	conn.lastState = WorldStateJson{
		0,
		[]EntityJson{MockEntity{}.Json()},
		nil,
		terrainMap.Slice(AABB{
			Cell{-2, 2},
			Cell{2, -2},
		}).Json(),
	}

	c.Specify("stores the next state as the last state", func() {
		c.Assume(conn.lastState.Time, Equals, Time(0))
		conn.SendJson("update", WorldStateJson{
			Time:       1,
			TerrainMap: conn.lastState.TerrainMap,
		})
		c.Expect(conn.lastState.Time, Equals, Time(1))
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
				nextState.TerrainMap = terrainMap.Slice(AABB{
					Cell{-3, 2},
					Cell{1, -2},
				}).Json()

				conn.SendJson("update", nextState)
				c.Expect(len(packets), Equals, 1)
			})

			// TODO
			c.Specify("when the terrain has changed", func() {})
		})
	})
}

func DescribeSimulation(c gospec.Context) {
	sim := newSimulation(40)

	c.Specify("starting and stopping", func() {

		sim.Start()
		c.Expect(sim.IsRunning(), IsTrue)

		sim.Stop()
		c.Expect(sim.IsRunning(), IsFalse)
	})

	c.Specify("clock ticks during each step", func() {
		c.Assume(sim.clock, Equals, Clock(0))

		sim.step()

		c.Expect(sim.clock, Equals, Clock(1))
	})

	c.Specify("Adding and removing players", func() {
		c.Assume(sim.nextEntityId, Equals, EntityId(0))

		// Need a Client endpoint
		var conn noopConn

		pd := PlayerDef{
			Name:          "thundercleese",
			Facing:        North,
			Cell:          Cell{0, 0},
			MovementSpeed: 40,
			Conn:          conn,
		}

		c.Specify("adding", func() {
			player := sim.addPlayer(pd)

			c.Expect(player.Id(), Equals, EntityId(0))
			c.Expect(sim.state.quadTree.Contains(player), IsTrue)
			c.Expect(len(sim.clients), Equals, 1)
		})

		c.Specify("adding while the simulation is running", func() {
			sim.Start()

			player := sim.AddPlayer(pd)

			sim.Stop()

			c.Expect(player.Id(), Equals, EntityId(0))
			c.Expect(sim.state.quadTree.Contains(player), IsTrue)
			c.Expect(len(sim.clients), Equals, 1)
		})

		c.Specify("removing", func() {
			player := sim.addPlayer(pd)

			c.Assume(player.Id(), Equals, EntityId(0))
			c.Assume(sim.state.quadTree.Contains(player), IsTrue)
			c.Assume(len(sim.clients), Equals, 1)

			removed := sim.removePlayer(player)

			c.Expect(removed, Equals, player)
			c.Expect(sim.state.quadTree.Contains(player), IsFalse)
			c.Expect(len(sim.clients), Equals, 0)
		})

		c.Specify("removing while the simulation is running", func() {
			player := sim.addPlayer(pd)

			c.Assume(player.Id(), Equals, EntityId(0))
			c.Assume(sim.state.quadTree.Contains(player), IsTrue)
			c.Assume(len(sim.clients), Equals, 1)

			sim.Start()

			sim.RemovePlayer(player)

			sim.Stop()

			c.Expect(sim.state.quadTree.Contains(player), IsFalse)
			c.Expect(len(sim.clients), Equals, 0)
		})

		c.Specify("player removes self while simulation is running", func() {
			player := sim.addPlayer(pd)

			c.Assume(player.Id(), Equals, EntityId(0))
			c.Assume(sim.state.quadTree.Contains(player), IsTrue)
			c.Assume(len(sim.clients), Equals, 1)

			sim.Start()

			player.Disconnect()

			sim.Stop()

			c.Expect(sim.state.quadTree.Contains(player), IsFalse)
			c.Expect(len(sim.clients), Equals, 0)
		})

		// TODO drain the newPlayer/dcedPlayer channels after the loop has broken
		c.Specify("when the simulation is stopping shouldn't block", func() {})
	})

	c.Specify("simulation loop runs at the intended fps", func() {})
}
