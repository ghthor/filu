package rpg2d

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/sim/stime"
)

// Used to calculate diff's
type TerrainMapState struct {
	TerrainMap
}

func (m TerrainMapState) MarshalJSON() ([]byte, error) {
	return json.Marshal(TerrainMapStateSlice{
		Bounds:  m.TerrainMap.Bounds,
		Terrain: m.TerrainMap.String(),
	})
}

func (m TerrainMapState) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := gob.NewEncoder(buf)
	err := enc.Encode(TerrainMapStateSlice{
		Bounds:  m.TerrainMap.Bounds,
		Terrain: m.TerrainMap.String(),
	})

	return buf.Bytes(), err
}

func (m *TerrainMapState) UnmarshalBinary(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	slice := TerrainMapStateSlice{}
	err := dec.Decode(&slice)
	if err != nil {
		return err
	}

	m.TerrainMap, err = NewTerrainMap(slice.Bounds, slice.Terrain)
	return err
}

type TerrainMapStateSlice struct {
	Bounds  coord.Bounds `json:"bounds"`
	Terrain string       `json:"terrain"`
}

type TerrainMapStateDiff struct {
	Bounds  coord.Bounds        `json:"bounds"`
	Changes []TerrainTypeChange `json:"changes"`
}

func (m *TerrainMapState) IsEmpty() bool {
	if m == nil {
		return true
	}
	return m.TerrainMap.TerrainTypes == nil
}

func (m *TerrainMapState) Diff(other *TerrainMapState) []TerrainMapStateSlice {
	if m.IsEmpty() || !m.Bounds.Overlaps(other.Bounds) {
		return []TerrainMapStateSlice{{
			Bounds:  other.Bounds,
			Terrain: other.String(),
		}}
	}

	mBounds, oBounds := m.TerrainMap.Bounds, other.TerrainMap.Bounds
	rects := mBounds.DiffFrom(oBounds)

	// mBounds == oBounds
	if len(rects) == 0 {
		// TODO Still need to calc changes to map types in cells
		return nil
	}

	slices := make([]TerrainMapStateSlice, 0, len(rects))
	for _, r := range rects {
		slice := other.Slice(r)
		slices = append(slices, TerrainMapStateSlice{
			Bounds:  r,
			Terrain: slice.String(),
		})
	}

	return slices
}

func (m *TerrainMapState) Clone() (*TerrainMapState, error) {
	if m == nil {
		return m, nil
	}

	tm, err := m.TerrainMap.Clone()
	if err != nil {
		return nil, err
	}

	return &TerrainMapState{TerrainMap: tm}, nil
}

func (m TerrainMapStateSlice) IsEmpty() bool {
	return m.Bounds == coord.Bounds{}
}

type WorldState struct {
	Time   stime.Time   `json:"time"`
	Bounds coord.Bounds `json:"bounds"`

	Entities entity.StateSlice `json:"entities"`

	TerrainMap *TerrainMapState `json:"terrainMap,omitempty"`
}

type WorldStateDiff struct {
	Time   stime.Time   `json:"time"`
	Bounds coord.Bounds `json:"bounds"`

	Entities entity.StateSlice `json:"entities"`
	Removed  entity.StateSlice `json:"removed"`

	TerrainMapSlices []TerrainMapStateSlice `json:"terrainMapSlices,omitempty"`
}

func (s WorldState) Clone() WorldState {
	terrainMap, err := s.TerrainMap.Clone()
	if err != nil {
		panic("error cloning terrain map: " + err.Error())
	}
	clone := WorldState{
		Time:       s.Time,
		Bounds:     s.Bounds,
		Entities:   make(entity.StateSlice, len(s.Entities)),
		TerrainMap: terrainMap,
	}
	copy(clone.Entities, s.Entities)
	return clone
}

// Returns a world state that only contains
// entities and terrain within bounds.
// Does NOT change world state type.
func (s WorldState) Cull(bounds coord.Bounds) (other WorldState) {
	other.Entities = make(entity.StateSlice, 0, len(s.Entities))
	return s.CullInto(other, bounds)
}

func (s WorldState) CullInto(other WorldState, bounds coord.Bounds) WorldState {
	other.Time = s.Time
	other.Bounds = bounds

	other.Entities = other.Entities[:0]
	other.Entities = s.Entities.FilterByBounds(other.Entities, bounds)

	// Cull Terrain
	// TODO Maybe remove the ability to have an empty TerrainMap
	// Requires updating some tests to have a terrain map that don't have one
	if !s.TerrainMap.IsEmpty() {
		other.TerrainMap = &TerrainMapState{TerrainMap: s.TerrainMap.Slice(bounds)}
	} else {
		other.TerrainMap = nil
	}

	return other
}

// Returns a world state that only contains
// entities and terrain that is different such
// that state + diff == other. Diff is therefor
// the changes necessary to get from state to other.
func (prev WorldState) Diff(next WorldState) (diff WorldStateDiff) {
	diff.Entities = make(entity.StateSlice, 0, len(next.Entities))
	diff.Removed = make(entity.StateSlice, 0, len(next.Entities))
	diff.Between(prev, next)
	return diff
}

// TODO Figure out a way to reuse the maps
func (diff *WorldStateDiff) Between(prev, next WorldState) {
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

		if old.IsDifferentFrom(e) {
			diff.Entities = append(diff.Entities, e)
			goto cleanup
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
			tm, err := NewTerrainMap(slice.Bounds, slice.Terrain)
			if err != nil {
				panic(fmt.Sprintf("error applying diff: %v", err))
			}
			state.TerrainMap.TerrainMap = tm
		}
	case 0:
	}

	state.Time = diff.Time
	state.Bounds = diff.Bounds
}
