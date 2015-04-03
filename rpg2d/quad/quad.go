package quad

import (
	"errors"
	"fmt"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/sim/stime"
)

// Represents one of the four quad corners
type Corner coord.Quad

// An interface used to abstract the implementation differences
// of a node and a leaf.
type Quad interface {
	Parent() Quad
	Child(Corner) Quad
	Children() []Quad

	Bounds() coord.Bounds

	// Mutators
	Insert(entity.Entity) Quad
	Remove(entity.Entity) Quad

	QueryCell(coord.Cell) []entity.Entity
	QueryBounds(coord.Bounds) []entity.Entity

	Chunk() Chunk

	//---- Internal methods to execute a phase calculation
	runUpdatePhase(UpdatePhaseHandler, stime.Time) (quad Quad, remaining, removed []entity.Entity)
	runInputPhase(InputPhaseHandler, stime.Time) (Quad, []entity.Entity)
	runBroadPhase(stime.Time) (cgroups []*CollisionGroup, solved, unsolved CollisionGroupIndex)
}

// Guards against unspecified behavior if the maxSize is 1
var ErrMaxSizeTooSmall = errors.New("max size must be > 1")

var ErrBoundsHeightMustBePowerOf2 = errors.New("bounds height must be a power of 2")
var ErrBoundsWidthMustBePowerOf2 = errors.New("bounds width must be a power of 2")

func isPowerOf2(v uint) bool {
	return (v != 0) && (v&(v-1)) == 0
}

func New(bounds coord.Bounds, maxSize int, entities []entity.Entity) (Quad, error) {
	if maxSize < 2 {
		return nil, ErrMaxSizeTooSmall
	}

	if !isPowerOf2(uint(bounds.Width())) {
		return nil, ErrBoundsWidthMustBePowerOf2
	}

	if !isPowerOf2(uint(bounds.Height())) {
		return nil, ErrBoundsHeightMustBePowerOf2
	}

	return quadRoot{
		Quad: quadLeaf{
			parent:  nil,
			bounds:  bounds,
			maxSize: maxSize,
		},
		entityIndex: make(map[entity.Id]entity.Entity),
	}, nil
}

// A group of entities within a bounding rectangle.
// A chunk is sent to the collision function that is implemented
// by the user of engine.
type Chunk struct {
	Bounds coord.Bounds

	Entities []entity.Entity
}

type quadRoot struct {
	Quad

	entityIndex map[entity.Id]entity.Entity
}

// A node in the quad tree that will contain 4 children,
// one in each corner of the quad nodes bounds.
type quadNode struct {
	parent   Quad
	children [4]Quad

	bounds coord.Bounds
}

func (q quadRoot) Insert(e entity.Entity) Quad {
	old, indexed := q.entityIndex[e.Id()]
	if indexed {
		q.Quad = q.Quad.Remove(old)
	}

	q.entityIndex[e.Id()] = e
	q.Quad = q.Quad.Insert(e)

	return q
}

func (q quadRoot) Remove(e entity.Entity) Quad {
	old, indexed := q.entityIndex[e.Id()]
	if !indexed {
		return q
	}

	q.Quad = q.Quad.Remove(old)
	delete(q.entityIndex, e.Id())

	return q
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
		if quad.Bounds().Contains(c) {
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
	}

	return chunk
}

// A leaf in the quad tree that contains the references
// to entities. A leaf represents one corner in the parents
// bounds.
type quadLeaf struct {
	parent Quad

	bounds coord.Bounds

	entities []entity.Entity

	maxSize int
}

func (q quadLeaf) Parent() Quad        { return q.parent }
func (q quadLeaf) Child(c Corner) Quad { return nil }
func (q quadLeaf) Children() []Quad    { return nil }

func (q quadLeaf) Bounds() coord.Bounds { return q.bounds }

func (q quadLeaf) Insert(e entity.Entity) Quad {
	// If the quad is full it must split
	if len(q.entities) >= q.maxSize {
		// Unless it has a width or height of 1
		// Then is can't split
		if q.Bounds().Height() > 1 && q.Bounds().Width() > 1 {
			return q.divide().Insert(e)
		}
	}

	q.entities = append(q.entities, e)

	//fmt.Printf("actual size: %d maxSize: %d\n", len(q.entities), q.maxSize)

	return q
}

func (q quadLeaf) divide() Quad {
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

			entities: make([]entity.Entity, 0, q.maxSize),

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

	return q
}

func (q quadLeaf) QueryCell(c coord.Cell) []entity.Entity {
	entities := make([]entity.Entity, 0, 1)
	for _, e := range q.entities {
		if e.Bounds().Contains(c) {
			entities = append(entities, e)
		}
	}

	if len(entities) == 0 {
		return nil
	}

	return entities
}

func (q quadLeaf) QueryBounds(b coord.Bounds) []entity.Entity {
	if !q.Bounds().Overlaps(b) {
		return nil
	}

	entities := make([]entity.Entity, 0, q.maxSize)
	for _, e := range q.entities {
		if b.Overlaps(e.Bounds()) {
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

	return chunk
}
