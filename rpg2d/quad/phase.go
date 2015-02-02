package quad

import (
	"fmt"

	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/sim/stime"
)

// 1. Input Application Phase - User Defined
//
// The input phase takes the user input and applies
// it to the entities movement state. This enables the
// broad phase to estimate where collisions might take
// place if the movement is valid.
// The narrow phase should revert any changes to the movement
// state that cannot happen because of collisions.
type InputPhaseHandler interface {
	ApplyInputsIn(Chunk, stime.Time) Chunk
}

// Convenience type so input phase handlers can be written
// as closures or as functions.
type InputPhaseHandlerFn func(Chunk, stime.Time) Chunk

func (f InputPhaseHandlerFn) ApplyInputsIn(c Chunk, t stime.Time) Chunk {
	return f(c, t)
}

// 2. Broad Phase - Internal
//
// The broad phase is only concerned with the bounds
// of all the entities potential futures. For this
// reason it will be and internal implementation.
// The broad phase will need to happen while recursively
// descending and ascending through the tree. It will be
// where chunks of interest are created based on the
// potential future overlapping bounds of the entities.
// The chunks of interest will then be passed into the
// user supplied narrow phase handler.

// 3. Narrow Phase - User Defined
//
// The narrow phase resolves any collisions that the
// broad phase identified by the bounds of the potential
// future state of the entity. The broad phase has grouped
// this chunk together, such that, there is no possible way
// any of these entities can interact with any entities that
// are not included in this chunk. The narrow phase should
// return a chunk.... Honestly I don't know what it should
// return and I need to stop thinking about it for now and
// just shut the fuck up and write the fucking code.
type NarrowPhaseHandler interface {
	ResolveCollisions(CollisionGroup, stime.Time) CollisionGroup
}

func RunPhasesOn(q Quad, inputPhase InputPhaseHandler, narrowPhase NarrowPhaseHandler, now stime.Time) Quad {
	q, _ = q.runInputPhase(inputPhase, now)
	q, cgroups, _, _ := q.runBroadPhase(now)
	q = q.runNarrowPhase(narrowPhase, cgroups, now)

	return q
}

func RunInputPhaseOn(q Quad, inputPhase InputPhaseHandler, now stime.Time) (Quad, []entity.Entity) {
	return q.runInputPhase(inputPhase, now)
}

func RunBroadPhaseOn(q Quad, now stime.Time) (quad Quad, cgroup []*CollisionGroup, solved, unsolved CollisionGroupIndex) {
	return q.runBroadPhase(now)
}

func RunNarrowPhaseOn(q Quad, cgroups []*CollisionGroup, narrowPhase NarrowPhaseHandler, now stime.Time) Quad {
	return q.runNarrowPhase(narrowPhase, cgroups, now)
}

func (q quadNode) runInputPhase(p InputPhaseHandler, at stime.Time) (Quad, []entity.Entity) {
	// TODO Implement this method more efficiently
	// It may use a lot of memory because of all the
	// slice creation/appending/copying.
	var outOfBounds []entity.Entity

	// TODO Implement concurrently
	// For each child, recursively descend and run input phase
	for i, quad := range q.children {
		quad, oobc := quad.runInputPhase(p, at)
		q.children[i] = quad
		outOfBounds = append(outOfBounds, oobc...)
	}

	// Use a var of the interface value
	// This enables us to use Insert() method functionally
	var quad Quad = q

	// For each entity that was out of bounds for a child
	// check if it is still within our own bounds.
	for i, e := range outOfBounds {
		if quad.Bounds().Contains(e.Cell()) {
			quad = quad.Insert(e)
			outOfBounds = append(outOfBounds[:i], outOfBounds[i+1:]...)
		}
	}

	return quad, outOfBounds
}

func (q quadLeaf) runInputPhase(p InputPhaseHandler, now stime.Time) (Quad, []entity.Entity) {
	chunk := p.ApplyInputsIn(q.Chunk(), now)

	// Use a var of the interface value
	// This enables us to use Remove() method functionally
	var quad Quad = q
	var outOfBounds []entity.Entity

	for _, e := range chunk.Entities {
		if !quad.Bounds().Contains(e.Cell()) {
			quad = quad.Remove(e)
			outOfBounds = append(outOfBounds, e)
		}
	}

	return quad, outOfBounds
}

func (q quadNode) runBroadPhase(now stime.Time) (quad Quad, cgroups []*CollisionGroup, solved, unsolved CollisionGroupIndex) {
	for i, cq := range q.children {
		cq, cgrps, s, u := cq.runBroadPhase(now)
		q.children[i] = cq

		// Join array of collision groups
		cgroups = append(cgroups, cgrps...)

		// TODO Rip this out into a method of collision group index.
		// Join solved collision group index
		if solved == nil {
			solved = s
		} else {
			for e, cg := range s {
				solved[e] = cg
			}
		}

		// Join unsolved collision group index
		if unsolved == nil {
			unsolved = u
		} else {
			for e, cg := range u {
				unsolved[e] = cg
			}
		}
	}

	// For each entity in the unsolved array
	// try to solve it by querying the children
	for e1, e1cg := range unsolved {
		if b, _ := q.Bounds().Intersection(e1.Bounds()); b != e1.Bounds() {
			// The entities bounds extend beyond the quad tree's bounds
			// and therefore we can't solve this entity here either
			// and will bubble to our parent
			// TODO Specify that this behavior works as expected near the
			// edges of the entire quad tree's bounds.
			continue
		}

		// Query for any overlapping entities
		overlappingEntities := q.QueryBounds(e1.Bounds())

		for _, e2 := range overlappingEntities {
			// ignore self
			if e1 == e2 {
				continue
			}

			e2cg, e2cgExist := solved[e2]

			switch {
			case e1cg == nil && !e2cgExist:
				// e1 is NOT in a collision group
				// e2 is NOT in a collision group

				// create a new collision group
				// and add a collision for e1 & e2
				cg := CollisionGroup{}.AddCollision(Collision{e1, e2})

				// add this new collision group to the array of collision groups
				cgroups = append(cgroups, &cg)

				// set e1 & e2's new collision group
				solved[e1] = &cg
				solved[e2] = &cg

				// set e1's collision group in the for loop
				// over the unsolved map.
				// This is done because we don't drop out
				// of this this outer loop just yet and
				// further iterations must know that e1
				// is now part of a collision group.
				e1cg = &cg

				// NOTE I don't know if this is necessary
				// due to the inverse reason that the above
				// statement is required.
				unsolved[e1] = &cg

			case e1cg != nil && !e2cgExist:
				// e1 is in a collision group
				// e2 is NOT in a collision group
				cg := e1cg

				// create a new collision of e1 & e2
				// add it to e1's collision group
				*cg = cg.AddCollision(Collision{e1, e2})

				// and set e2's collision group in the collision group index
				solved[e2] = cg

			case e1cg == nil && e2cgExist:
				// e1 is NOT in a collision group
				// e2 is in a collision group

				// add a collision for e1 & e2 to e2's collision group
				*e2cg = e2cg.AddCollision(Collision{e1, e2})

				// set e1's collision group
				solved[e1] = e2cg

				// set e1's collision group in the for loop
				// over the unsolved map.
				// This is done because we don't drop out
				// of this this outer loop just yet and
				// further iterations must know that e1
				// is now part of a collision group.
				e1cg = e2cg

				// NOTE I don't know if this is necessary
				// due to the inverse reason that the above
				// statement is required.
				unsolved[e1] = e2cg

			case e1cg != nil && e2cgExist && e1cg != e2cg:
				// e1 is in a collision group
				// e2 is in a collision group
				// The collision groups are different

				// merge the collision groups into e1's collision group
				for _, c := range e2cg.Collisions {
					*e1cg = e1cg.AddCollision(c)
				}

				// create a collision for e1 & e2
				// add it to the collision group
				*e1cg = e1cg.AddCollision(Collision{e1, e2})

				// set e2's new collision group and
				// remove e2's old collision group from the cgindex
				for e, cg := range solved {
					if cg == e2cg {
						solved[e] = e1cg
					}
				}

				// set e2's new collision group and
				// remove e2's old collision group from the cgindex
				for e, cg := range unsolved {
					if cg == e2cg {
						unsolved[e] = e1cg
					}
				}

				// remove e2's collision group from the array
				for i, cg := range cgroups {
					if cg == e2cg {
						cgroups = append(cgroups[:i], cgroups[i+1:]...)
						break
					}
				}

			case e1cg != nil && e2cgExist && e1cg == e2cg:
				// e1 and e2 exist in the same collision group
				// NOTE this means e2 is from the same quad as
				// e1 so we can do nothing because there should be
				// a collision already.

			default:
				panic("unexpected index state when solving bubbled entities")
			}
		}

		// remove the entity from the unsolved map for it is now solved
		delete(unsolved, e1)
	}

	return q, cgroups, solved, unsolved
}

func (q quadLeaf) runBroadPhase(stime.Time) (quad Quad, cgroups []*CollisionGroup, solved, unsolved CollisionGroupIndex) {
	if !(len(q.entities) > 0) {
		return q, nil, nil, nil
	}

	cgindex := make(map[entity.Entity]*CollisionGroup, len(q.entities))
	cgroups = make([]*CollisionGroup, 0, len(q.entities))

	for _, e1 := range q.entities {
		for _, e2 := range q.entities {
			// Check for self
			if e1 == e2 {
				continue
			}

			// Check for overlap
			if !e1.Bounds().Overlaps(e2.Bounds()) {
				continue
			}

			e1cg, e1cgExists := cgindex[e1]
			e2cg, e2cgExists := cgindex[e2]

			switch {
			case e1cgExists && !e2cgExists:
				// e1 exists in a collision group already
				// create a collision and add it to e1's group
				c := Collision{e1, e2}
				*e1cg = e1cg.AddCollision(c)

				cgindex[e1] = e1cg
				cgindex[e2] = e1cg

			case !e1cgExists && e2cgExists:
				// e2 exists in a collision group already
				// create a collision and add it to e2's group
				c := Collision{e1, e2}
				*e2cg = e2cg.AddCollision(c)

				cgindex[e1] = e2cg
				cgindex[e2] = e2cg

			case !e1cgExists && !e2cgExists:
				// neither e1 or e2 have been assigned to a collision group
				// create a new collision group
				cg := &CollisionGroup{
					make([]entity.Entity, 0, 2),
					make([]Collision, 0, 1),
				}

				// add a collision between e1 and e2
				*cg = cg.AddCollision(Collision{e1, e2})

				// set the cgroup in the cgroup index
				cgindex[e1] = cg
				cgindex[e2] = cg

				// append the cgroup to the array of cgroups
				cgroups = append(cgroups, cg)

			case (e1cgExists && e2cgExists) && (e1cg != e2cg):
				// both entities exist in a collision group
				// but those collision groups are different

				// merge the collision groups
				for _, c := range e2cg.Collisions {
					*e1cg = e1cg.AddCollision(c)
				}

				// add a collision for e1 && e2
				*e1cg = e1cg.AddCollision(Collision{e1, e2})

				// set e2's new collision group and
				// remove e2's old collision group from the cgindex
				for e, cg := range cgindex {
					if cg == e2cg {
						cgindex[e] = e1cg
					}
				}

				// remove e2's collision group from the array
				for i, cg := range cgroups {
					if cg == e2cg {
						cgroups = append(cgroups[:i], cgroups[i+1:]...)
						break
					}
				}

			case (e1cgExists && e2cgExists) && (e1cg == e2cg):
				// both entities exist in the same collision group already
				// Add a collision for e1 && e2
				cg := e1cg
				*cg = cg.AddCollision(Collision{e1, e2})

			default:
				panic(fmt.Sprintf(`unexpected index state during broad phase
e1cg = %v
e2cg = %v
`, e1cg, e2cg))
			}
		}
	}

	// Solved entities are in the cgindex
	solved = cgindex
	// Build a collision group index for the unsolvable entities
	unsolved = make(map[entity.Entity]*CollisionGroup)

	// If the bounds of the entity extend beyond the quad leaf's
	// bounds, it is passed to the parent with the collision
	// group the entity belongs to
	for _, e := range q.entities {
		// If entities bounds extend past quad's bounds the
		// intersection of the 2 will be different than
		// the entities bounds.
		if b, _ := e.Bounds().Intersection(q.Bounds()); b != e.Bounds() {
			unsolved[e] = cgindex[e]
		}
	}

	return q, cgroups, cgindex, unsolved
}

func (q quadNode) runNarrowPhase(narrowPhase NarrowPhaseHandler, cgrpsToSolve []*CollisionGroup, now stime.Time) Quad {
	return q
}

func (q quadLeaf) runNarrowPhase(narrowPhase NarrowPhaseHandler, cgrpsToSolve []*CollisionGroup, now stime.Time) Quad {
	return q
}
