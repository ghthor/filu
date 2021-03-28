package worldstate

import (
	"fmt"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/quad/quadstate"
	"github.com/ghthor/filu/rpg2d/worldterrain"
	"github.com/ghthor/filu/sim/stime"
)

type Snapshot struct {
	Time   stime.Time
	Bounds coord.Bounds

	*quadstate.Entities

	TerrainMap *worldterrain.MapState
}

type Update struct {
	Time   stime.Time
	Bounds coord.Bounds

	Entities []*quadstate.Entity
	Removed  []*quadstate.Entity

	TerrainMapSlices *worldterrain.MapStateSlices
}

func NewSnapshot(now stime.Time, bounds coord.Bounds, defaultSize int) *Snapshot {
	return &Snapshot{
		Time:     now,
		Bounds:   bounds,
		Entities: quadstate.NewEntities(defaultSize),
	}
}

func (s *Snapshot) String() string {
	return fmt.Sprintf("%#v", s)
}

func (s *Snapshot) Clone() *Snapshot {
	clone := &Snapshot{
		Time:   s.Time,
		Bounds: s.Bounds,
		Entities: &quadstate.Entities{
			Removed:   make([]*quadstate.Entity, len(s.Removed)),
			New:       make([]*quadstate.Entity, len(s.New)),
			Changed:   make([]*quadstate.Entity, len(s.Changed)),
			Unchanged: make([]*quadstate.Entity, len(s.Unchanged)),
		},
		TerrainMap: s.TerrainMap,
	}

	copy(clone.Removed, s.Removed)
	copy(clone.New, s.New)
	copy(clone.Changed, s.Changed)
	copy(clone.Unchanged, s.Unchanged)

	return clone
}

func NewUpdate(size int) *Update {
	return &Update{
		Removed:  make([]*quadstate.Entity, 0, size),
		Entities: make([]*quadstate.Entity, 0, size),
	}
}

func (u *Update) FromSnapshot(prev, next *Snapshot) {
	u.Time = next.Time
	u.Bounds = next.Bounds

	u.TerrainMapSlices = nil

	u.Entities = append(u.Entities[:0], next.Entities.Changed...)
	u.Entities = append(u.Entities, next.Entities.New...)
	u.Removed = append(u.Removed[:0], next.Entities.Removed...)

	// Diff the TerrainMap
	if prev.TerrainMap != nil && !prev.TerrainMap.IsEmpty() {
		u.TerrainMapSlices = prev.TerrainMap.Diff(next.TerrainMap)
	} else {
		u.TerrainMapSlices = nil
	}
}

func (s *Snapshot) Set(now stime.Time, bounds coord.Bounds, quad quadstate.Quad, terrain *worldterrain.MapState) *Snapshot {
	quad.QueryBounds(bounds, s, quadstate.QueryAll)
	if terrain != nil && !terrain.IsEmpty() {
		s.TerrainMap = &worldterrain.MapState{Map: terrain.Map.Slice(bounds)}
	} else {
		s.TerrainMap = nil
	}
	return s
}

func CullForInitialState(now stime.Time, bounds coord.Bounds, quad quadstate.Quad, terrain *worldterrain.MapState, defaultSize int) *Snapshot {
	result := NewSnapshot(now, bounds, defaultSize)
	return result.Set(now, bounds, quad, terrain)
}

func (s *Snapshot) Apply(update *Update) {
	// TODO
}
