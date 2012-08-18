package engine

import (
	"../server/protocol/encoding"
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
)

type noopConn int

func (c noopConn) SendMessage(msg, payload string) error {
	c++
	return nil
}

type spyConn struct {
	packets chan string
}

func (c spyConn) SendMessage(msg, payload string) error {
	c.packets <- msg + ":" + payload
	return nil
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

	c.Specify("Adding a player", func() {
		c.Assume(sim.nextEntityId, Equals, EntityId(0))

		// Need a Client endpoint
		var conn noopConn

		pd := PlayerDef{
			Name:   "thundercleese",
			Facing: North,
			Coord:  WorldCoord{0, 0},
			Conn:   conn,
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
				Conn:   conn,
			}

			player = sim.AddPlayer(pd)

			sim.Stop()

			c.Expect(player.Id(), Equals, EntityId(1))
			c.Expect(len(sim.state.entities), Equals, 2)
		})
	})

	c.Specify("simulation loop runs at the intended fps", nil)
}

func DescribeWorldState(c gospec.Context) {
	c.Specify("processes movement requests and generates appropiate actions", nil)
}

func DescribePlayer(c gospec.Context) {
	conn := spyConn{make(chan string)}

	player := &Player{
		Name:     "thundercleese",
		entityId: 0,
		mi:       newMotionInfo(WorldCoord{0, 0}, North),
		conn:     conn,
	}

	player.mux()

	c.Specify("motionInfo becomes locked when accessed by the simulation until the worldstate is published", func() {
		_ = player.motionInfo()

		locked := make(chan bool)

		go func() {
			select {
			case player.collectInput <- encoding.Packet{}:
				panic("MotionInfo not locked")
			case <-conn.packets:
				locked <- true
			}
		}()

		player.SendWorldState(newWorldState(Clock(0)))
		c.Expect(<-locked, IsTrue)

		c.Specify("and is unlocked afterwards", func() {
			select {
			case player.collectInput <- encoding.Packet{}:
			default:
				panic("MotionInfo not unlocked")
			}
		})
	})
}