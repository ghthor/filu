package entitytest

import (
	"fmt"

	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
)

type (
	MockEntity struct {
		EntityId   int64
		EntityCell coord.Cell
	}

	MockEntityWithBounds struct {
		EntityId     int64
		EntityCell   coord.Cell
		EntityBounds coord.Bounds
	}

	MockEntityState struct {
		EntityId int64  `json:"id"`
		Name     string `json:"name"`

		EntityCell coord.Cell `json:"cell"`
		bounds     coord.Bounds
	}
)

func (e MockEntity) String() string       { return fmt.Sprintf("MockEntity%v", e.Id()) }
func (e MockEntity) Id() int64            { return e.EntityId }
func (e MockEntity) Cell() coord.Cell     { return e.EntityCell }
func (e MockEntity) Bounds() coord.Bounds { return coord.Bounds{e.EntityCell, e.EntityCell} }
func (e MockEntity) ToState() entity.State {
	return MockEntityState{
		EntityId:   e.EntityId,
		EntityCell: e.EntityCell,
		Name:       e.String(),
		bounds: coord.Bounds{
			e.EntityCell,
			e.EntityCell,
		},
	}
}

func (e MockEntityWithBounds) String() string       { return fmt.Sprintf("MockEntityWithBounds%v", e.Id()) }
func (e MockEntityWithBounds) Id() int64            { return e.EntityId }
func (e MockEntityWithBounds) Cell() coord.Cell     { return e.EntityCell }
func (e MockEntityWithBounds) Bounds() coord.Bounds { return e.EntityBounds }
func (e MockEntityWithBounds) ToState() entity.State {
	return MockEntityState{
		EntityId:   e.EntityId,
		EntityCell: e.EntityCell,
		Name:       e.String(),
	}
}

func (e MockEntityState) Id() int64            { return e.EntityId }
func (e MockEntityState) Bounds() coord.Bounds { return e.bounds }
func (e MockEntityState) IsDifferentFrom(other entity.State) bool {
	switch other := other.(type) {
	case MockEntityState:
		if e.EntityId != other.EntityId {
			return true
		}

		if e.EntityCell != other.EntityCell {
			return true
		}

		if e.bounds != other.bounds {
			return true
		}

		return false
	}
	panic(fmt.Sprintf("invalid entity comparision {%v to %v}", e, other))
}
