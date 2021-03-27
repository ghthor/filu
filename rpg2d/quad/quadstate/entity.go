package quadstate

import (
	"fmt"

	"github.com/ghthor/filu/rpg2d/entity"
)

type Type uint

const (
	TypeRemoved Type = 1 << iota
	TypeNew
	TypeChanged
	TypeUnchanged
)

type Entity struct {
	entity.State
	Type
	// TODO add encoding cache to this type
}

type Entities struct {
	Removed,
	New,
	Unchanged,
	Changed []Entity
}

func NewEntities(size int) *Entities {
	return &Entities{
		Removed:   make([]Entity, 0, size),
		New:       make([]Entity, 0, size),
		Changed:   make([]Entity, 0, size),
		Unchanged: make([]Entity, 0, size),
	}
}

func (entities *Entities) Insert(e Entity) {
	switch {
	case e.Type&TypeRemoved != 0:
		entities.Removed = append(entities.Removed, e)
	case e.Type&TypeNew != 0:
		entities.New = append(entities.New, e)
	case e.Type&TypeChanged != 0:
		entities.Changed = append(entities.Changed, e)
	case e.Type&TypeUnchanged != 0:
		entities.Unchanged = append(entities.Unchanged, e)
	default:
		panic(fmt.Sprintf("unknown entity state type %#v", e))
	}
}

func (e *Entities) Clear() {
	e.Removed =
		e.Removed[:0]
	e.New =
		e.New[:0]
	e.Changed =
		e.Changed[:0]
	e.Unchanged =
		e.Unchanged[:0]
}

var _ Accumulator = &Entities{}

func (entities *Entities) Add(e Entity) {
	switch {
	case e.Type&TypeRemoved != 0:
		entities.Removed = append(entities.Removed, e)
	case e.Type&TypeNew != 0:
		entities.New = append(entities.New, e)
	case e.Type&TypeChanged != 0:
		entities.Changed = append(entities.Changed, e)
	case e.Type&TypeUnchanged != 0:
		entities.Unchanged = append(entities.Unchanged, e)
	default:
		panic(fmt.Sprintf("unknown entity state type %#v", e))
	}
}

func (entities *Entities) AddSlice(others []Entity, t Type) {
	switch {
	case t&TypeRemoved != 0:
		entities.Removed = append(entities.Removed, others...)
	case t&TypeNew != 0:
		entities.New = append(entities.New, others...)
	case t&TypeChanged != 0:
		entities.Changed = append(entities.Changed, others...)
	case t&TypeUnchanged != 0:
		entities.Unchanged = append(entities.Unchanged, others...)
	default:
		panic(fmt.Sprintf("unknown entity state type %#v", t))
	}
}
