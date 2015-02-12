package rpg2d

import (
	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/quad"
	"github.com/ghthor/engine/sim/stime"
)

type World struct {
	time     stime.Time
	quadTree quad.Quad
	terrain  TerrainMap
}

func newWorld(clock stime.Clock, quad quad.Quad, terrain TerrainMap) *World {
	return &World{
		time:     clock.Now(),
		quadTree: quad,
		terrain:  terrain,
	}
}

func (w *World) Insert(e entity.Entity) {
	w.quadTree = w.quadTree.Insert(e)
}

func (w *World) Remove(e entity.Entity) {
	w.quadTree = w.quadTree.Remove(e)
}

func (w World) ToState() WorldState {
	entities := w.quadTree.QueryBounds(w.quadTree.Bounds())
	s := WorldState{
		w.time,
		make([]entity.State, len(entities)),
		nil,
		nil,
	}

	i := 0
	for _, e := range entities {
		s.Entities[i] = e.ToState()
		i++
	}

	terrain := w.terrain.ToState()
	if !terrain.IsEmpty() {
		s.TerrainMap = terrain
	}
	return s
}
