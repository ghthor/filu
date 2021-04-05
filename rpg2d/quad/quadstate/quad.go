package quadstate

import (
	"errors"
	"fmt"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
)

type Quad interface {
	Bounds() coord.Bounds

	Insert(*Entity) Quad
	// Remove all State's that have been inserted
	Clear() Quad

	QueryBounds(coord.Bounds, Accumulator, QueryFlag)

	AccumulateAll(Accumulator, QueryFlag)
}

type Accumulator interface {
	Add(*Entity)
	AddSlice([]*Entity, Type)
}

var _ Quad = root{}
var _ Quad = node{}
var _ Quad = leaf{}

type root struct {
	quad Quad
}

type node struct {
	bounds   coord.Bounds
	children [4]Quad
}

type leaf struct {
	bounds  coord.Bounds
	size    int
	maxSize int
	*Entities
}

func newLeaf(bounds coord.Bounds, maxSize int) leaf {
	return leaf{
		bounds:   bounds,
		Entities: NewEntities(maxSize),
		maxSize:  maxSize,
	}
}

func (q root) Bounds() coord.Bounds { return q.quad.Bounds() }
func (q node) Bounds() coord.Bounds { return q.bounds }
func (q leaf) Bounds() coord.Bounds { return q.bounds }

func (q root) Insert(e *Entity) Quad {
	q.quad = q.quad.Insert(e)
	return q
}

func (q node) Insert(e *Entity) Quad {
	for i, quad := range q.children {
		// If the child's bounds contain the entities cell
		if quad.Bounds().Contains(e.EntityCell()) {
			q.children[i] = quad.Insert(e)
			return q
		}
	}
	panic("entity out of bounds")
}

func (q leaf) Insert(e *Entity) Quad {
	// If the quad is full it must split
	if q.size >= q.maxSize {
		// Unless it has a width or height of 1 then it can't be split.
		if q.Bounds().Height() > 1 && q.Bounds().Width() > 1 {
			return q.divide().Insert(e)
		}
	}

	//fmt.Printf("actual size: %d maxSize: %d\n", len(q.entities), q.maxSize)

	q.Entities.Insert(e)
	q.size++

	return q
}

func (q leaf) divide() Quad {
	qn := node{
		bounds: q.bounds,
	}

	quads, err := q.bounds.Quads()
	// TODO Remove the need for this panic
	if err != nil {
		panic(fmt.Sprintf("error splitting bounds into quads: %e", err))
	}

	// TODO Reuse this leaf forming 3 new leaves + this 1
	for i, _ := range qn.children {
		qn.children[i] = newLeaf(quads[i], q.maxSize)
	}

	var quad Quad = qn

	for t := range q.Entities.ByType {
		for _, e := range q.Entities.ByType[t] {
			quad = quad.Insert(e)
		}
	}

	return quad
}

func (q root) Clear() Quad {
	q.quad = q.quad.Clear()
	return q
}

func (q node) Clear() Quad {
	for i, _ := range q.children {
		q.children[i] = q.children[i].Clear()
	}
	return q
}

func (q leaf) Clear() Quad {
	q.size = 0
	q.Entities.Clear()
	return q
}

func (q root) QueryBounds(bounds coord.Bounds, acc Accumulator, types QueryFlag) {
	q.quad.QueryBounds(bounds, acc, types)
}

func (q node) QueryBounds(bounds coord.Bounds, acc Accumulator, types QueryFlag) {
	intersection, err := q.bounds.Intersection(bounds)
	if err != nil {
		return
	}

	if intersection == q.bounds {
		for i, _ := range q.children {
			q.children[i].AccumulateAll(acc, types)
		}
		return
	}

	for i, _ := range q.children {
		q.children[i].QueryBounds(bounds, acc, types)
	}
}

func (q leaf) QueryBounds(bounds coord.Bounds, acc Accumulator, types QueryFlag) {
	intersection, err := q.bounds.Intersection(bounds)
	if err != nil {
		return
	}

	if intersection == q.bounds {
		q.AccumulateAll(acc, types)
		return
	}

	for i := range allTypes {
		if types&allTypes[i].QueryFlag != 0 {
			t := allTypes[i].Type
			filterBounds(acc, q.Entities.ByType[t], t, bounds)
		}
	}
}

func filterBounds(acc Accumulator, src []*Entity, t Type, bounds coord.Bounds) {
	for i, _ := range src {
		if e, hasBounds := src[i].State.(entity.HasBounds); hasBounds {
			if e.Bounds().Overlaps(bounds) {
				acc.Add(src[i])
			}
			continue
		}

		if bounds.Contains(src[i].EntityCell()) {
			acc.Add(src[i])
		}
	}
}

func (q root) AccumulateAll(acc Accumulator, types QueryFlag) {
	q.quad.AccumulateAll(acc, types)
}

func (q node) AccumulateAll(acc Accumulator, types QueryFlag) {
	for i, _ := range q.children {
		q.children[i].AccumulateAll(acc, types)
	}
}

func (q leaf) AccumulateAll(acc Accumulator, types QueryFlag) {
	for i := range allTypes {
		if types&allTypes[i].QueryFlag != 0 {
			t := allTypes[i].Type
			acc.AddSlice(q.ByType[t], t)
		}
	}
}

// Guards against unspecified behavior if the maxSize is 1
var ErrMaxSizeTooSmall = errors.New("max size must be > 1")
var ErrBoundsHeightMustBePowerOf2 = errors.New("bounds height must be a power of 2")
var ErrBoundsWidthMustBePowerOf2 = errors.New("bounds width must be a power of 2")

func New(bounds coord.Bounds, maxSize int) (Quad, error) {
	if maxSize < 2 {
		return nil, ErrMaxSizeTooSmall
	}

	if !isPowerOf2(uint(bounds.Width())) {
		return nil, ErrBoundsWidthMustBePowerOf2
	}

	if !isPowerOf2(uint(bounds.Height())) {
		return nil, ErrBoundsHeightMustBePowerOf2
	}

	return root{
		quad: newLeaf(bounds, maxSize),
	}, nil
}

func NewMust(bounds coord.Bounds, maxSize int) Quad {
	q, err := New(bounds, maxSize)
	if err != nil {
		panic(err)
	}
	return q
}

func isPowerOf2(v uint) bool {
	return (v != 0) && (v&(v-1)) == 0
}
