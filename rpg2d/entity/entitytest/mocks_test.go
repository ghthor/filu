package entitytest

import (
	"testing"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeMockEntities(c gospec.Context) {
	c.Specify("mock entity", func() {
		e := entity.Entity(&MockEntity{})

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity.Entity)
			c.Expect(isAnEntity, IsTrue)
		})
	})

	c.Specify("a mock entity with bounds", func() {
		e := entity.Entity(&MockEntityWithBounds{})

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity.Entity)
			c.Expect(isAnEntity, IsTrue)
		})
	})

	c.Specify("a mock entity state", func() {
		c.Specify("can be the same as another state", func() {
			e1 := MockEntity{EntityCell: coord.Cell{0, 0}}
			c.Expect(e1.ToState().IsDifferentFrom(e1.ToState()), IsFalse)
		})

		c.Specify("can be different from another state", func() {
			e1 := MockEntity{EntityCell: coord.Cell{0, 0}}
			e2 := MockEntity{EntityCell: coord.Cell{0, 1}}

			c.Expect(e1.ToState().IsDifferentFrom(e2.ToState()), IsTrue)
			c.Expect(e2.ToState().IsDifferentFrom(e1.ToState()), IsTrue)
		})
	})
}

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeMockEntities)

	gospec.MainGoTest(r, t)
}
