package entitytest

import (
	"fmt"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
)

type (
	MockEntity struct {
		EntityId   entity.Id
		EntityCell coord.Cell
	}

	MockEntityWithBounds struct {
		EntityId     entity.Id
		EntityCell   coord.Cell
		EntityBounds coord.Bounds
	}

	MockEntityState struct {
		Id   entity.Id `json:"id"`
		Name string    `json:"name"`

		Cell   coord.Cell `json:"cell"`
		bounds coord.Bounds
	}
)

func (e MockEntity) String() string       { return fmt.Sprintf("MockEntity%v", e.Id()) }
func (e MockEntity) Id() entity.Id        { return e.EntityId }
func (e MockEntity) Cell() coord.Cell     { return e.EntityCell }
func (e MockEntity) Bounds() coord.Bounds { return coord.Bounds{e.EntityCell, e.EntityCell} }
func (e MockEntity) ToState() entity.State {
	return MockEntityState{
		Id:   e.EntityId,
		Cell: e.EntityCell,
		Name: e.String(),
		bounds: coord.Bounds{
			e.EntityCell,
			e.EntityCell,
		},
	}
}

func (e MockEntityWithBounds) String() string       { return fmt.Sprintf("MockEntityWithBounds%v", e.Id()) }
func (e MockEntityWithBounds) Id() entity.Id        { return e.EntityId }
func (e MockEntityWithBounds) Cell() coord.Cell     { return e.EntityCell }
func (e MockEntityWithBounds) Bounds() coord.Bounds { return e.EntityBounds }
func (e MockEntityWithBounds) ToState() entity.State {
	return MockEntityState{
		Id:   e.EntityId,
		Cell: e.EntityCell,
		Name: e.String(),
	}
}

func (e MockEntityState) EntityId() entity.Id  { return e.Id }
func (e MockEntityState) Bounds() coord.Bounds { return e.bounds }
func (e MockEntityState) IsDifferentFrom(other entity.State) bool {
	switch other := other.(type) {
	case MockEntityState:
		if e.Id != other.Id {
			return true
		}

		if e.Cell != other.Cell {
			return true
		}

		if e.bounds != other.bounds {
			return true
		}

		return false
	}
	panic(fmt.Sprintf("invalid entity comparision {%v to %v}", e, other))
}
