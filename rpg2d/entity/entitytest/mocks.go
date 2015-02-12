package entitytest

import (
	"fmt"

	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"
)

type (
	MockEntity struct {
		id   int64
		cell coord.Cell
	}

	MockEntityWithBounds struct {
		id     int64
		cell   coord.Cell
		bounds coord.Bounds
	}
)

func (e MockEntity) String() string                    { return fmt.Sprintf("MockEntity%v", e.Id()) }
func (e MockEntity) Id() int64                         { return e.id }
func (e MockEntity) Cell() coord.Cell                  { return e.cell }
func (e MockEntity) Bounds() coord.Bounds              { return coord.Bounds{e.cell, e.cell} }
func (e MockEntity) ToState() entity.State             { return e }
func (e MockEntity) IsDifferentFrom(entity.State) bool { return true }

func (e MockEntityWithBounds) String() string                    { return fmt.Sprintf("MockEntityWithBounds%v", e.Id()) }
func (e MockEntityWithBounds) Id() int64                         { return e.id }
func (e MockEntityWithBounds) Cell() coord.Cell                  { return e.cell }
func (e MockEntityWithBounds) Bounds() coord.Bounds              { return e.bounds }
func (e MockEntityWithBounds) ToState() entity.State             { return e }
func (e MockEntityWithBounds) IsDifferentFrom(entity.State) bool { return true }
