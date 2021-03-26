package rpg2dtest

import (
	"github.com/ghthor/filu/rpg2d"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/worldterrain"
	"github.com/ghthor/gospec"
)

type (
	worldState     rpg2d.WorldState
	worldStateDiff rpg2d.WorldStateDiff
)

type (
	terrainMapState       worldterrain.MapState
	terrainMapStateSlices worldterrain.MapStateSlices
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

func entitiesAreEqual(a, b entity.StateSlice) bool {
	matched := make(map[entity.Id]entity.State, len(a))

nextEntity:
	for _, e1 := range a {
		for _, e2 := range b {
			// Skip checking against b entities that we've already matched
			if _, matched := matched[e2.EntityId()]; matched {
				continue
			}

			// Skip deep compare if Id's don't match
			if e1.EntityId() != e2.EntityId() {
				continue
			}

			if ee1, canChange := e1.(entity.StateCanChange); canChange {
				if ee1.IsDifferentFrom(e2) {
					return false
				}
			}

			// Cached that we've verified Equality for the EntityID
			matched[e1.EntityId()] = e1
			continue nextEntity
		}

		// We didn't find an instance of e1.EntityId() in slice b
		return false
	}

	return true
}

func (s worldState) hasSameEntities(other worldState) bool {
	return entitiesAreEqual(s.Entities, other.Entities)
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
	case !(*terrainMapStateSlices)(s.TerrainMapSlices).isEqual((*terrainMapStateSlices)(other.TerrainMapSlices)):
		return false

	default:
	}
	return true
}

func (s worldStateDiff) hasSameEntities(other worldStateDiff) bool {
	return entitiesAreEqual(s.Entities, other.Entities) &&
		entitiesAreEqual(s.Removed, other.Removed)
}

func (s *terrainMapStateSlices) isEqual(other *terrainMapStateSlices) bool {
	if s.Bounds != other.Bounds {
		return false
	}

	if len(s.Slices) != len(other.Slices) {
		return false
	}

nextMap:
	for _, m1 := range s.Slices {
		for _, m2 := range other.Slices {
			if m1 == m2 {
				continue nextMap
			}
		}
		return false
	}

	return true
}
