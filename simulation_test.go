package engine

import (
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
)

func DescribeSimulation(c gospec.Context) {
	sim := NewSimulation(40)

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

	c.Specify("Adding a player", func() {
		c.Assume(sim.nextEntityId, Equals, EntityId(0))

		pd := PlayerDef{
			Name:   "thundercleese",
			Facing: North,
			Coord:  WorldCoord{0, 0},
		}

		player := sim.addPlayer(pd)

		c.Expect(player.Id(), Equals, EntityId(0))
		c.Expect(len(sim.state.entities), Equals, 1)

		c.Specify("while the simulation is running", func() {
			sim.Start()

			pd = PlayerDef{
				Name:   "zorak",
				Facing: South,
				Coord:  WorldCoord{0, 0},
			}

			player = sim.AddPlayer(pd)

			sim.Stop()

			c.Expect(player.Id(), Equals, EntityId(1))
			c.Expect(len(sim.state.entities), Equals, 2)
		})
	})
}
