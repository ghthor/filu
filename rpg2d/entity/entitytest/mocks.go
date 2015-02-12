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
)

func (e MockEntity) String() string                    { return fmt.Sprintf("MockEntity%v", e.Id()) }
func (e MockEntity) Id() int64                         { return e.EntityId }
func (e MockEntity) Cell() coord.Cell                  { return e.EntityCell }
func (e MockEntity) Bounds() coord.Bounds              { return coord.Bounds{e.EntityCell, e.EntityCell} }
func (e MockEntity) ToState() entity.State             { return e }
func (e MockEntity) IsDifferentFrom(entity.State) bool { return true }

func (e MockEntityWithBounds) String() string                    { return fmt.Sprintf("MockEntityWithBounds%v", e.Id()) }
func (e MockEntityWithBounds) Id() int64                         { return e.EntityId }
func (e MockEntityWithBounds) Cell() coord.Cell                  { return e.EntityCell }
func (e MockEntityWithBounds) Bounds() coord.Bounds              { return e.EntityBounds }
func (e MockEntityWithBounds) ToState() entity.State             { return e }
func (e MockEntityWithBounds) IsDifferentFrom(entity.State) bool { return true }
