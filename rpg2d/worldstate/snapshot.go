package worldstate

import (
	"fmt"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/quadstate"
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

	RemovedIds []entity.Id
	Removed    []*quadstate.Entity
	Entities   []*quadstate.Entity

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
		Time:       s.Time,
		Bounds:     s.Bounds,
		Entities:   &quadstate.Entities{},
		TerrainMap: s.TerrainMap,
	}

	for t := range s.ByType {
		clone.ByType[t] = make([]*quadstate.Entity, len(s.ByType[t]))
		copy(clone.ByType[t], s.ByType[t])

	}

	return clone
}

func NewUpdate(size int) *Update {
	return &Update{
		RemovedIds: make([]entity.Id, 0, size),
		Removed:    make([]*quadstate.Entity, 0, size),
		Entities:   make([]*quadstate.Entity, 0, size),
	}
}

type EntityInverseBloom interface {
	AddId(id entity.Id)
	AddEntities(e []*quadstate.Entity)
	Exists(id entity.Id) bool
}

// TODO utilize the unknown map
func (u *Update) FromSnapshot(prev, next *Snapshot, prevBloom, nextBloom EntityInverseBloom, unknown map[entity.Id]struct{}) {
	u.Time = next.Time
	u.Bounds = next.Bounds

	u.TerrainMapSlices = nil

	u.Entities = append(u.Entities[:0], next.Entities.ByType[quadstate.TypeChanged]...)
	u.Entities = append(u.Entities, next.Entities.ByType[quadstate.TypeNew]...)
	u.Entities = append(u.Entities, next.Entities.ByType[quadstate.TypeInstant]...)
	u.Removed = append(u.Removed[:0], next.Entities.ByType[quadstate.TypeRemoved]...)
	u.RemovedIds = u.RemovedIds[:0]

	nextBloom.AddEntities(next.Entities.ByType[quadstate.TypeChanged])
	nextBloom.AddEntities(next.Entities.ByType[quadstate.TypeNew])
	for _, e := range next.Entities.ByType[quadstate.TypeUnchanged] {
		id := e.EntityId()
		nextBloom.AddId(id)
		if !prevBloom.Exists(id) {
			u.Entities = append(u.Entities, e)
		}
	}

	// TODO Remove once quad is updated with entity viewports
	for _, array := range [...][]*quadstate.Entity{
		prev.Entities.ByType[quadstate.TypeChanged],
		prev.Entities.ByType[quadstate.TypeNew],
		prev.Entities.ByType[quadstate.TypeUnchanged],
	} {
		for _, e := range array {
			id := e.EntityId()
			if !nextBloom.Exists(id) {
				u.RemovedIds = append(u.RemovedIds, id)
			}
		}
	}

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
