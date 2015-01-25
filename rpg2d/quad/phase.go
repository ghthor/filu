package quad

import "github.com/ghthor/engine/rpg2d/entity"

// 1. Input Application Phase - User Defined
//
// The input phase takes the user input and applies
// it to the entities movement state. This enables the
// broad phase to estimate where collisions might take
// place if the movement is valid.
// The narrow phase should revert any changes to the movement
// state that cannot happen because of collisions.
type InputPhaseHandler interface {
	ApplyInputsIn(Chunk) Chunk
}

// Convenience type so input phase handlers can be written
// as closures or as functions.
type InputPhaseHandlerFn func(Chunk) Chunk

func (f InputPhaseHandlerFn) ApplyInputsIn(c Chunk) Chunk {
	return f(c)
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
	ResolveCollisions(Chunk) Chunk
}

func (q quadNode) RunInputPhase(p InputPhaseHandler) (Quad, []entity.Entity) {
	// TODO Implement this method more efficiently
	// It may use a lot of memory because of all the
	// slice creation/appending/copying.
	var outOfBounds []entity.Entity

	// TODO Implement concurrently
	// For each child, recursively descend and run input phase
	for i, quad := range q.children {
		quad, oobc := quad.RunInputPhase(p)
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

func (q quadLeaf) RunInputPhase(p InputPhaseHandler) (Quad, []entity.Entity) {
	chunk := p.ApplyInputsIn(q.Chunk())

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
