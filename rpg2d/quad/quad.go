package quad

import (
	"fmt"

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
		parent:  nil,
		bounds:  bounds,
		maxSize: maxSize,
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
	for i, quad := range q.children {
		// If the child's bounds contain the entities cell
		if quad.Bounds().Contains(e.Cell()) {
			q.children[i] = quad.Insert(e)
			return q
		}
	}
	panic("entity out of bounds")
}

func (q quadNode) Remove(e entity.Entity) Quad {
	for i, quad := range q.children {
		// If the child's bounds contain the entities cell
		if quad.Bounds().Contains(e.Cell()) {
			q.children[i] = quad.Remove(e)
			break
		}
	}
	return q
}

func (q quadNode) QueryCell(c coord.Cell) []entity.Entity {
	for _, quad := range q.children {
		// If the cell is within the childs bounds
		if quad.Bounds().Expand(1).Contains(c) {
			return quad.QueryCell(c)
		}
	}

	return nil
}

func (q quadNode) QueryBounds(b coord.Bounds) []entity.Entity {
	var matches []entity.Entity
	for _, quad := range q.children {
		if quad.Bounds().Overlaps(b) {
			matches = append(matches, quad.QueryBounds(b)...)
			// We don't return here in case the bounds overlap
			// with some of the other children
		}
	}
	return matches
}

func (q quadNode) Chunk() Chunk {
	var chunk Chunk = Chunk{Bounds: q.bounds}

	for _, quad := range q.children {
		cchunk := quad.Chunk()
		chunk.Entities = append(chunk.Entities, cchunk.Entities...)
		chunk.Moveables = append(chunk.Moveables, cchunk.Moveables...)
		chunk.Collidables = append(chunk.Collidables, cchunk.Collidables...)
	}

	return chunk
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
	// If the quad is full it must split
	if len(q.entities) >= q.maxSize {
		if q.Bounds().Height() > 1 && q.Bounds().Width() > 1 {
			return q.divide().Insert(e)
		}
	}

	q.entities = append(q.entities, e)

	// Index Movable Entities
	if me, canMove := e.(entity.Movable); canMove {
		q.moveables = append(q.moveables, me)
	}

	// Index collidable Entities
	if ce, canCollide := e.(entity.Collidable); canCollide {
		q.collidables = append(q.collidables, ce)
	}

	return q
}

func (q quadLeaf) divide() Quad {
	// TODO Remove the need for these panics
	if q.bounds.Width() == 1 {
		panic("unable to divide quad with width of 1")
	}

	if q.bounds.Height() == 1 {
		panic("unable to divide quad with height of 1")
	}

	qn := quadNode{
		parent: q.parent,
		bounds: q.bounds,
	}

	quads, err := q.bounds.Quads()

	// TODO Remove the need for this panic
	if err != nil {
		panic(fmt.Sprintf("error splitting bounds into quads: %e", err))
	}

	//TODO Reuse this leaf forming 3 new leaves + this 1
	for i, _ := range qn.children {
		qn.children[i] = quadLeaf{
			parent: qn,
			bounds: quads[i],

			entities:    make([]entity.Entity, 0, q.maxSize),
			moveables:   make([]entity.Movable, 0, q.maxSize),
			collidables: make([]entity.Collidable, 0, q.maxSize),

			maxSize: q.maxSize,
		}
	}

	var quad Quad = qn

	for _, e := range q.entities {
		quad = quad.Insert(e)
	}

	return quad
}

func (q quadLeaf) Remove(remove entity.Entity) Quad {
	// Remove Entity
	for i, entity := range q.entities {
		if entity.Id() == remove.Id() {
			q.entities = append(q.entities[:i], q.entities[i+1:]...)
			break
		}
	}

	// Remove Movable Entity
	if _, canMove := remove.(entity.Movable); canMove {
		for i, me := range q.moveables {
			if me.Id() == remove.Id() {
				q.moveables = append(q.moveables[:i], q.moveables[i+1:]...)
				break
			}
		}
	}

	if _, canCollide := remove.(entity.Collidable); canCollide {
		for i, ce := range q.collidables {
			if ce.Id() == remove.Id() {
				q.collidables = append(q.collidables[:i], q.collidables[i+1:]...)
			}
		}
	}

	return q
}

func (q quadLeaf) QueryCell(c coord.Cell) []entity.Entity {
	entities := make([]entity.Entity, 0, 1)
	for _, e := range q.entities {
		if e.Cell() == c {
			entities = append(entities, e)
		}
	}

	if len(entities) == 0 {
		return nil
	}

	return entities
}

func (q quadLeaf) QueryBounds(b coord.Bounds) []entity.Entity {
	entities := make([]entity.Entity, 0, q.maxSize)
	for _, e := range q.entities {
		if b.Contains(e.Cell()) {
			entities = append(entities, e)
		}
	}

	if len(entities) == 0 {
		return nil
	}

	return entities
}

func (q quadLeaf) Chunk() Chunk {
	chunk := Chunk{Bounds: q.bounds}

	chunk.Entities = append(chunk.Entities, q.entities...)
	chunk.Moveables = append(chunk.Moveables, q.moveables...)
	chunk.Collidables = append(chunk.Collidables, q.collidables...)

	return chunk
}
