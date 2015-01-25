package quad

// The movement phase is the naive phase that takes
// the user input and applies it to the entities movement
// state. This enables the broad phase estimate where
// collisions might take place if the movement is valid.
// The narrow phase should revert any changes to the movement
// state that cannot happen because of collisions.
type MovementPhaseHandler interface {
	MovementPhaseIn(Chunk) Chunk
}

// The broad phase is mostly handled internally.
// But this handler allows pkg user's to hook into
// the broad phase. The broad phase is what discovers
// chunks of interest.
type BroadPhaseHandler interface {
	BroadPhaseIn(Chunk) Chunk
}

// A collisionhandler takes a chunk of entities
// and does collision checks for the chunk and modifies
// the entities. All entities in the returned chunk will
// be reinserted into the quad tree to update their location.
type NarrowPhaseHandler interface {
	NarrowPhaseIn(Chunk) Chunk
}
