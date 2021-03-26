package rpg2d

import (
	"fmt"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/worldterrain"
	"github.com/ghthor/filu/sim/stime"
)

type WorldState struct {
	Time   stime.Time   `json:"time"`
	Bounds coord.Bounds `json:"bounds"`

	Entities          entity.StateSlice `json:"entities"`
	EntitiesRemoved   entity.StateSlice `json:"entitiesRemoved"`
	EntitiesNew       entity.StateSlice `json:"entitiesNew"`
	EntitiesChanged   entity.StateSlice `json:"entitiesChanged"`
	EntitiesUnchanged entity.StateSlice `json:"entitiesUnchanged"`

	TerrainMap *worldterrain.MapState `json:"terrainMap,omitempty"`
}

type WorldStateDiff struct {
	Time   stime.Time   `json:"time"`
	Bounds coord.Bounds `json:"bounds"`

	Entities entity.StateSlice `json:"entities"`
	Removed  entity.StateSlice `json:"removed"`

	TerrainMapSlices []worldterrain.MapStateSlice `json:"terrainMapSlices,omitempty"`
}

func (s WorldState) Clone() WorldState {
	terrainMap, err := s.TerrainMap.Clone()
	if err != nil {
		panic("error cloning terrain map: " + err.Error())
	}
	clone := WorldState{
		Time:              s.Time,
		Bounds:            s.Bounds,
		Entities:          make(entity.StateSlice, len(s.Entities)),
		EntitiesRemoved:   make(entity.StateSlice, len(s.EntitiesRemoved)),
		EntitiesNew:       make(entity.StateSlice, len(s.EntitiesNew)),
		EntitiesChanged:   make(entity.StateSlice, len(s.EntitiesChanged)),
		EntitiesUnchanged: make(entity.StateSlice, len(s.EntitiesUnchanged)),
		TerrainMap:        terrainMap,
	}
	copy(clone.Entities, s.Entities)
	copy(clone.EntitiesRemoved, s.EntitiesRemoved)
	copy(clone.EntitiesNew, s.EntitiesNew)
	copy(clone.EntitiesChanged, s.EntitiesChanged)
	copy(clone.EntitiesUnchanged, s.EntitiesUnchanged)
	return clone
}

// Returns a world state that only contains
// entities and terrain within bounds.
// Does NOT change world state type.
func (s WorldState) Cull(bounds coord.Bounds) (other WorldState) {
	other.Entities = make(entity.StateSlice, 0, len(s.Entities))
	return s.CullInto(other, bounds)
}

func (s WorldState) CullForInitialState(bounds coord.Bounds) (result WorldState) {
	result = WorldState{
		Time:     s.Time,
		Bounds:   bounds,
		Entities: s.Entities.FilterByBounds(make(entity.StateSlice, 0, len(s.Entities)), bounds),
	}

	// Cull Terrain
	// TODO Maybe remove the ability to have an empty TerrainMap
	// Requires updating some tests to have a terrain map that don't have one
	if !s.TerrainMap.IsEmpty() {
		result.TerrainMap = &worldterrain.MapState{Map: s.TerrainMap.Slice(bounds)}
	} else {
		result.TerrainMap = nil
	}

	return result
}

func (s WorldState) CullInto(other WorldState, bounds coord.Bounds) (result WorldState) {
	result = WorldState{
		Time:   s.Time,
		Bounds: bounds,

		Entities:          s.Entities.FilterByBounds(other.Entities[:0], bounds),
		EntitiesRemoved:   s.EntitiesRemoved.FilterByBounds(other.EntitiesRemoved[:0], bounds),
		EntitiesNew:       s.EntitiesNew.FilterByBounds(other.EntitiesNew[:0], bounds),
		EntitiesChanged:   s.EntitiesChanged.FilterByBounds(other.EntitiesChanged[:0], bounds),
		EntitiesUnchanged: s.EntitiesUnchanged.FilterByBounds(other.EntitiesUnchanged[:0], bounds),
	}

	// Cull Terrain
	// TODO Maybe remove the ability to have an empty TerrainMap
	// Requires updating some tests to have a terrain map that don't have one
	if !s.TerrainMap.IsEmpty() {
		result.TerrainMap = &worldterrain.MapState{Map: s.TerrainMap.Slice(bounds)}
	} else {
		result.TerrainMap = nil
	}

	return result
}

// Returns a world state that only contains
// entities and terrain that is different such
// that state + diff == other. Diff is therefor
// the changes necessary to get from state to other.
func (prev WorldState) Diff(next WorldState) (diff WorldStateDiff) {
	diff.Entities = make(entity.StateSlice, 0, len(next.Entities))
	diff.Removed = make(entity.StateSlice, 0, len(next.Entities))
	return prev.diffExpensive(next, diff)
}

func (prev WorldState) diffExpensive(next WorldState, diff WorldStateDiff) WorldStateDiff {
	diff.Time = next.Time
	diff.Bounds = next.Bounds
	diff.Entities = diff.Entities[:0]
	diff.Removed = diff.Removed[:0]
	diff.TerrainMapSlices = nil

	newById := make(entity.StateById, len(next.Entities))
	for _, entity := range next.Entities {
		newById[entity.EntityId()] = entity
	}

	// Check if all the entities in old state exist in the new state
	for _, old := range prev.Entities {
		e, exists := newById[old.EntityId()]
		if !exists {
			diff.Removed = append(diff.Removed, old)
			continue
		}

		if old, canChange := old.(entity.StateCanChange); canChange {
			if old.IsDifferentFrom(e) {
				diff.Entities = append(diff.Entities, e)
				goto cleanup
			}
		}

	cleanup:
		delete(newById, old.EntityId())
	}

	for _, entity := range newById {
		// This is a new Entity
		diff.Entities = append(diff.Entities, entity)
	}

	// Diff the TerrainMap
	diff.TerrainMapSlices = prev.TerrainMap.Diff(next.TerrainMap)
	return diff
}

// TODO Figure out a way to reuse the maps
func (diff *WorldStateDiff) Between(prev, next WorldState) {
	diff.Time = next.Time
	diff.Bounds = next.Bounds

	diff.TerrainMapSlices = nil

	diff.Entities = append(diff.Entities[:0], next.EntitiesChanged...)
	diff.Entities = append(diff.Entities, next.EntitiesNew...)
	diff.Removed = append(diff.Removed[:0], next.EntitiesRemoved...)

	// Diff the TerrainMap
	diff.TerrainMapSlices = prev.TerrainMap.Diff(next.TerrainMap)
	return
}

// Modifies the world state with the
// changes in a world state diff.
func (state *WorldState) Apply(diff WorldStateDiff) {

nextRemoved:
	for _, removed := range diff.Removed {
		for i, e := range state.Entities {
			if e.EntityId() == removed.EntityId() {
				state.Entities = append(state.Entities[:i], state.Entities[i+1:]...)
				continue nextRemoved
			}
		}
	}

nextAddedOrModified:
	for _, e := range diff.Entities {
		for i, old := range state.Entities {
			if old.EntityId() == e.EntityId() {
				state.Entities[i] = e
				continue nextAddedOrModified
			}
		}

		state.Entities = append(state.Entities, e)
	}

	switch len(diff.TerrainMapSlices) {
	default:
		state.TerrainMap.MergeDiff(diff.Bounds, diff.TerrainMapSlices...)
	case 1:
		if state.Bounds.Overlaps(diff.Bounds) {
			state.TerrainMap.MergeDiff(diff.Bounds, diff.TerrainMapSlices...)
		} else {
			slice := diff.TerrainMapSlices[0]
			tm, err := worldterrain.NewMap(slice.Bounds, slice.Terrain)
			if err != nil {
				panic(fmt.Sprintf("error applying diff: %v", err))
			}
			state.TerrainMap.Map = tm
		}
	case 0:
	}

	state.Time = diff.Time
	state.Bounds = diff.Bounds
}

func clear(s entity.StateSlice) entity.StateSlice {
	if s == nil {
		return make(entity.StateSlice, 0, 1)
	}

	return s[:0]
}

func (state WorldState) Clear() WorldState {
	return WorldState{
		Time:              state.Time,
		Bounds:            state.Bounds,
		Entities:          clear(state.Entities),
		EntitiesRemoved:   clear(state.EntitiesRemoved),
		EntitiesNew:       clear(state.EntitiesNew),
		EntitiesChanged:   clear(state.EntitiesChanged),
		EntitiesUnchanged: clear(state.EntitiesUnchanged),
	}
}

// TODO Fix the need to convert from quadstate.Entity to entity.State
var _ quadstate.Accumulator = &WorldState{}

func (state *WorldState) Add(e quadstate.Entity, flag entity.Flag) {
	switch {
	case flag&entity.FlagRemoved != 0:
		state.EntitiesRemoved = append(state.EntitiesRemoved, e.State)
	case flag&entity.FlagNew != 0:
		state.EntitiesNew = append(state.EntitiesNew, e.State)
	case flag&entity.FlagChanged != 0:
		state.EntitiesChanged = append(state.EntitiesChanged, e.State)
	default:
		state.EntitiesUnchanged = append(state.EntitiesUnchanged, e.State)
	}
}

func (state *WorldState) AddSlice(entities []quadstate.Entity, flag entity.Flag) {
	for _, e := range entities {
		state.Add(e, flag)
	}
}
