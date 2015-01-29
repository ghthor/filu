package quad

import (
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
)

// A collisions between 2 entities because the
// entities bounds are overlapping.
type Collision struct {
	A, B entity.Entity
}

func (c Collision) Bounds() coord.Bounds {
	return coord.JoinBounds(c.A.Bounds(), c.B.Bounds())
}

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
type CollisionGroup interface {
	Bounds() coord.Bounds
	Collisions() []Collision
}

// None of these should EVER reach the narrow phase.
// They are returned from the quadLeaf during the
// broad phase and they should be destroyed by the
// quadNode's broad phase if they aren't merged
// with an actuall collision group.
type singleEntity struct {
	entity.Entity
}

func (singleEntity) Collisions() []Collision { return nil }

type collisionGroup struct {
	collisions []Collision
}

// Merge 2 collision groups into a single group.
// Doesn't verify that the 2 collision groups should me
// merged. Only used during internal broad phase.
func mergeCollisionGroups(a, b CollisionGroup) CollisionGroup {
	return a
}
