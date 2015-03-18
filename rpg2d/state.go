package rpg2d

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/sim/stime"
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
		Entities:   make(entity.StateSlice, len(s.Entities)),
		TerrainMap: terrainMap,
	}
	copy(clone.Entities, s.Entities)
	return clone
}

// Returns a world state that only contains
// entities and terrain within bounds.
// Does NOT change world state type.
func (s WorldState) Cull(bounds coord.Bounds) (culled WorldState) {
	culled.Time = s.Time
	culled.Bounds = bounds

	// Cull Entities
	for _, e := range s.Entities {
		if bounds.Overlaps(e.Bounds()) {
			culled.Entities = append(culled.Entities, e)
		}
	}

	// Cull Terrain
	// TODO Maybe remove the ability to have an empty TerrainMap
	// Requires updating some tests to have a terrain map that don't have one
	if !s.TerrainMap.IsEmpty() {
		culled.TerrainMap = &TerrainMapState{TerrainMap: s.TerrainMap.Slice(bounds)}
	}
	return
}

// Returns a world state that only contains
// entities and terrain that is different such
// that state + diff == other. Diff is therefor
// the changes necessary to get from state to other.
func (state WorldState) Diff(other WorldState) (diff WorldStateDiff) {
	diff.Time = other.Time
	diff.Bounds = other.Bounds

	if len(state.Entities) == 0 && len(other.Entities) > 0 {
		diff.Entities = other.Entities
	} else {
		// Find the entities that have changed from the old state to the new one
	nextEntity:
		for _, entity := range other.Entities {
			for _, old := range state.Entities {
				if entity.EntityId() == old.EntityId() {
					if old.IsDifferentFrom(entity) {
						diff.Entities = append(diff.Entities, entity)
					}
					continue nextEntity
				}
			}
			// This is a new Entity
			diff.Entities = append(diff.Entities, entity)
		}

		// Check if all the entities in old state exist in the new state
	entityStillExists:
		for _, old := range state.Entities {
			for _, entity := range other.Entities {
				if old.EntityId() == entity.EntityId() {
					continue entityStillExists
				}
			}
			diff.Removed = append(diff.Removed, old)
		}
	}

	// Diff the TerrainMap
	diff.TerrainMapSlices = state.TerrainMap.Diff(other.TerrainMap)
	return
}

// Modifies the world state with the
// changes in a world state diff.
func (state *WorldState) Apply(diff WorldStateDiff) {
	state.Time = diff.Time

nextRemoved:
	for _, removed := range diff.Removed {
		for i, e := range state.Entities {
			if e.EntityId() == removed.EntityId() {
				state.Entities = append(state.Entities[:i], state.Entities[i+1:]...)
				break nextRemoved
			}
		}
	}

	for _, added := range diff.Entities {
		state.Entities = append(state.Entities, added)
	}
}
