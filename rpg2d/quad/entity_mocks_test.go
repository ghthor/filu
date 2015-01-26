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
)

func (e MockEntity) String() string       { return fmt.Sprintf("MockEntity%v", e.Id()) }
func (e MockEntity) Id() int64            { return e.id }
func (e MockEntity) Cell() coord.Cell     { return e.cell }
func (e MockEntity) Bounds() coord.Bounds { return coord.Bounds{e.cell, e.cell} }

func DescribeMockEntities(c gospec.Context) {
	c.Specify("mock entity", func() {
		e := entity.Entity(&MockEntity{})

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity.Entity)
			c.Expect(isAnEntity, IsTrue)
		})
	})
}
