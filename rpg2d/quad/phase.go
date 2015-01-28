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
	ResolveCollisions(Chunk, stime.Time) Chunk
}

func RunPhasesOn(q Quad, inputPhase InputPhaseHandler, narrowPhase NarrowPhaseHandler, now stime.Time) Quad {
	q, _ = q.runInputPhase(inputPhase, now)
	q, chunksToSolve := q.runBroadPhase(now)
	q = q.runNarrowPhase(narrowPhase, chunksToSolve, now)

	return q
}

func RunInputPhaseOn(q Quad, inputPhase InputPhaseHandler, now stime.Time) (Quad, []entity.Entity) {
	return q.runInputPhase(inputPhase, now)
}

func RunBroadPhaseOn(q Quad, now stime.Time) (quad Quad, chunksOfInterest []Chunk) {
	return q.runBroadPhase(now)
}

func RunNarrowPhaseOn(q Quad, chunksToSolve []Chunk, narrowPhase NarrowPhaseHandler, now stime.Time) []Chunk {
	return nil
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

func (q quadNode) runBroadPhase(now stime.Time) (quad Quad, chunksOfActivity []Chunk) {
	var chunks []Chunk

	for i, cq := range q.children {
		cq, chunksOfActivity := cq.runBroadPhase(now)
		q.children[i] = cq
		chunks = append(chunks, chunksOfActivity...)
	}

	// TODO Join chunks that are overlapping

	return q, chunks
}

func (q quadLeaf) runBroadPhase(stime.Time) (quad Quad, chunksOfActivity []Chunk) {
	// A map of the chunks generated thus far
	chunkIndex := make(map[entity.Entity]*Chunk, len(q.entities))
	chunkPtrs := make([]*Chunk, 0, len(q.entities))

	for _, e1 := range q.entities {
		for _, e2 := range q.entities {
			if e1 == e2 {
				continue
			}

			if !e1.Bounds().Overlaps(e2.Bounds()) {
				continue
			}

			// We have 2 entities with overlapping bounds.
			// Are either of them indexed in the chunks map?
			e1c, e1cExists := chunkIndex[e1]
			e2c, e2cExists := chunkIndex[e2]

			switch {

			case (e1cExists && !e2cExists):
				// e1 exists in a chunk already
				// add e2 to that chunk
				e1c.Entities = append(e1c.Entities, e2)

				// set e2's chunk to e1's chunk in the chunks index.
				chunkIndex[e2] = e1c

			case (!e1cExists && e2cExists):
				// e2 exists in a chunk already
				// add e1 to that chunk
				e2c.Entities = append(e2c.Entities, e1)

				// set e1's chunk to e1's chunk in the index
				chunkIndex[e1] = e2c
			case !e1cExists && !e2cExists:
				// neither e1 or e2 have been assigned to a chunk
				// create a new chunk with containing e1 and e2
				ch := &Chunk{Entities: []entity.Entity{e1, e2}}

				// set e1 and e2's chunk in the index
				chunkIndex[e1] = ch
				chunkIndex[e2] = ch

				// add chunk into the array of chunks
				chunkPtrs = append(chunkPtrs, ch)

			case (e1cExists && e2cExists) && (e1c != e2c):
				// both entities exist in chunks
				// but those chunks are different

				// merge the chunks
				e1c.Entities = append(e1c.Entities, e2c.Entities...)
				chunkIndex[e2] = e1c

			case (e1cExists && e2cExists) && (e1c == e2c):
			// both entities exist in the same chunk already.
			// do nothing

			default:
				panic(fmt.Sprintf(`unexpected index state during broad phase
e1c = %v
e2c = %v
`, e1c, e2c))
			}
		}
	}

	chunks := make([]Chunk, 0, len(chunkPtrs))
	for _, cptr := range chunkPtrs {
		chunks = append(chunks, *cptr)
	}

	// TODO Combine bounds of all entities to from bounds for each chunk

	return q, chunks
}

func (q quadNode) runNarrowPhase(narrowPhase NarrowPhaseHandler, chunksToSolve []Chunk, now stime.Time) Quad {
	return q
}

func (q quadLeaf) runNarrowPhase(narrowPhase NarrowPhaseHandler, chunksToSolve []Chunk, now stime.Time) Quad {
	return q
}
