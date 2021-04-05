package worldstate_test

import (
	"math/rand"
	"testing"

	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/worldstate"
)

func TestBloomFilter(t *testing.T) {
	filter := worldstate.NewInverseBloom(10000)
	filter.PreallocBuckets(3)

	r := rand.New(rand.NewSource(0))
	ids := make([]entity.Id, idMax)
	idsMap := make(map[entity.Id]struct{}, idMax)
	for i := 0; i < idMax; i++ {
		id := entity.Id(r.Intn(idMax))
		ids[i] = id
		idsMap[id] = struct{}{}
	}

	idChks := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		idChks[i] = entity.Id(r.Intn(idMax))
	}

	filter.Add(ids)

	for _, id := range ids {
		if !filter.Exists(id) {
			t.Errorf("expected %v to exist", id)
			t.Fatal()
		}
	}

	for _, id := range idChks {
		if _, exists := idsMap[id]; exists {
			if !filter.Exists(id) {
				t.Errorf("expected %v to exist", id)
				t.Fatal()
			}
		} else {
			if filter.Exists(id) {
				t.Errorf("expected %v not to exist", id)
				t.Fatal()
			}
		}
	}

	filter.Reset()

	for _, id := range ids {
		if filter.Exists(id) {
			t.Errorf("expected %v to not exist", id)
			t.Fatal()
		}
	}

}

const idMax = 3000

func BenchmarkMapFilter(b *testing.B) {
	filters := make([]map[entity.Id]struct{}, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = make(map[entity.Id]struct{})
	}

	r := rand.New(rand.NewSource(0))
	ids := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		ids[i] = entity.Id(r.Intn(idMax))
	}

	idChks := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		idChks[i] = entity.Id(r.Intn(idMax))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter := filters[i]
		for _, id := range ids {
			filter[id] = struct{}{}
		}

		for _, id := range idChks {
			_, _ = filter[id]
		}

		for k := range filter {
			delete(filter, k)
		}
	}
}

func BenchmarkMapFilterPrealloc(b *testing.B) {
	filters := make([]map[entity.Id]struct{}, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = make(map[entity.Id]struct{}, idMax)
	}

	r := rand.New(rand.NewSource(0))
	ids := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		ids[i] = entity.Id(r.Intn(idMax))
	}

	idChks := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		idChks[i] = entity.Id(r.Intn(idMax))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter := filters[i]
		for _, id := range ids {
			filter[id] = struct{}{}
		}

		for _, id := range idChks {
			_, _ = filter[id]
		}

		for k := range filter {
			delete(filter, k)
		}
	}
}

func BenchmarkBloomFilter100(b *testing.B) {
	const size = 100
	const prealloc = 1
	filters := make([]*worldstate.InverseBloom, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = worldstate.NewInverseBloom(size)
		filters[i].PreallocBuckets(prealloc)
	}

	r := rand.New(rand.NewSource(0))
	ids := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		ids[i] = entity.Id(r.Intn(idMax))
	}

	idChks := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		idChks[i] = entity.Id(r.Intn(idMax))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter := filters[i]
		filter.Add(ids)

		for _, id := range idChks {
			filter.Exists(id)
		}

		filter.Reset()
	}
}

func BenchmarkBloomFilter1000(b *testing.B) {
	const size = 1000
	const prealloc = 3
	filters := make([]*worldstate.InverseBloom, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = worldstate.NewInverseBloom(size)
		filters[i].PreallocBuckets(prealloc)
	}

	r := rand.New(rand.NewSource(0))
	ids := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		ids[i] = entity.Id(r.Intn(idMax))
	}

	idChks := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		idChks[i] = entity.Id(r.Intn(idMax))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter := filters[i]
		filter.Add(ids)

		for _, id := range idChks {
			filter.Exists(id)
		}

		filter.Reset()
	}
}

func BenchmarkBloomFilter10000(b *testing.B) {
	const size = 10000
	const prealloc = 3

	r := rand.New(rand.NewSource(0))
	ids := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		ids[i] = entity.Id(r.Intn(idMax))
	}

	idChks := make([]entity.Id, idMax)
	for i := 0; i < idMax; i++ {
		idChks[i] = entity.Id(r.Intn(idMax))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		filter := worldstate.NewInverseBloom(size)
		filter.PreallocBuckets(prealloc)

		b.StartTimer()
		filter.Add(ids)

		for _, id := range idChks {
			filter.Exists(id)
		}

		filter.Reset()
	}
}
