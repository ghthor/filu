package worldstate

import (
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/quadstate"
)

var _ EntityInverseBloom = &InverseBloom{}

type InverseBloom struct {
	hashs   [][]entity.Id
	size    int
	indexed []entity.Id
}

func NewInverseBloom(size int) *InverseBloom {
	return &InverseBloom{
		make([][]entity.Id, size),
		size,
		make([]entity.Id, size),
	}
}

func (s *InverseBloom) AddId(id entity.Id) {
	index := id % entity.Id(s.size)
	s.hashs[index] = append(s.hashs[index], id)
	s.indexed = append(s.indexed, index)
}

func (s *InverseBloom) AddEntities(e []*quadstate.Entity) {
	for _, e := range e {
		id := e.EntityId()
		index := id % entity.Id(s.size)
		s.hashs[index] = append(s.hashs[index], id)
		s.indexed = append(s.indexed, index)
	}
}

func (s *InverseBloom) Add(ids []entity.Id) {
	for _, id := range ids {
		index := id % entity.Id(s.size)
		s.hashs[index] = append(s.hashs[index], id)
		s.indexed = append(s.indexed, index)
	}
}

func (s InverseBloom) Exists(id entity.Id) bool {
	index := id % entity.Id(s.size)
	set := s.hashs[index]
	for _, other := range set {
		if id == other {
			return true
		}
	}

	return false
}

func (s *InverseBloom) Reset() {
	// TODO Optimize Resets by keeping a cache of which indexs have been modified
	for _, index := range s.indexed {
		s.hashs[index] = s.hashs[index][:0]
	}

	s.indexed = s.indexed[:0]
}

func (s InverseBloom) PreallocBuckets(size int) {
	for i, hash := range s.hashs {
		if cap(hash) < size {
			s.hashs[i] = make([]entity.Id, 0, size)
		}
	}
}

var _ EntityInverseBloom = InverseBloomMap{}

type InverseBloomMap struct {
	m map[entity.Id]struct{}
}

func NewInverseBloomMap(size int) InverseBloomMap {
	return InverseBloomMap{
		make(map[entity.Id]struct{}),
	}
}

func (s InverseBloomMap) AddId(id entity.Id) {
	s.m[id] = struct{}{}
}

func (s InverseBloomMap) AddEntities(e []*quadstate.Entity) {
	for _, e := range e {
		id := e.EntityId()
		s.m[id] = struct{}{}
	}
}

func (s InverseBloomMap) Add(ids []entity.Id) {
	for _, id := range ids {
		s.m[id] = struct{}{}
	}
}

func (s InverseBloomMap) Exists(id entity.Id) bool {
	_, exists := s.m[id]
	return exists
}

func (s InverseBloomMap) Reset() {
	for k := range s.m {
		delete(s.m, k)
	}
}
