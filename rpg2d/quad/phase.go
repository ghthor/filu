package quad

import (
	"fmt"

	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/sim/stime"
)

// 1. Update Position Phase - User Defined
//
// This phase is for updating an entity's
// position if their previous movement has
// be completed on this tick. This must happen
// before the input application phase so the
// quad tree can be queried during that phase
// and the information won't be racy depending
// on the order that the input hase been applied.
type UpdatePhaseHandler interface {
	Update(entity.Entity, stime.Time) entity.Entity
}

// Convenience type so input phase handlers
// can be written as closures or as functions.
type UpdatePhaseHandlerFn func(entity.Entity, stime.Time) entity.Entity

func (f UpdatePhaseHandlerFn) Update(e entity.Entity, now stime.Time) entity.Entity {
	return f(e, now)
}

// 2. Input Application Phase - User Defined
//
// The input phase takes the user input and applies it.
// This application can modify the entities movement
// state or create new entities.
// The input phase should return a slice of
// entities that includes the entity that it was applying
// input to. This slice of entities can also include
// any entities that may have been created by the actor's input
// (using skills and spells, chat messages, etc).
// These new entities will be inserted into the
// quad tree.
type InputPhaseHandler interface {
	ApplyInputsTo(entity.Entity, stime.Time) []entity.Entity
}

// Convenience type so input phase handlers
// can be written as closures or as functions.
type InputPhaseHandlerFn func(entity.Entity, stime.Time) []entity.Entity

func (f InputPhaseHandlerFn) ApplyInputsTo(e entity.Entity, now stime.Time) []entity.Entity {
	return f(e, now)
}

// 3. Broad Phase - Internal
//
// The broad phase is only concerned with the bounds
// of all the entities and therefor can be implemented
// internally.
// The broad phase creates collision groups of entities.
// A collision group is a collection of entities
// such that no entity within the group can interact
// with any entities outside of it's collision group.
// The collision groups will be passed to the
// user supplied narrow phase implementation.

// 4. Narrow Phase - User Defined
//
// The narrow phase resolves all the collisions in
// a collision group. The phase handler should return
// 2 slices of entities. The first is the entities
// that still exist or have been created. The second
// is any entities that have been destroyed.
// The user implementation should also be
// where movement actions are accepted
// and an entities position is modified.
type NarrowPhaseHandler interface {
	ResolveCollisions(*CollisionGroup, stime.Time) (entities []entity.Entity, removed []entity.Entity)
}

// Convenience type so narrow phase handlers
// can be written as closures or as functions.
type NarrowPhaseHandlerFn func(*CollisionGroup, stime.Time) ([]entity.Entity, []entity.Entity)

func (f NarrowPhaseHandlerFn) ResolveCollisions(cgrp *CollisionGroup, now stime.Time) ([]entity.Entity, []entity.Entity) {
	return f(cgrp, now)
}

func RunPhasesOn(
	q Quad,
	updatePhase UpdatePhaseHandler,
	inputPhase InputPhaseHandler,
	narrowPhase NarrowPhaseHandler,
	now stime.Time) Quad {

	q, _ = RunUpdatePositionPhaseOn(q, updatePhase, now)
	q, _ = RunInputPhaseOn(q, inputPhase, now)
	cgroups, _, _ := RunBroadPhaseOn(q, now)
	q, _ = RunNarrowPhaseOn(q, cgroups, narrowPhase, now)

	return q
}

func RunUpdatePositionPhaseOn(q Quad, updatePhase UpdatePhaseHandler, now stime.Time) (Quad, []entity.Entity) {
	return q.runUpdatePositionPhase(updatePhase, now)
}

func RunInputPhaseOn(
	q Quad,
	inputPhase InputPhaseHandler,
	now stime.Time) (Quad, []entity.Entity) {

	return q.runInputPhase(inputPhase, now)
}

func RunBroadPhaseOn(
	q Quad,
	now stime.Time) (cgroups []*CollisionGroup, solved, unsolved CollisionGroupIndex) {

	return q.runBroadPhase(now)
}

func RunNarrowPhaseOn(
	q Quad,
	cgroups []*CollisionGroup,
	narrowPhase NarrowPhaseHandler,
	now stime.Time) (Quad, []entity.Entity) {

	var toBeInserted, toBeRemoved []entity.Entity

	for _, cg := range cgroups {
		existing, removed := narrowPhase.ResolveCollisions(cg, now)
		toBeInserted = append(toBeInserted, existing...)
		toBeRemoved = append(toBeRemoved, removed...)
	}

	for _, e := range toBeRemoved {
		q = q.Remove(e)
	}

	for _, e := range toBeInserted {
		q = q.Insert(e)
	}

	return q, nil
}

func (q quadRoot) runUpdatePositionPhase(p UpdatePhaseHandler, now stime.Time) (quad Quad, bubbled []entity.Entity) {
	q.Quad, bubbled = q.Quad.runUpdatePositionPhase(p, now)

	quad = q
	for _, e := range bubbled {
		quad = quad.Insert(e)
	}

	return quad, nil
}

func (q quadLeaf) runUpdatePositionPhase(p UpdatePhaseHandler, now stime.Time) (Quad, []entity.Entity) {
	var bubbled []entity.Entity

	for _, e := range q.entities {
		e := p.Update(e, now)
		bubbled = append(bubbled, e)
	}

	return q, bubbled
}

func (q quadNode) runUpdatePositionPhase(p UpdatePhaseHandler, now stime.Time) (Quad, []entity.Entity) {
	var bubbled []entity.Entity

	// TODO Implement concurrently
	//      For each child, recursively descend and run input phase
	for i, quad := range q.children {
		quad, entities := quad.runUpdatePositionPhase(p, now)
		q.children[i] = quad
		bubbled = append(bubbled, entities...)
	}

	return q, bubbled
}

func (q quadRoot) runInputPhase(p InputPhaseHandler, now stime.Time) (quad Quad, bubbled []entity.Entity) {
	q.Quad, bubbled = q.Quad.runInputPhase(p, now)

	quad = q
	for _, e := range bubbled {
		quad = quad.Insert(e)
	}

	return quad, nil
}

func (q quadNode) runInputPhase(p InputPhaseHandler, now stime.Time) (Quad, []entity.Entity) {
	var bubbled []entity.Entity

	// TODO Implement concurrently
	//      For each child, recursively descend and run input phase
	for i, quad := range q.children {
		quad, entities := quad.runInputPhase(p, now)
		q.children[i] = quad
		bubbled = append(bubbled, entities...)
	}

	return q, bubbled
}

func (q quadLeaf) runInputPhase(p InputPhaseHandler, now stime.Time) (Quad, []entity.Entity) {
	var bubbled []entity.Entity

	for _, e := range q.entities {
		entities := p.ApplyInputsTo(e, now)
		bubbled = append(bubbled, entities...)
	}

	return q, bubbled
}

// Broad phase is non-mutative and therefor doesn't
// require a method on the quadRoot type.

func (q quadNode) runBroadPhase(now stime.Time) (cgroups []*CollisionGroup, solved, unsolved CollisionGroupIndex) {
	for _, cq := range q.children {
		cgrps, s, u := cq.runBroadPhase(now)

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

	return cgroups, solved, unsolved
}

func (q quadLeaf) runBroadPhase(stime.Time) (cgroups []*CollisionGroup, solved, unsolved CollisionGroupIndex) {
	if !(len(q.entities) > 0) {
		return nil, nil, nil
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

	return cgroups, cgindex, unsolved
}
