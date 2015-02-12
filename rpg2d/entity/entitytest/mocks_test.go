package entitytest

import (
	"testing"

	"github.com/ghthor/engine/rpg2d/entity"

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
}

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeMockEntities)

	gospec.MainGoTest(r, t)
}
