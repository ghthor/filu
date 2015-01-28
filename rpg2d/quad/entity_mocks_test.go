package quad_test

import (
	"fmt"

	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
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

func (e MockEntity) String() string       { return fmt.Sprintf("MockEntity%v", e.Id()) }
func (e MockEntity) Id() int64            { return e.id }
func (e MockEntity) Cell() coord.Cell     { return e.cell }
func (e MockEntity) Bounds() coord.Bounds { return coord.Bounds{e.cell, e.cell} }

func (e MockEntityWithBounds) String() string       { return fmt.Sprintf("MockEntityWithBounds%v", e.Id()) }
func (e MockEntityWithBounds) Id() int64            { return e.id }
func (e MockEntityWithBounds) Cell() coord.Cell     { return e.cell }
func (e MockEntityWithBounds) Bounds() coord.Bounds { return e.bounds }

func DescribeMockEntities(c gospec.Context) {
	c.Specify("mock entity", func() {
		e := entity.Entity(&MockEntity{})

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity.Entity)
			c.Expect(isAnEntity, IsTrue)
		})
	})

	c.Specify("a mock entity with boudns", func() {
		e := entity.Entity(&MockEntityWithBounds{})

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity.Entity)
			c.Expect(isAnEntity, IsTrue)
		})
	})
}
