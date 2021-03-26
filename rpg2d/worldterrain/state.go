package worldterrain

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/ghthor/filu/rpg2d/coord"
)

// Used to calculate diff's
type MapState struct {
	Map
}

func (m MapState) MarshalJSON() ([]byte, error) {
	return json.Marshal(MapStateSlice{
		Bounds:  m.Map.Bounds,
		Terrain: m.Map.String(),
	})
}

func (m MapState) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := gob.NewEncoder(buf)
	err := enc.Encode(MapStateSlice{
		Bounds:  m.Map.Bounds,
		Terrain: m.Map.String(),
	})

	return buf.Bytes(), err
}

func (m *MapState) UnmarshalBinary(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	slice := MapStateSlice{}
	err := dec.Decode(&slice)
	if err != nil {
		return err
	}

	m.Map, err = NewMap(slice.Bounds, slice.Terrain)
	return err
}

type MapStateSlice struct {
	Bounds  coord.Bounds `json:"bounds"`
	Terrain string       `json:"terrain"`
}

type MapStateDiff struct {
	Bounds  coord.Bounds        `json:"bounds"`
	Changes []TypeChange `json:"changes"`
}

func (m *MapState) IsEmpty() bool {
	if m == nil {
		return true
	}
	return m.Map.Types == nil
}

func (m *MapState) Diff(other *MapState) []MapStateSlice {
	if m.IsEmpty() || !m.Bounds.Overlaps(other.Bounds) {
		return []MapStateSlice{{
			Bounds:  other.Bounds,
			Terrain: other.String(),
		}}
	}

	mBounds, oBounds := m.Map.Bounds, other.Map.Bounds
	rects := mBounds.DiffFrom(oBounds)

	// mBounds == oBounds
	if len(rects) == 0 {
		// TODO Still need to calc changes to map types in cells
		return nil
	}

	slices := make([]MapStateSlice, 0, len(rects))
	for _, r := range rects {
		slice := other.Slice(r)
		slices = append(slices, MapStateSlice{
			Bounds:  r,
			Terrain: slice.String(),
		})
	}

	return slices
}

func (m *MapState) Clone() (*MapState, error) {
	if m == nil {
		return m, nil
	}

	tm, err := m.Map.Clone()
	if err != nil {
		return nil, err
	}

	return &MapState{Map: tm}, nil
}

func (m MapStateSlice) IsEmpty() bool {
	return m.Bounds == coord.Bounds{}
}
