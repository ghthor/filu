package quad

import (
	"bytes"
	"fmt"
)

func (c Collision) Equals(other interface{}) bool {
	switch oc := other.(type) {
	case Collision:
		return c.IsSameAs(oc)
	default:
	}

	return false
}

func (a CollisionIndex) Equals(other interface{}) bool {
	switch b := other.(type) {
	case CollisionIndex:
		return a.isTheSameAs(b) && b.isTheSameAs(a)
	default:
	}

	return false
}

func (a CollisionIndex) isTheSameAs(b CollisionIndex) bool {
	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for e, a_collisions := range a {
		b_collisions, exists := b[e]
		if !exists {
			return false
		}

		if len(a_collisions) != len(b_collisions) {
			return false
		}

	toNextCollision:
		for _, ac := range a_collisions {
			for _, bc := range b_collisions {
				if ac.IsSameAs(bc) {
					continue toNextCollision
				}
			}
			return false
		}
	}

	return true
}

func (cg CollisionGroup) Equals(other interface{}) bool {
	switch ocg := other.(type) {
	case CollisionGroup:
		return cg.hasSameCollisionsAs(ocg) && ocg.hasSameCollisionsAs(cg)
	case *CollisionGroup:
		return cg.hasSameCollisionsAs(*ocg) && ocg.hasSameCollisionsAs(cg)
	default:
	}

	return false
}

func (cg CollisionGroup) hasSameCollisionsAs(ocg CollisionGroup) bool {
toNextEntity:
	for id, c1 := range cg.CollisionsById {
		if c2, exists := ocg.CollisionsById[id]; exists {
			if c1.Equals(c2) {
				continue toNextEntity
			}
		}
		return false
	}
	return true
}

func (cg CollisionGroup) String() string {
	b := bytes.NewBuffer(make([]byte, 0, 1024))

	fmt.Fprint(b, "CollisionGroup {\n")
	fmt.Fprint(b, "\tCollisions:\n")
	for _, c := range cg.CollisionsById {
		fmt.Fprintf(b, "\t\t%v\n", c)
	}
	fmt.Fprint(b, "}\n")

	return b.String()
}

func (a CollisionGroupIndex) Equals(other interface{}) bool {
	switch b := other.(type) {
	case CollisionGroupIndex:
		return a.indexsAreTheSame(b) && b.indexsAreTheSame(a)
	default:
	}
	return false
}

func (a CollisionGroupIndex) indexsAreTheSame(b CollisionGroupIndex) bool {
	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for e, cgA := range a {
		cgB, exists := b[e]

		switch {
		case !exists:
			return false

		case cgA == nil && cgB == nil:

		case cgB == nil && cgA != nil:
			fallthrough
		case cgA == nil && cgB != nil:
			fallthrough

		case !cgA.Equals(cgB):
			return false

		default:
		}
	}

	return true
}
