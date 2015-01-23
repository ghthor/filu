package quad

import (
	"github.com/ghthor/engine/coord"
	"github.com/ghthor/engine/rpg2d/entity"
)

// Represents one of the four quad corners
type Corner int

const (
	NW Corner = iota
	NE
	SE
	SW
)

// An interface used to abstract the implementation differences
// of a node and a leaf.
type Quad interface {
	Insert(entity.Entity) Quad
	Remove(entity.Entity) Quad

	QueryCell(coord.Cell) []entity.Entity
	QueryBounds(coord.AABB) []entity.Entity

	Bounds() coord.AABB

	Parent() Quad
	Child(Corner) Quad
	Children() []Quad
}

// A node in the quad tree that will contain 4 children,
// one in each corner of the quad nodes bounds.
type quadNode struct {
	parent Quad

	bounds coord.AABB

	children [4]Quad
}

// A leaf in the quad tree that contains the references
// to entities. A leaf represents one corner in the parents
// bounds.
type quadLeaf struct {
	parent Quad

	bounds coord.AABB

	entities    []entity.Entity
	moveables   []entity.Movable
	collidables []entity.Collidable

	maxSize int
}
