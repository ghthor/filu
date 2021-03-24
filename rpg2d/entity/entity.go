package entity

import (
	"sync"

	"github.com/ghthor/filu/rpg2d/coord"
)

// A unique Id for an entity
type Id int64

type Flag uint64

const (
	FlagNew = 1 << iota
	FlagChanged
	FlagNoCollide
	FlagUserDefined
)

func (f Flag) Set(bits Flag) Flag {
	return f | bits
}

func (f Flag) Unset(bits Flag) Flag {
	return f &^ bits
}

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

	// Returns the set of entity flags that can
	// set whether something collides or whether
	// an entity has been changed or not.
	Flags() Flag

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
	EntityId() Id

	// Bounds of the entity
	Bounds() coord.Bounds

	// Compare to another entity
	IsDifferentFrom(State) bool
}

type StateSlice []State

func (s StateSlice) FilterByBounds(result StateSlice, bounds coord.Bounds) StateSlice {
	if result == nil {
		result = make(StateSlice, 0, len(s))
	} else {
		result = result[:0]
	}

	for _, e := range s {
		if bounds.Overlaps(e.Bounds()) {
			result = append(result, e)
			continue
		}
	}

	return result
}

type StateById map[Id]State

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
