package entity

import (
	"sync"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/sim/stime"
)

// A unique Id for an entity
type Id int64

type Flag uint64

const (
	FlagRemoved = 1 << iota
	FlagNew
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

// TODO Factor out Bounds from the Entity Interface
type HasBounds interface {
	Bounds() coord.Bounds
}

type CanChange interface {
	HasChanged(nextState State, now stime.Time) bool
}

// Used by the world state to calculate
// differences between world states.
// An object that implements this interface
// should also be friendly to the Json
// marshaller and expect to be sent to the
// client over the wire.
type State interface {
	EntityId() Id
	EntityCell() coord.Cell
}

type StateHasBounds interface {
	HasBounds
}

// TODO Replace this with the CanChange interface from above
type StateCanChange interface {
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
		if ee, yes := e.(HasBounds); yes {
			if bounds.Overlaps(ee.Bounds()) {
				result = append(result, e)
				continue
			}
		}

		if bounds.Contains(e.EntityCell()) {
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

var _ Entity = Removed{}
var _ State = RemovedState{}

type Removed struct {
	Entity
	RemovedAt stime.Time
}

type RemovedState struct {
	Id   Id
	Cell coord.Cell
}

func (e Removed) Flags() Flag {
	return e.Entity.Flags() | FlagRemoved | FlagNoCollide
}

func (e Removed) ToState() State {
	return RemovedState{
		Id:   e.Id(),
		Cell: e.Cell(),
	}
}

func (e RemovedState) EntityId() Id           { return e.Id }
func (e RemovedState) EntityCell() coord.Cell { return e.Cell }
func (e RemovedState) IsDifferentFrom(other State) bool {
	switch other := other.(type) {
	case RemovedState:
		return e.Id != other.Id || e.Cell != other.Cell
	default:
	}

	return true
}
