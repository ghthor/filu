package quad

// 1. Input Application Phase - User Defined
//
// The input phase takes the user input and applies
// it to the entities movement state. This enables the
// broad phase to estimate where collisions might take
// place if the movement is valid.
// The narrow phase should revert any changes to the movement
// state that cannot happen because of collisions.
type InputPhaseHandler interface {
	InputPhaseIn(Chunk) Chunk
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
	NarrowPhaseIn(Chunk) Chunk
}
