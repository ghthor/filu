package entity

import (
	"sync"

	"github.com/ghthor/engine/rpg2d/coord"
)

// A unique Id for an entity
type Id int64

// A basic entity in the world.
type Entity interface {
	// Unique ID
	Id() Id

	// Location in the world
	Cell() coord.Cell

	// Returns a bounding object incorporating
	// the entities potential interaction with
	// the other entities in the world.
	Bounds() coord.Bounds

	// Returns a state value that represents
	// the entity in its current state.
	ToState() State
}

// Used by the world state to calculate
// differences between world states.
// An object that implements this interface
// should also be friendly to the Json
// marshaller and expect to be sent to the
// client over the wire.
type State interface {
	// Unique ID
	Id() Id

	// Bounds of the entity
	Bounds() coord.Bounds

	// Compare to another entity
	IsDifferentFrom(State) bool
}

type StateSlice []State

// Returns a function to generate consecutive
// entity Id's that is safe to call concurrently.
func NewIdGenerator() func() Id {
	var (
		mu     sync.Mutex
		nextId Id
	)

	return func() Id {
		mu.Lock()
		defer func() {
			nextId++
			mu.Unlock()
		}()
		return nextId
	}
}
