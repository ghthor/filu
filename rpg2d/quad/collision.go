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
type CollisionIndex map[entity.Entity][]Collision

func (i CollisionIndex) add(collisions map[CollisionId]Collision) CollisionIndex {
	for _, c := range collisions {
		a, b := c.A, c.B
		i[a] = append(i[a], c)
		i[b] = append(i[b], c)
	}
	return i
}

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
	// A slice of the all the entities that are in
	// the collisions of the group.
	Entities []entity.Entity

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
		make(map[CollisionId]Collision, size),
	}
}

func (cg CollisionGroup) Bounds() coord.Bounds {
	bounds := make([]coord.Bounds, 0, len(cg.CollisionsById))
	for _, c := range cg.CollisionsById {
		bounds = append(bounds, c.Bounds())
	}

	return coord.JoinBounds(bounds...)
}

func (cg CollisionGroup) CollisionIndex() CollisionIndex {
	index := make(CollisionIndex, len(cg.Entities))
	return index.add(cg.CollisionsById)
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

	a, b := c.A, c.B

	for _, e := range cg.Entities {
		if a == e {
			goto check_B_Exists
		}
	}
	cg.Entities = append(cg.Entities, a)

check_B_Exists:
	for _, e := range cg.Entities {
		if b == e {
			return
		}
	}
	cg.Entities = append(cg.Entities, b)

	return
}

// An entity may ONLY be assigned to 1 collision group.
// If an entity has collisions that are in separate collision
// groups, those groups must be merged. This rules make the
// collision group index possible.
type CollisionGroupIndex map[entity.Entity]*CollisionGroup
