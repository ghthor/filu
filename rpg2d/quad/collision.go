package quad

import (
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
)

// A collision between 2 entities because the
// entities bounds are overlapping. Intended to
// be solved by the user defined NarrowPhaseHandler.
type Collision struct {
	A, B entity.Entity
}

// A collision index stores all the
// collisions an entity is involved in.
type CollisionIndex map[entity.Entity][]Collision

// The bounds of A and B joined together.
func (c Collision) Bounds() coord.Bounds {
	return coord.JoinBounds(c.A.Bounds(), c.B.Bounds())
}

// Compares to Collisions and returns if
// they are representing the same collision.
func (c Collision) IsSameAs(oc Collision) bool {
	switch {
	case c.A == oc.A && c.B == oc.B:
		fallthrough
	case c.A == oc.B && c.B == oc.A:
		return true
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

	// A slice of all the collisions in the group.
	Collisions []Collision
}

func (cg CollisionGroup) Bounds() coord.Bounds {
	bounds := make([]coord.Bounds, 0, len(cg.Collisions))
	for _, c := range cg.Collisions {
		bounds = append(bounds, c.Bounds())
	}

	return coord.JoinBounds(bounds...)
}

// Adds a collision to the group. Also adds the
// entities from the collision to the entities slice.
// Filters out collisions it already has and entities
// that are already in the entities slice.
func (cg CollisionGroup) AddCollision(c Collision) CollisionGroup {
	for _, cc := range cg.Collisions {
		if c.IsSameAs(cc) {
			return cg
		}
	}

	cg.Collisions = append(cg.Collisions, c)

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
			return cg
		}
	}
	cg.Entities = append(cg.Entities, b)

	return cg
}

// An entity may ONLY be assigned to 1 collision group.
// If an entity has collisions that are in separate collision
// groups, those groups must be merged. This rules make the
// collision group index possible.
type CollisionGroupIndex map[entity.Entity]*CollisionGroup
