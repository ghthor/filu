package quad

import (
	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
)

type CollisionId struct {
	AId, BId entity.Id
}

// A collision between 2 entities because the
// entities bounds are overlapping. Intended to
// be solved by the user defined NarrowPhaseHandler.
type Collision struct {
	CollisionId
	A, B entity.Entity
}

type CollisionById map[CollisionId]Collision

func NewCollision(a, b entity.Entity) Collision {
	aId, bId := a.Id(), b.Id()

	if aId > bId {
		return Collision{
			CollisionId{bId, aId},
			b, a,
		}
	}

	return Collision{
		CollisionId{aId, bId},
		a, b,
	}
}

// A collision index stores all the
// collisions an entity is involved in.
type CollisionIndex map[entity.Id][]Collision

// The bounds of A and B joined together.
func (c Collision) Bounds() coord.Bounds {
	return coord.JoinBounds(c.A.Bounds(), c.B.Bounds())
}

// Compares to Collisions and returns if
// they are representing the same collision.
func (c Collision) IsSameAs(oc Collision) bool {
	if c.AId == oc.AId {
		if c.BId == oc.BId {
			return true
		}
	}
	return false
}

// A group of collisions where each collision
// may have an effect on the others. A dependency
// tree should be created by the user to resolve
// the collisions in the correct order.
type CollisionGroup struct {
	Entities []entity.Entity
	CollisionIndex

	// A map of all the collisions in the group.
	CollisionsById map[CollisionId]Collision
}

func NewCollisionGroup(size int) *CollisionGroup {
	eSize := 2
	if size != 1 {
		eSize = size * 7 / 4
	}
	return &CollisionGroup{
		make([]entity.Entity, 0, eSize),
		make(CollisionIndex, eSize),
		make(map[CollisionId]Collision, size),
	}
}

func (cg *CollisionGroup) Reset() {
	cg.Entities = cg.Entities[:0]
	for k := range cg.CollisionIndex {
		delete(cg.CollisionIndex, k)
	}
	for k := range cg.CollisionsById {
		delete(cg.CollisionsById, k)
	}
}

func (cg CollisionGroup) Bounds() coord.Bounds {
	bounds := make([]coord.Bounds, 0, len(cg.CollisionsById))
	for _, c := range cg.CollisionsById {
		bounds = append(bounds, c.Bounds())
	}

	return coord.JoinBounds(bounds...)
}

// Adds a collision to the group. Also adds the
// entities from the collision to the entities slice.
// Filters out collisions it already has and entities
// that are already in the entities slice.
func (cg *CollisionGroup) AddCollision(a, b entity.Entity) {
	cg.addCollision(NewCollision(a, b))
}

func (cg *CollisionGroup) AddCollisionFromMerge(c Collision) {
	cg.addCollision(c)
}

func (cg *CollisionGroup) addCollision(c Collision) {
	id := c.CollisionId
	if _, exists := cg.CollisionsById[id]; exists {
		return
	}

	cg.CollisionsById[id] = c
	if _, exists := cg.CollisionIndex[c.AId]; !exists {
		cg.CollisionIndex[c.AId] = nil
		cg.Entities = append(cg.Entities, c.A)
	}
	if _, exists := cg.CollisionIndex[c.BId]; !exists {
		cg.CollisionIndex[c.BId] = nil
		cg.Entities = append(cg.Entities, c.B)
	}
	return
}

func (i CollisionIndex) addEntity(id entity.Id, c Collision, prealloc [][]Collision) [][]Collision {
	if collisions := i[id]; collisions != nil {
		i[id] = append(collisions, c)
		return prealloc
	}

	var collisions []Collision
	if len(prealloc) == 0 {
		collisions = make([]Collision, 0, 1)
	} else {
		collisions = prealloc[len(prealloc)-1][:0]
		prealloc = prealloc[:len(prealloc)-1]
	}
	i[id] = append(collisions, c)

	return prealloc
}

func (cg *CollisionGroup) FillIndex(prealloc [][]Collision) (CollisionIndex, [][]Collision) {
	for _, c := range cg.CollisionsById {
		prealloc = cg.CollisionIndex.addEntity(c.AId, c, prealloc)
		prealloc = cg.CollisionIndex.addEntity(c.BId, c, prealloc)
	}

	return cg.CollisionIndex, prealloc
}

// An entity may ONLY be assigned to 1 collision group.
// If an entity has collisions that are in separate collision
// groups, those groups must be merged. This rules make the
// collision group index possible.
type CollisionGroupIndex map[entity.Id]*CollisionGroup

type collisionGroupPair struct {
	entity.Entity
	*CollisionGroup
}

type CollisionGroupPairIndex map[entity.Id]collisionGroupPair

type CollisionGroupPool struct {
	free []*CollisionGroup
	used []*CollisionGroup
}

func (pool *CollisionGroupPool) NewGroup() *CollisionGroup {
	if len(pool.free) == 0 {
		cg := NewCollisionGroup(1)
		pool.used = append(pool.used, cg)
		return cg
	}

	i := len(pool.free) - 1
	cg := pool.free[i]
	pool.free = pool.free[:i]
	return cg
}

func (pool *CollisionGroupPool) Reset() {
	for _, cg := range pool.used {
		cg.Reset()
		pool.free = append(pool.free, cg)
	}

	pool.used = pool.used[:0]
}
