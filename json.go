package engine

import (
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/time"
)

type EntityJson interface {
	Id() EntityId
	AABB() coord.Bounds
	IsDifferentFrom(EntityJson) bool
}

type PlayerJson struct {
	// TODO When go updates to 1.1? this can be converted to an embedded type
	// The current json marshaller doesn't marshall embedded types
	EntityId    EntityId               `json:"id"`
	Name        string                 `json:"name"`
	Facing      string                 `json:"facing"`
	PathActions []coord.PathActionJson `json:"pathActions"`
	Cell        coord.Cell             `json:"cell"`
}

type TerrainMapJson struct {
	// Used to calculate diff's
	TerrainMap `json:"-"`

	// A Slice of new terrain the client doesn't have
	Bounds  *coord.Bounds `json:"bounds,omitempty"`
	Terrain string        `json:"terrain,omitempty"`

	// An array of type changes
	Changes []TerrainTypeChange `json:"changes,omitempty"`
}

// External format used to send state to the clients
type WorldStateJson struct {
	Time       time.Time       `json:"time"`
	Entities   []EntityJson    `json:"entities"`
	Removed    []EntityJson    `json:"removed"`
	TerrainMap *TerrainMapJson `json:"terrainMap,omitempty"`
}

func (p PlayerJson) Id() EntityId       { return p.EntityId }
func (p PlayerJson) AABB() coord.Bounds { return coord.Bounds{p.Cell, p.Cell} }
func (p PlayerJson) IsDifferentFrom(other EntityJson) (different bool) {
	o := other.(PlayerJson)

	switch {
	case p.Facing != o.Facing:
		fallthrough
	case len(p.PathActions) != len(o.PathActions):
		different = true
	case len(p.PathActions) == len(o.PathActions):
		for i, _ := range o.PathActions {
			different = different || (p.PathActions[i] != o.PathActions[i])
		}
	}
	return
}

// Prepare to be Marshalled
func (m *TerrainMapJson) Prepare() {
	// Set the bounds
	m.Bounds = &m.TerrainMap.Bounds
	// Write out the Map as a string
	m.Terrain = m.TerrainMap.String()
}

func (m *TerrainMapJson) IsEmpty() bool {
	if m == nil {
		return true
	}
	return m.TerrainMap.TerrainTypes == nil
}

func (m *TerrainMapJson) Diff(other *TerrainMapJson) (diff *TerrainMapJson) {
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
				diff = &TerrainMapJson{TerrainMap: other.Slice(coord.Bounds{
					oaabb.TopL,
					coord.Cell{oaabb.BotR.X, maabb.TopL.Y + 1},
				})}

			} else if oaabb.BotR.Y < maabb.BotR.Y {
				// Overlaps the bottom
				diff = &TerrainMapJson{TerrainMap: other.Slice(coord.Bounds{
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
				diff = &TerrainMapJson{TerrainMap: other.Slice(coord.Bounds{
					oaabb.TopL,
					coord.Cell{maabb.TopL.X - 1, oaabb.BotR.Y},
				})}
			} else if oaabb.BotR.X > maabb.BotR.X {
				// Overlaps the right
				diff = &TerrainMapJson{TerrainMap: other.Slice(coord.Bounds{
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

func (m *TerrainMapJson) Clone() (*TerrainMapJson, error) {
	if m == nil {
		return m, nil
	}

	tm, err := m.TerrainMap.Clone()
	if err != nil {
		return nil, err
	}

	return &TerrainMapJson{TerrainMap: tm}, nil
}

func (s WorldStateJson) Clone() WorldStateJson {
	terrainMap, err := s.TerrainMap.Clone()
	if err != nil {
		panic("error cloning terrain map: " + err.Error())
	}
	clone := WorldStateJson{
		s.Time,
		make([]EntityJson, len(s.Entities)),
		nil,
		terrainMap,
	}
	copy(clone.Entities, s.Entities)
	return clone
}

func (s WorldStateJson) Cull(aabb coord.Bounds) (culled WorldStateJson) {
	culled.Time = s.Time

	// Cull Entities
	for _, e := range s.Entities {
		if aabb.Overlaps(e.AABB()) {
			culled.Entities = append(culled.Entities, e)
		}
	}

	// Cull Terrain
	// TODO Maybe remove the ability to have an empty TerrainMap
	// Requires updating some tests to have a terrain map that don't have one
	if !s.TerrainMap.IsEmpty() {
		culled.TerrainMap = &TerrainMapJson{TerrainMap: s.TerrainMap.Slice(aabb)}
	}
	return
}

func (s WorldStateJson) Diff(ss WorldStateJson) (diff WorldStateJson) {
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
func (s WorldStateJson) Prepare() {
	if !s.TerrainMap.IsEmpty() {
		s.TerrainMap.Prepare()
	}
}
