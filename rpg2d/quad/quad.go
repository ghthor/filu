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
	Parent() Quad
	Child(Corner) Quad
	Children() []Quad

	Bounds() coord.Bounds

	Insert(entity.Entity) Quad
	Remove(entity.Entity) Quad

	QueryCell(coord.Cell) []entity.Entity
	QueryBounds(coord.Bounds) []entity.Entity

	Chunk() Chunk
}

func New(bounds coord.Bounds, maxSize int, entities []entity.Entity) (Quad, error) {
	return quadLeaf{
		parent: nil,

		bounds: bounds,
	}, nil
}

// A group of entities within a bounding rectangle.
// A chunk is sent to the collision function that is implemented
// by the user of engine.
type Chunk struct {
	Bounds coord.Bounds

	Entities    []entity.Entity
	Moveables   []entity.Movable
	Collidables []entity.Collidable
}

// A collisionhandler takes a chunk of entities
// and does collision checks for the chunk and modifies
// the entities. All entities in the returned chunk will
// be reinserted into the quad tree to update their location.
type CollisionHandler func(Chunk) Chunk

// A node in the quad tree that will contain 4 children,
// one in each corner of the quad nodes bounds.
type quadNode struct {
	parent   Quad
	children [4]Quad

	bounds coord.Bounds
}

func (q quadNode) Parent() Quad        { return q.parent }
func (q quadNode) Child(c Corner) Quad { return q.children[c] }
func (q quadNode) Children() []Quad    { return q.children[0:] }

func (q quadNode) Bounds() coord.Bounds { return q.bounds }

func (q quadNode) Insert(e entity.Entity) Quad {
	return q
}

func (q quadNode) Remove(e entity.Entity) Quad {
	return q
}

func (q quadNode) QueryCell(coord.Cell) []entity.Entity {
	return nil
}

func (q quadNode) QueryBounds(coord.Bounds) []entity.Entity {
	return nil
}

func (q quadNode) Chunk() Chunk {
	return Chunk{}
}

// A leaf in the quad tree that contains the references
// to entities. A leaf represents one corner in the parents
// bounds.
type quadLeaf struct {
	parent Quad

	bounds coord.Bounds

	entities    []entity.Entity
	moveables   []entity.Movable
	collidables []entity.Collidable

	maxSize int
}

func (q quadLeaf) Parent() Quad        { return q.parent }
func (q quadLeaf) Child(c Corner) Quad { return nil }
func (q quadLeaf) Children() []Quad    { return nil }

func (q quadLeaf) Bounds() coord.Bounds { return q.bounds }

func (q quadLeaf) Insert(e entity.Entity) Quad {
	return q
}

func (q quadLeaf) Remove(e entity.Entity) Quad {
	return q
}

func (q quadLeaf) QueryCell(coord.Cell) []entity.Entity {
	return nil
}

func (q quadLeaf) QueryBounds(coord.Bounds) []entity.Entity {
	return nil
}

func (q quadLeaf) Chunk() Chunk {
	return Chunk{}
}
