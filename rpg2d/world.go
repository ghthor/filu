package rpg2d

import (
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/quad"
	"github.com/ghthor/filu/sim/stime"
)

type World struct {
	time     stime.Time
	quadTree quad.Quad
	terrain  TerrainMap

	state WorldState
}

func NewWorld(now stime.Time, quad quad.Quad, terrain TerrainMap) *World {
	const defaultEntitiesSize = 300
	return &World{
		time:     now,
		quadTree: quad,
		terrain:  terrain,

		state: WorldState{
			Entities: make(entity.StateSlice, 0, defaultEntitiesSize),
		},
	}
}

type stepToFn func(quad.Quad, stime.Time) quad.Quad

func (w *World) stepTo(t stime.Time, stepTo stepToFn) {
	w.quadTree = stepTo(w.quadTree, t)

	w.time = t
}

func (w *World) Insert(e entity.Entity) {
	w.quadTree = w.quadTree.Insert(e)
}

func (w *World) Remove(e entity.Entity) {
	w.quadTree = w.quadTree.Remove(e)
}

func (world World) ToState() WorldState {
	// Reuse existing slices
	state := WorldState{
		Time:     world.time,
		Bounds:   world.quadTree.Bounds(),
		Entities: world.state.Entities[:0],
	}

	entities := world.quadTree.QueryBounds(world.quadTree.Bounds())
	for _, e := range entities {
		state.Entities = append(state.Entities, e.ToState())
	}

	terrain := world.terrain.ToState()
	if !terrain.IsEmpty() {
		// Handle TerrainMap
		state.TerrainMap = terrain
	}

	return state
}
