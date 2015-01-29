package quad_test

import (
	"github.com/ghthor/engine/rpg2d/quad"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeCollision(c gospec.Context) {
	c.Specify("a collision", func() {
		c.Specify("is the same", func() {
			aptr, bptr := &MockEntity{id: 0}, &MockEntity{id: 1}
			a, b := MockEntity{id: 0}, MockEntity{id: 1}

			c.Specify("if a and b are the same", func() {
				c1, c2 := quad.Collision{aptr, bptr}, quad.Collision{aptr, bptr}
				c.Expect(c1, Equals, c2)
				c.Expect(c2, Equals, c1)

				c1, c2 = quad.Collision{a, b}, quad.Collision{a, b}
				c.Expect(c1, Equals, c2)
				c.Expect(c2, Equals, c1)
			})

			c.Specify("if a and b are swapped", func() {
				c1, c2 := quad.Collision{aptr, bptr}, quad.Collision{bptr, aptr}
				c.Expect(c1, Equals, c2)
				c.Expect(c2, Equals, c1)

				c1, c2 = quad.Collision{a, b}, quad.Collision{b, a}
				c.Expect(c1, Equals, c2)
				c.Expect(c2, Equals, c1)
			})
		})
	})
}
