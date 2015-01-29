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
	for _, c1 := range cg.Collisions {
		for _, c2 := range ocg.Collisions {
			if c1 == c2 {
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
	for _, c := range cg.Collisions {
		fmt.Fprintf(b, "\t\t%v\n", c)
	}
	fmt.Fprint(b, "}\n")

	return b.String()
}
