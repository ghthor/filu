package rpg2d_test

import (
	"github.com/ghthor/engine/rpg2d"
	"github.com/ghthor/engine/sim"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type mockActor struct{}

func (mockActor) Id() int64           { return 0 }
func (mockActor) Conn() sim.ActorConn { return nil }

func DescribeASimulation(c gospec.Context) {
	def := rpg2d.SimulationDef{40}

	c.Specify("a simulation can be started", func() {
		rs, err := def.Begin()
		c.Expect(err, IsNil)

		_, err = rs.Halt()
		c.Expect(err, IsNil)
	})

	c.Specify("a simulation can be stopped", func() {
		rs, err := def.Begin()
		c.Assume(err, IsNil)

		_, err = rs.Halt()
		c.Expect(err, IsNil)
	})

	c.Specify("a simulation can have actors", func() {
		rs, err := def.Begin()
		c.Assume(err, IsNil)

		defer func() {
			_, err := rs.Halt()
			c.Assume(err, IsNil)
		}()

		a := mockActor{}

		c.Specify("added to it", func() {
			c.Expect(rs.ConnectActor(a), IsNil)
		})

		c.Specify("removed from it", func() {
			c.Assume(rs.ConnectActor(a), IsNil)
			c.Expect(rs.RemoveActor(a), IsNil)
		})
	})
}
