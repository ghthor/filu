package rpg2d

import (
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/sim/stime"
)

type TerrainMapState struct {
	// Used to calculate diff's
	TerrainMap `json:"-"`

	// A Slice of new terrain the client doesn't have
	Bounds  *coord.Bounds `json:"bounds,omitempty"`
	Terrain string        `json:"terrain,omitempty"`

	// An array of type changes
	Changes []TerrainTypeChange `json:"changes,omitempty"`
}

// Prepare to be Marshalled
func (m *TerrainMapState) Prepare() {
	// Set the bounds
	m.Bounds = &m.TerrainMap.Bounds
	// Write out the Map as a string
	m.Terrain = m.TerrainMap.String()
}

func (m *TerrainMapState) IsEmpty() bool {
	if m == nil {
		return true
	}
	return m.TerrainMap.TerrainTypes == nil
}

func (m *TerrainMapState) Diff(other *TerrainMapState) (diff *TerrainMapState) {
	if m.IsEmpty() {
		return other
	}

	maabb, oaabb := m.TerrainMap.Bounds, other.TerrainMap.Bounds
	if maabb == oaabb {
		// No Overlaps
	} else {

		// Find the non overlapped section and set that in the diff
		switch {
		// Overlap top or bottom
		case maabb.Width() == oaabb.Width() &&
			maabb.TopL.X == oaabb.TopL.X &&
			maabb.BotR.X == oaabb.BotR.X:

			if maabb.Height() != oaabb.Height() {
				panic("invalid diff attempt")
			}

			// Overlaps the top
			if oaabb.TopL.Y > maabb.TopL.Y {
				diff = &TerrainMapState{TerrainMap: other.Slice(coord.Bounds{
					oaabb.TopL,
					coord.Cell{oaabb.BotR.X, maabb.TopL.Y + 1},
				})}

			} else if oaabb.BotR.Y < maabb.BotR.Y {
				// Overlaps the bottom
				diff = &TerrainMapState{TerrainMap: other.Slice(coord.Bounds{
					coord.Cell{oaabb.TopL.X, maabb.BotR.Y - 1},
					oaabb.BotR,
				})}
			} else {
				panic("invalid diff attempt")
			}

			// Overlaps left of right
		case maabb.Height() == oaabb.Height() &&
			maabb.TopL.Y == oaabb.TopL.Y &&
			maabb.BotR.Y == oaabb.BotR.Y:

			if maabb.Width() != oaabb.Width() {
				panic("invalid diff attempt")
			}

			// Overlaps the left
			if oaabb.TopL.X < maabb.TopL.X {
				diff = &TerrainMapState{TerrainMap: other.Slice(coord.Bounds{
					oaabb.TopL,
					coord.Cell{maabb.TopL.X - 1, oaabb.BotR.Y},
				})}
			} else if oaabb.BotR.X > maabb.BotR.X {
				// Overlaps the right
				diff = &TerrainMapState{TerrainMap: other.Slice(coord.Bounds{
					coord.Cell{maabb.BotR.X + 1, oaabb.TopL.Y},
					oaabb.BotR,
				})}
			} else {
				panic("invalid diff attempt")
			}

		default:
			panic("invalid diff attempt")
		}
	}
	return
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

type WorldState struct {
	Time       stime.Time       `json:"time"`
	Entities   []entity.State   `json:"entities"`
	Removed    []entity.State   `json:"removed"`
	TerrainMap *TerrainMapState `json:"terrainMap,omitempty"`
}

func (s WorldState) Clone() WorldState {
	terrainMap, err := s.TerrainMap.Clone()
	if err != nil {
		panic("error cloning terrain map: " + err.Error())
	}
	clone := WorldState{
		s.Time,
		make([]entity.State, len(s.Entities)),
		nil,
		terrainMap,
	}
	copy(clone.Entities, s.Entities)
	return clone
}

func (s WorldState) Cull(bounds coord.Bounds) (culled WorldState) {
	culled.Time = s.Time

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

func (s WorldState) Diff(ss WorldState) (diff WorldState) {
	diff.Time = ss.Time

	if len(s.Entities) == 0 && len(ss.Entities) > 0 {
		diff.Entities = ss.Entities
	} else {
		// Find the entities that have changed from the old state to the new one
	nextEntity:
		for _, entity := range ss.Entities {
			for _, old := range s.Entities {
				if entity.Id() == old.Id() {
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
		for _, old := range s.Entities {
			for _, entity := range ss.Entities {
				if old.Id() == entity.Id() {
					continue entityStillExists
				}
			}
			diff.Removed = append(diff.Removed, old)
		}
	}

	// Diff the TerrainMap
	diff.TerrainMap = s.TerrainMap.Diff(ss.TerrainMap)
	return
}

// TerrainMap needs an extra step before sending
// TODO remove this maybe?
// The extra step is to avoid casting the entire terrain map to a string
// when the world state json is created. The Diff function could run this step
// and we could call it "Finalize"
func (s WorldState) Prepare() {
	if !s.TerrainMap.IsEmpty() {
		s.TerrainMap.Prepare()
	}
}
