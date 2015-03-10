package rpg2dtest

import (
	"github.com/ghthor/engine/rpg2d"
	"github.com/ghthor/gospec"
)

type worldState rpg2d.WorldState
type terrainMapState rpg2d.TerrainMapState

func StateEquals(actual interface{}, expected interface{}) (match bool, pos, neg gospec.Message, err error) {
	s1, ok1 := actual.(rpg2d.WorldState)
	s2, ok2 := expected.(rpg2d.WorldState)
	if !(ok1 && ok2) {
		return gospec.Equals(actual, expected)
	}

	return gospec.Equals(worldState(s1), worldState(s2))
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
			if e1.Id() == e2.Id() {
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
