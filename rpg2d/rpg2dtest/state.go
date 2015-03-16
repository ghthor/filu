package rpg2dtest

import (
	"github.com/ghthor/engine/rpg2d"
	"github.com/ghthor/gospec"
)

type (
	worldState     rpg2d.WorldState
	worldStateDiff rpg2d.WorldStateDiff
)

type (
	terrainMapState       rpg2d.TerrainMapState
	terrainMapStateSlices []rpg2d.TerrainMapStateSlice
)

func StateEquals(actual interface{}, expected interface{}) (match bool, pos, neg gospec.Message, err error) {
	s1, ok1 := actual.(rpg2d.WorldState)
	s2, ok2 := expected.(rpg2d.WorldState)
	if ok1 && ok2 {
		return gospec.Equals(worldState(s1), worldState(s2))
	}

	d1, ok1 := actual.(rpg2d.WorldStateDiff)
	d2, ok2 := expected.(rpg2d.WorldStateDiff)

	if ok1 && ok2 {
		return gospec.Equals(worldStateDiff(d1), worldStateDiff(d2))
	}

	return gospec.Equals(actual, expected)

}

func (s worldState) Equals(other interface{}) bool {
	switch other := other.(type) {
	case worldState:
		return s.isEqual(other) && other.isEqual(s)
	default:
	}

	return false
}

func (s worldState) isEqual(other worldState) bool {
	switch {
	case s.Time != other.Time:
		return false
	case s.Bounds != other.Bounds:
		return false
	case len(s.Entities) != len(other.Entities):
		return false
	case !s.hasSameEntities(other):
		return false
	case !(*terrainMapState)(s.TerrainMap).isEqualTo((*terrainMapState)(other.TerrainMap)):
		return false
	}
	return true
}

func (s worldState) hasSameEntities(other worldState) bool {
nextEntity:
	for _, e1 := range s.Entities {
		for _, e2 := range other.Entities {
			if e1.EntityId() == e2.EntityId() {
				if e1.IsDifferentFrom(e2) {
					return false
				}

				continue nextEntity
			}
		}
		return false
	}

	return true
}

func (m *terrainMapState) isEqualTo(other *terrainMapState) bool {
	return m.Bounds == other.Bounds &&
		m.String() == other.String()
}

func (s worldStateDiff) Equals(other interface{}) bool {
	switch other := other.(type) {
	case worldStateDiff:
		return s.isEqual(other) && other.isEqual(s)

	default:
	}
	return false
}

func (s worldStateDiff) isEqual(other worldStateDiff) bool {
	switch {
	case s.Time != other.Time:
		return false
	case s.Bounds != other.Bounds:
		return false
	case len(s.Entities) != len(other.Entities) || len(s.Removed) != len(other.Removed):
		return false
	case !s.hasSameEntities(other):
		return false
	case !(terrainMapStateSlices)(s.TerrainMapSlices).isEqual((terrainMapStateSlices)(other.TerrainMapSlices)):
		return false

	default:
	}
	return true
}

func (s worldStateDiff) hasSameEntities(other worldStateDiff) bool {
nextEntity:
	for _, e1 := range s.Entities {
		for _, e2 := range other.Entities {
			if e1.EntityId() == e2.EntityId() {
				if e1.IsDifferentFrom(e2) {
					return false
				}

				continue nextEntity
			}
		}
		return false
	}

nextRemovedEntity:
	for _, e1 := range s.Removed {
		for _, e2 := range other.Removed {
			if e1.EntityId() == e2.EntityId() {
				if e1.IsDifferentFrom(e2) {
					return false
				}

				continue nextRemovedEntity
			}
		}
		return false
	}

	return true
}

func (s terrainMapStateSlices) isEqual(other terrainMapStateSlices) bool {
	if len(s) != len(other) {
		return false
	}

nextMap:
	for _, m1 := range s {
		for _, m2 := range other {
			if m1 == m2 {
				continue nextMap
			}
		}
		return false
	}

	return true
}
