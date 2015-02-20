package rpg2d_test

import (
	"github.com/ghthor/engine/rpg2d"
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/quad"
	"github.com/ghthor/engine/sim/stime"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type mockActor struct {
	id int64

	cell coord.Cell
}

// Implement Actor
func (a mockActor) Entity() entity.Entity     { return a }
func (mockActor) WriteState(rpg2d.WorldState) {}

// Implement entity.Entity
func (a mockActor) Id() int64        { return a.id }
func (a mockActor) Cell() coord.Cell { return a.cell }

func (a mockActor) Bounds() coord.Bounds {
	return coord.Bounds{a.cell, a.cell}
}

func (a mockActor) ToState() entity.State             { return a }
func (a mockActor) IsDifferentFrom(entity.State) bool { return true }

type mockInputPhase struct{}

func (mockInputPhase) ApplyInputsTo(e entity.Entity, now stime.Time) []entity.Entity {
	return []entity.Entity{e}
}

type mockNarrowPhase struct{}

func (mockNarrowPhase) ResolveCollisions(c *quad.CollisionGroup, now stime.Time) ([]entity.Entity, []entity.Entity) {
	return c.Entities, nil
}

func DescribeASimulation(c gospec.Context) {
	bounds := coord.Bounds{
		TopL: coord.Cell{-1024, 1024},
		BotR: coord.Cell{1023, -1023},
	}

	quad, err := quad.New(bounds, 10, nil)

	terrainMap, err := rpg2d.NewTerrainMap(bounds, string(rpg2d.TT_GRASS))
	c.Assume(err, IsNil)

	c.Assume(err, IsNil)
	def := rpg2d.SimulationDef{
		FPS: 40,

		QuadTree:   quad,
		TerrainMap: terrainMap,

		InputPhaseHandler:  mockInputPhase{},
		NarrowPhaseHandler: mockNarrowPhase{},
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

		hs, err := rs.Halt()
		c.Expect(err, IsNil)
		c.Expect(hs.Quad(), Not(IsNil))
	})

	c.Specify("a simulation can have actors", func() {
		rs, err := def.Begin()
		c.Assume(err, IsNil)

		defer func() {
			_, err := rs.Halt()
			c.Assume(err, IsNil)
		}()

		a := mockActor{id: 1}

		c.Specify("added to it", func() {
			rs.ConnectActor(a)

			hs, err := rs.Halt()
			c.Assume(err, IsNil)

			entities := hs.Quad().QueryCell(coord.Cell{})
			c.Expect(len(entities), Equals, 1)
			c.Expect(entities[0], Equals, a)
		})

		c.Specify("removed from it", func() {
			rs.ConnectActor(a)
			rs.RemoveActor(a)

			hs, err := rs.Halt()
			c.Assume(err, IsNil)

			entities := hs.Quad().QueryCell(coord.Cell{})
			c.Expect(len(entities), Equals, 0)
		})
	})
}
