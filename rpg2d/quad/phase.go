package quad

import (
	"fmt"
	"log"

	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/sim/stime"
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
// If update returns nil instead of the entity,
// it will be removed.
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
	ApplyInputsTo(target entity.Entity, now stime.Time, entities InputPhaseChanges) (targetIfChanged entity.Entity)
}

type InputPhaseChanges interface {
	New(entity.Entity)
}

// Convenience type so input phase handlers
// can be written as closures or as functions.
type InputPhaseHandlerFn func(entity.Entity, stime.Time, InputPhaseChanges) entity.Entity

func (f InputPhaseHandlerFn) ApplyInputsTo(e entity.Entity, now stime.Time, entities InputPhaseChanges) entity.Entity {
	return f(e, now, entities)
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
// a all collision groups. The phase handler should return
// an implementation of NarrowPhaseChanges that will be
// applied to the QuadTree as a result of solving the collision
// groups.
type NarrowPhaseHandler interface {
	ResolveCollisions([]*CollisionGroup, stime.Time) NarrowPhaseChanges
}

type NarrowPhaseChanges interface {
	Updated() []entity.Entity
}

var _ NarrowPhaseChanges = SliceNarrowPhaseChanges{}

type SliceNarrowPhaseChanges []entity.Entity

func (s SliceNarrowPhaseChanges) Updated() []entity.Entity {
	return s
}

// Convenience type so narrow phase handlers
// can be written as closures or as functions.
type NarrowPhaseHandlerFn func([]*CollisionGroup, stime.Time) NarrowPhaseChanges

func (f NarrowPhaseHandlerFn) ResolveCollisions(cgrps []*CollisionGroup, now stime.Time) NarrowPhaseChanges {
	return f(cgrps, now)
}

func RunPhasesOn(
	q QuadRoot,
	updatePhase UpdatePhaseHandler,
	inputPhase InputPhaseHandler,
	narrowPhase NarrowPhaseHandler,
	now stime.Time) QuadRoot {

	q =
		q.RunUpdatePhase(updatePhase, now)
	q =
		q.RunInputPhase(inputPhase, now)
	cgroups :=
		q.RunBroadPhase(now)
	q =
		q.RunNarrowPhase(narrowPhase, cgroups, now)

	return q
}

type updatePhaseChanges interface {
	modified(entity.Entity)
	removed(entity.Entity)
}

type updatePhaseSliceChanges struct {
	allModified []entity.Entity
	allRemoved  []entity.Entity
}

func (c *updatePhaseSliceChanges) modified(e entity.Entity) {
	c.allModified = append(c.allModified, e)
}
func (c *updatePhaseSliceChanges) removed(e entity.Entity) {
	c.allRemoved = append(c.allRemoved, e)
}

func (q QuadRoot) RunUpdatePhase(p UpdatePhaseHandler, now stime.Time) QuadRoot {
	if q.updatePhaseSliceChanges == nil {
		q.updatePhaseSliceChanges = &updatePhaseSliceChanges{
			allModified: make([]entity.Entity, 0, 10),
			allRemoved:  make([]entity.Entity, 0, 10),
		}
	} else {
		q.updatePhaseSliceChanges.allModified = q.updatePhaseSliceChanges.allModified[:0]
		q.updatePhaseSliceChanges.allRemoved = q.updatePhaseSliceChanges.allRemoved[:0]
	}
	q.node = q.node.runUpdatePhase(p, now, q.updatePhaseSliceChanges)

	for _, e := range q.updatePhaseSliceChanges.allModified {
		q = q.Insert(e)
	}

	for _, e := range q.updatePhaseSliceChanges.allRemoved {
		q = q.Remove(e.Id())
	}

	return q
}

func (q quadNode) runUpdatePhase(p UpdatePhaseHandler, now stime.Time, changes updatePhaseChanges) Quad {
	// TODO Implement concurrently
	//      For each child, recursively descend and run input phase
	for i, quad := range q.children {
		q.children[i] = quad.runUpdatePhase(p, now, changes)
	}

	return q
}

func (q quadLeaf) runUpdatePhase(p UpdatePhaseHandler, now stime.Time, changes updatePhaseChanges) Quad {
	for i, e := range q.entities {
		updatedEntity := p.Update(e, now)
		if updatedEntity == nil {
			changes.removed(e)
		} else {
			if updatedEntity.Cell() == e.Cell() {
				q.entities[i] = updatedEntity
			} else {
				// Return to parent to be reinserted
				changes.modified(updatedEntity)
			}
		}
	}

	return q
}

type inputPhaseChanges struct {
	entities []entity.Entity
}

func (c *inputPhaseChanges) New(e entity.Entity) {
	c.entities = append(c.entities, e)
}

func (q QuadRoot) RunInputPhase(p InputPhaseHandler, now stime.Time) (root QuadRoot) {
	if q.inputPhaseChanges == nil {
		q.inputPhaseChanges = &inputPhaseChanges{make([]entity.Entity, 0, 100)}
	} else {
		q.inputPhaseChanges.entities = q.inputPhaseChanges.entities[:0]
	}
	q.node = q.node.runInputPhase(p, now, q.inputPhaseChanges)

	root = q
	for _, e := range q.inputPhaseChanges.entities {
		root = root.Insert(e)
	}

	return root
}

func (q quadNode) runInputPhase(p InputPhaseHandler, now stime.Time, changes InputPhaseChanges) Quad {
	for i, quad := range q.children {
		q.children[i] = quad.runInputPhase(p, now, changes)
	}
	return q
}

func (q quadLeaf) runInputPhase(p InputPhaseHandler, now stime.Time, changes InputPhaseChanges) Quad {
	for i, e := range q.entities {
		var changeToE entity.Entity
		changeToE = p.ApplyInputsTo(e, now, changes)

		if changeToE != nil {
			if changeToE.Cell() != e.Cell() {
				// TODO Does the Input Phase need to be allowed to modify an entities cell?
				log.Printf("%#v", e)
				log.Printf("%#v", changeToE)
				panic("invalid modification to entities Cell during input phase")
			}
			q.entities[i] = changeToE
		}
	}

	return q
}

func (q QuadRoot) RunBroadPhase(now stime.Time) (cgroups []*CollisionGroup) {
	cgroups, _, _ = q.node.runBroadPhase(now, q.CollisionGroupPool)
	return cgroups
}

type broadPhase struct {
	cgroups          []*CollisionGroup
	solved, unsolved CollisionGroupIndex
}

func newBroadPhase(size int) *broadPhase {
	return &broadPhase{
		cgroups:  make([]*CollisionGroup, 0, size),
		solved:   make(CollisionGroupIndex, size),
		unsolved: make(CollisionGroupIndex, size),
	}
}

func (p *broadPhase) reset() {
	p.cgroups = p.cgroups[:0]
	for k := range p.solved {
		delete(p.solved, k)
	}
	for k := range p.unsolved {
		delete(p.unsolved, k)
	}
}

func (q quadNode) runBroadPhase(now stime.Time, cgroupPool *CollisionGroupPool) ([]*CollisionGroup, CollisionGroupIndex, CollisionGroupIndex) {
	if q.broadPhase == nil {
		q.broadPhase = newBroadPhase(4)
	} else {
		q.broadPhase.reset()
	}
	for _, cq := range q.children {
		cgrps, solved, unsolved := cq.runBroadPhase(now, cgroupPool)

		// Join array of collision groups
		q.cgroups = append(q.cgroups, cgrps...)

		// Joins solved collision group index
		for e, cg := range solved {
			q.solved[e] = cg
		}

		// Join unsolved collision group index
		for e, cg := range unsolved {
			q.unsolved[e] = cg
		}
	}

	// For each entity in the unsolved array
	// try to solve it by querying the children
	for e1, e1cg := range q.unsolved {
		if b, err := q.Bounds().Intersection(e1.Bounds()); err == nil && b != e1.Bounds() {
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

			e2cg, e2cgExist := q.solved[e2]

			switch {
			case e1cg == nil && !e2cgExist:
				// e1 is NOT in a collision group
				// e2 is NOT in a collision group

				// create a new collision group
				// and add a collision for e1 & e2
				cg := cgroupPool.NewGroup()
				cg.AddCollision(e1, e2)

				// add this new collision group to the array of collision groups
				q.cgroups = append(q.cgroups, cg)

				// set e1 & e2's new collision group
				q.solved[e1] = cg
				q.solved[e2] = cg

				// set e1's collision group in the for loop
				// over the unsolved map.
				// This is done because we don't drop out
				// of this this outer loop just yet and
				// further iterations must know that e1
				// is now part of a collision group.
				e1cg = cg

				// NOTE I don't know if this is necessary
				// due to the inverse reason that the above
				// statement is required.
				q.unsolved[e1] = cg

			case e1cg != nil && !e2cgExist:
				// e1 is in a collision group
				// e2 is NOT in a collision group
				cg := e1cg

				// create a new collision of e1 & e2
				// add it to e1's collision group
				cg.AddCollision(e1, e2)

				// and set e2's collision group in the collision group index
				q.solved[e2] = cg

			case e1cg == nil && e2cgExist:
				// e1 is NOT in a collision group
				// e2 is in a collision group

				// add a collision for e1 & e2 to e2's collision group
				e2cg.AddCollision(e1, e2)

				// set e1's collision group
				q.solved[e1] = e2cg

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
				q.unsolved[e1] = e2cg

			case e1cg != nil && e2cgExist && e1cg != e2cg:
				// e1 is in a collision group
				// e2 is in a collision group
				// The collision groups are different

				// merge the collision groups into e1's collision group
				for _, c := range e2cg.CollisionsById {
					e1cg.AddCollisionFromMerge(c)
				}

				// create a collision for e1 & e2
				// add it to the collision group
				e1cg.AddCollision(e1, e2)

				// set e2's new collision group and
				// remove e2's old collision group from the cgindex
				for e, cg := range q.solved {
					if cg == e2cg {
						q.solved[e] = e1cg
					}
				}

				// set e2's new collision group and
				// remove e2's old collision group from the cgindex
				for e, cg := range q.unsolved {
					if cg == e2cg {
						q.unsolved[e] = e1cg
					}
				}

				// remove e2's collision group from the array
				for i, cg := range q.cgroups {
					if cg == e2cg {
						q.cgroups = append(q.cgroups[:i], q.cgroups[i+1:]...)
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
		delete(q.unsolved, e1)
	}

	return q.cgroups, q.solved, q.unsolved
}

func (q quadLeaf) runBroadPhase(now stime.Time, cgroupPool *CollisionGroupPool) ([]*CollisionGroup, CollisionGroupIndex, CollisionGroupIndex) {
	if !(len(q.entities) > 0) {
		return nil, nil, nil
	}

	if q.broadPhase == nil {
		q.broadPhase = newBroadPhase(q.maxSize)
	} else {
		q.broadPhase.reset()
	}

	for _, e1 := range q.entities {
		// TODO Add test cases for no collisions
		// Ignore entities that have no collisions
		if e1.Flags()&entity.FlagNoCollide != 0 {
			continue
		}

		for _, e2 := range q.entities {
			// Check for self
			if e1.Id() == e2.Id() {
				continue
			}

			// Ignore entities that have no collisions
			if e2.Flags()&entity.FlagNoCollide != 0 {
				continue
			}

			// Check for overlap
			if !e1.Bounds().Overlaps(e2.Bounds()) {
				continue
			}

			e1cg, e1cgExists := q.solved[e1]
			e2cg, e2cgExists := q.solved[e2]

			switch {
			case e1cgExists && !e2cgExists:
				// e1 exists in a collision group already
				// create a collision and add it to e1's group
				e1cg.AddCollision(e1, e2)

				q.solved[e1] = e1cg
				q.solved[e2] = e1cg

			case !e1cgExists && e2cgExists:
				// e2 exists in a collision group already
				// create a collision and add it to e2's group
				e2cg.AddCollision(e1, e2)

				q.solved[e1] = e2cg
				q.solved[e2] = e2cg

			case !e1cgExists && !e2cgExists:
				// neither e1 or e2 have been assigned to a collision group
				// create a new collision group
				cg := cgroupPool.NewGroup()

				// add a collision between e1 and e2
				cg.AddCollision(e1, e2)

				// set the cgroup in the cgroup index
				q.solved[e1] = cg
				q.solved[e2] = cg

				// append the cgroup to the array of cgroups
				q.cgroups = append(q.cgroups, cg)

			case (e1cgExists && e2cgExists) && (e1cg != e2cg):
				// both entities exist in a collision group
				// but those collision groups are different

				// merge the collision groups
				for _, c := range e2cg.CollisionsById {
					e1cg.AddCollisionFromMerge(c)
				}

				// add a collision for e1 && e2
				e1cg.AddCollision(e1, e2)

				// set e2's new collision group and
				// remove e2's old collision group from the cgindex
				for e, cg := range q.solved {
					if cg == e2cg {
						q.solved[e] = e1cg
					}
				}

				// remove e2's collision group from the array
				for i, cg := range q.cgroups {
					if cg == e2cg {
						q.cgroups = append(q.cgroups[:i], q.cgroups[i+1:]...)
						break
					}
				}

			case (e1cgExists && e2cgExists) && (e1cg == e2cg):
				// both entities exist in the same collision group already
				// Add a collision for e1 && e2
				cg := e1cg
				cg.AddCollision(e1, e2)

			default:
				panic(fmt.Sprintf(`unexpected index state during broad phase
e1cg = %v
e2cg = %v
`, e1cg, e2cg))
			}
		}
	}

	// If the bounds of the entity extend beyond the quad leaf's
	// bounds, it is passed to the parent with the collision
	// group the entity belongs to
	for _, e := range q.entities {
		// If entities bounds extend past quad's bounds the
		// intersection of the 2 will be different than
		// the entities bounds.
		if b, _ := e.Bounds().Intersection(q.Bounds()); b != e.Bounds() {
			q.unsolved[e] = q.solved[e]
			// TODO should this collision group be removed from the solved index?
			// TODO test if this breaks anything
		}
	}

	return q.cgroups, q.solved, q.unsolved
}

func (q QuadRoot) RunNarrowPhase(narrowPhase NarrowPhaseHandler, cgroups []*CollisionGroup, now stime.Time) QuadRoot {
	changes := narrowPhase.ResolveCollisions(cgroups, now)
	for _, e := range changes.Updated() {
		q = q.Insert(e)
	}
	return q
}
