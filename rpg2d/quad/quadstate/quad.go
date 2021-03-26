package quadstate

import (
	"errors"
	"fmt"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
)

type Type uint

const (
	TypeRemoved Type = 1 << iota
	TypeNew
	TypeChanged
	TypeUnchanged
)

type QueryFlag uint

const (
	QueryAll  QueryFlag = QueryFlag(TypeRemoved | TypeNew | TypeChanged | TypeUnchanged)
	QueryDiff QueryFlag = QueryFlag(TypeRemoved | TypeNew | TypeChanged)
)

type Entity struct {
	entity.State
	Type
	// TODO add encoding cache to this type
}

type Quad interface {
	Bounds() coord.Bounds

	Insert(Entity) Quad
	// Remove all State's that have been inserted
	Clear() Quad

	QueryBounds(coord.Bounds, Accumulator, QueryFlag)

	AccumulateAll(Accumulator, QueryFlag)
}

type Accumulator interface {
	Add(Entity, entity.Flag)
	AddSlice([]Entity, entity.Flag)
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
	bounds            coord.Bounds
	entitiesRemoved   []Entity
	entitiesNew       []Entity
	entitiesChanged   []Entity
	entitiesUnchanged []Entity
	size              int
	maxSize           int
}

func newLeaf(bounds coord.Bounds, maxSize int) leaf {
	return leaf{
		bounds:            bounds,
		entitiesRemoved:   make([]Entity, 0, maxSize),
		entitiesNew:       make([]Entity, 0, maxSize),
		entitiesChanged:   make([]Entity, 0, maxSize),
		entitiesUnchanged: make([]Entity, 0, maxSize),
		maxSize:           maxSize,
	}
}

func (q root) Bounds() coord.Bounds { return q.quad.Bounds() }
func (q node) Bounds() coord.Bounds { return q.bounds }
func (q leaf) Bounds() coord.Bounds { return q.bounds }

func (q root) Insert(e Entity) Quad {
	q.quad = q.quad.Insert(e)
	return q
}

func (q node) Insert(e Entity) Quad {
	for i, quad := range q.children {
		// If the child's bounds contain the entities cell
		if quad.Bounds().Contains(e.EntityCell()) {
			q.children[i] = quad.Insert(e)
			return q
		}
	}
	panic("entity out of bounds")
}

func (q leaf) Insert(e Entity) Quad {
	// If the quad is full it must split
	if q.size >= q.maxSize {
		// Unless it has a width or height of 1 then it can't be split.
		if q.Bounds().Height() > 1 && q.Bounds().Width() > 1 {
			return q.divide().Insert(e)
		}
	}

	//fmt.Printf("actual size: %d maxSize: %d\n", len(q.entities), q.maxSize)

	switch {
	case e.Type&TypeRemoved != 0:
		q.entitiesRemoved = append(q.entitiesRemoved, e)
	case e.Type&TypeNew != 0:
		q.entitiesNew = append(q.entitiesNew, e)
	case e.Type&TypeChanged != 0:
		q.entitiesChanged = append(q.entitiesChanged, e)
	case e.Type&TypeUnchanged != 0:
		q.entitiesUnchanged = append(q.entitiesUnchanged, e)
	default:
		return q
	}

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

	for _, e := range q.entitiesRemoved {
		quad = quad.Insert(e)
	}
	for _, e := range q.entitiesNew {
		quad = quad.Insert(e)
	}
	for _, e := range q.entitiesChanged {
		quad = quad.Insert(e)
	}
	for _, e := range q.entitiesUnchanged {
		quad = quad.Insert(e)
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
	q.entitiesRemoved = q.entitiesRemoved[:0]
	q.entitiesNew = q.entitiesNew[:0]
	q.entitiesChanged = q.entitiesChanged[:0]
	q.entitiesUnchanged = q.entitiesUnchanged[:0]
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

	if types&QueryFlag(TypeRemoved) != 0 {
		filterBounds(acc, q.entitiesRemoved, entity.FlagRemoved, bounds)
	}
	if types&QueryFlag(TypeNew) != 0 {
		filterBounds(acc, q.entitiesNew, entity.FlagNew, bounds)
	}
	if types&QueryFlag(TypeChanged) != 0 {
		filterBounds(acc, q.entitiesChanged, entity.FlagChanged, bounds)
	}
	if types&QueryFlag(TypeUnchanged) != 0 {
		filterBounds(acc, q.entitiesUnchanged, 0, bounds)
	}
}

func filterBounds(acc Accumulator, src []Entity, flag entity.Flag, bounds coord.Bounds) {
	for i, _ := range src {
		if e, hasBounds := src[i].State.(entity.HasBounds); hasBounds {
			if e.Bounds().Overlaps(bounds) {
				acc.Add(src[i], flag)
			}
			continue
		}

		if bounds.Contains(src[i].EntityCell()) {
			acc.Add(src[i], flag)
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
	if types&QueryFlag(TypeRemoved) != 0 {
		acc.AddSlice(q.entitiesRemoved, entity.FlagRemoved)
	}
	if types&QueryFlag(TypeNew) != 0 {
		acc.AddSlice(q.entitiesNew, entity.FlagNew)
	}
	if types&QueryFlag(TypeChanged) != 0 {
		acc.AddSlice(q.entitiesChanged, entity.FlagChanged)
	}
	if types&QueryFlag(TypeUnchanged) != 0 {
		acc.AddSlice(q.entitiesUnchanged, 0)
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
