package entity

import "github.com/ghthor/engine/rpg2d/coord"

// A basic entity in the world.
type Entity interface {
	// Unique ID
	Id() int64

	// Location in the world
	Cell() coord.Cell

	// Returns a bounding object incorporating
	// the entities potential movement targets
	Bounds() coord.Bounds
}

// Represents the current motion state of an
// entity that can move around in the world.
type MotionState struct{}

// An entity that has the ability to move
// through the world.
type Movable interface {
	Entity

	// Returns the entities current motion state
	MotionState() MotionState
}

// An entity that can collide with other entities.
type Collidable interface {
	Entity

	CanCollideWith(other Collidable) bool
	HadCollisionWith(other Collidable)
}
