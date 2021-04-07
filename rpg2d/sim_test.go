package rpg2d_test

import (
	"github.com/ghthor/filu/rpg2d"
	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/quad"
	"github.com/ghthor/filu/rpg2d/quadstate"
	"github.com/ghthor/filu/rpg2d/worldterrain"
	"github.com/ghthor/filu/sim/stime"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type mockActorEntity struct {
	id    entity.Id
	cell  coord.Cell
	flags entity.Flag
}

type mockActor struct {
	id rpg2d.ActorId
	mockActorEntity
}

// Implement Actor
func (a mockActor) Id() rpg2d.ActorId     { return a.id }
func (a mockActor) Entity() entity.Entity { return a.mockActorEntity }
func (mockActor) WriteStateNext(stime.Time, quadstate.Quad, *worldterrain.MapState, chan<- quadstate.EncodingRequest) {
}

// Implement entity.Entity
func (a mockActorEntity) Id() entity.Id          { return a.id }
func (a mockActorEntity) EntityId() entity.Id    { return a.id }
func (a mockActorEntity) Cell() coord.Cell       { return a.cell }
func (a mockActorEntity) EntityCell() coord.Cell { return a.cell }
func (a mockActorEntity) Flags() entity.Flag     { return a.flags }
func (a mockActorEntity) Bounds() coord.Bounds {
	return coord.Bounds{a.cell, a.cell}
}

func (a mockActorEntity) ToState() entity.State             { return a }
func (a mockActorEntity) IsDifferentFrom(entity.State) bool { return true }

type mockInputPhase struct{}

func (mockInputPhase) ApplyInputsTo(e entity.Entity, now stime.Time, changes quad.InputPhaseChanges) entity.Entity {
	return e
}

type mockNarrowPhase struct{}

func (mockNarrowPhase) ResolveCollisions(cgrps []*quad.CollisionGroup, now stime.Time) quad.NarrowPhaseChanges {
	return quad.SliceNarrowPhaseChanges(cgrps[0].Entities)
}

func DescribeASimulation(c gospec.Context) {
	bounds := coord.Bounds{
		TopL: coord.Cell{-1024, 1024},
		BotR: coord.Cell{1023, -1023},
	}

	quad, err := quad.New(bounds, 10, nil)

	terrainMap, err := worldterrain.NewMap(bounds, string(worldterrain.TT_GRASS))
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

		a := mockActor{
			id: 1,
			mockActorEntity: mockActorEntity{
				id: 2,
			},
		}

		c.Specify("added to it", func() {
			rs.ConnectActor(a)

			hs, err := rs.Halt()
			c.Assume(err, IsNil)

			entities := hs.Quad().QueryCell(coord.Cell{}, nil)
			c.Expect(len(entities), Equals, 1)
			c.Expect(entities[0], Equals, a.Entity())
		})

		c.Specify("removed from it", func() {
			rs.ConnectActor(a)
			rs.RemoveActor(a)

			hs, err := rs.Halt()
			c.Assume(err, IsNil)

			entities := hs.Quad().QueryCell(coord.Cell{}, nil)
			c.Expect(len(entities), Equals, 1)
			c.Expect(entities[0].ToState(), Equals, entity.RemovedState{a.EntityId(), a.Cell()})
		})
	})
}
