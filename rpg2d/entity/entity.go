package entity

import "github.com/ghthor/engine/rpg2d/coord"

// A basic entity in the world.
type Entity interface {
	// Unique ID
	Id() int64

	// Location in the world
	Cell() coord.Cell

	// Returns a bounding object incorporating
	// the entities potential interaction with
	// the other entities in the world.
	Bounds() coord.Bounds
}
