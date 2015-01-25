package rpg2d_test

import (
	"github.com/ghthor/engine/coord"
	"github.com/ghthor/engine/rpg2d"
	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/quad"
	"github.com/ghthor/engine/sim"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type mockActor struct{}

// Implement sim.Actor
func (mockActor) Id() int64           { return 0 }
func (mockActor) Conn() sim.ActorConn { return nil }

// Implement entity.Entity
func (mockActor) Cell() coord.Cell {
	return coord.Cell{}
}

func (mockActor) Bounds() coord.Bounds {
	return coord.Bounds{
		coord.Cell{},
		coord.Cell{},
	}
}

type mockEntityResolver struct{}

func (mockEntityResolver) EntityForActor(a sim.Actor) entity.Entity {
	return a.(entity.Entity)
}

func DescribeASimulation(c gospec.Context) {
	quad, err := quad.New(coord.Bounds{
		TopL: coord.Cell{-1000, 1000},
		BotR: coord.Cell{1000, -1000},
	}, 10, nil)

	c.Assume(err, IsNil)
	def := rpg2d.SimulationDef{
		FPS: 40,

		QuadTree: quad,

		EntityResolver: mockEntityResolver{},
	}

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
