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
			Entities:          make(entity.StateSlice, 0, defaultEntitiesSize),
			EntitiesNew:       make(entity.StateSlice, 0, defaultEntitiesSize),
			EntitiesChanged:   make(entity.StateSlice, 0, defaultEntitiesSize),
			EntitiesUnchanged: make(entity.StateSlice, 0, defaultEntitiesSize),
			EntitiesRemoved:   make(entity.StateSlice, 0, defaultEntitiesSize),
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
	now := world.time

	// Reuse existing slices
	nextState := WorldState{
		Time:              now,
		Bounds:            world.quadTree.Bounds(),
		EntitiesRemoved:   world.state.EntitiesRemoved[:0],
		EntitiesNew:       world.state.EntitiesNew[:0],
		EntitiesChanged:   world.state.EntitiesChanged[:0],
		EntitiesUnchanged: world.state.EntitiesUnchanged[:0],
		Entities:          world.state.Entities[:0],
	}

	entities := world.quadTree.QueryBounds(world.quadTree.Bounds())
	for _, e := range entities {
		flags := e.Flags()
		entityState := e.ToState()
		nextState.Entities = append(nextState.Entities, entityState)

		if flags&entity.FlagRemoved != 0 {
			if e.(entity.Removed).RemovedAt == now {
				nextState.EntitiesRemoved = append(nextState.EntitiesRemoved, entityState)
			}
			continue
		}

		if flags&entity.FlagNew != 0 {
			nextState.EntitiesNew = append(nextState.EntitiesNew, entityState)
			continue
		}

		if entity, canChange := e.(entity.CanChange); canChange {
			if entity.HasChanged(entityState, now) {
				nextState.EntitiesChanged = append(nextState.EntitiesChanged, entityState)
				continue
			}
		}

		nextState.EntitiesUnchanged = append(nextState.EntitiesUnchanged, entityState)
	}

	terrain := world.terrain.ToState()
	if !terrain.IsEmpty() {
		// Handle TerrainMap
		nextState.TerrainMap = terrain
	}

	return nextState
}
