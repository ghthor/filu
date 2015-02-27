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

func newWorld(now stime.Time, quad quad.Quad, terrain TerrainMap) *World {
	return &World{
		time:     now,
		quadTree: quad,
		terrain:  terrain,
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

func (w World) ToState() WorldState {
	entities := w.quadTree.QueryBounds(w.quadTree.Bounds())
	s := WorldState{
		Type:       ST_FULL.String(),
		Time:       w.time,
		Entities:   make([]entity.State, len(entities)),
		Removed:    nil,
		TerrainMap: nil,
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
