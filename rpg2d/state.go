package rpg2d

import "github.com/ghthor/engine/rpg2d/coord"

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
