package quad_test

import (
	"github.com/ghthor/engine/coord"
	"github.com/ghthor/engine/rpg2d/quad"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeQuad(c gospec.Context) {
	c.Specify("a quad tree", func() {
		q, err := quad.New(coord.Bounds{
			TopL: coord.Cell{-1000, 1000},
			BotR: coord.Cell{1000, -1000},
		}, 1, nil)
		c.Assume(err, IsNil)

		c.Specify("will subdivide", func() {
			c.Assume(len(q.Children()), Equals, 0)
			q = q.Insert(MockEntity{0, coord.Cell{0, 0}})
			q = q.Insert(MockEntity{1, coord.Cell{5, 5}})
			c.Expect(len(q.Children()), Equals, 4)
		})

		c.Specify("can remove an entity", func() {
			e := MockEntity{0, coord.Cell{0, 0}}

			c.Assume(len(q.QueryCell(e.Cell())), Equals, 0)

			q = q.Insert(e)
			c.Assume(len(q.QueryCell(e.Cell())), Equals, 1)

			q = q.Remove(e)
			c.Expect(len(q.QueryCell(e.Cell())), Equals, 0)
		})

		c.Specify("can be queried by cell", func() {
			for maxSize := 1; maxSize < 20*20+1; maxSize++ {
				q, err := quad.New(coord.Bounds{
					coord.Cell{0, 19},
					coord.Cell{-19, 19},
				}, maxSize, nil)
				c.Assume(err, IsNil)

				id := int64(0)
				for j := 0; j > -20; j-- {
					for i := 0; i < 20; i++ {
						q = q.Insert(MockEntity{id, coord.Cell{i, j}})
						id++
					}
				}

				id = 0
				for j := 0; j > -20; j-- {
					for i := 0; i < 20; i++ {
						c.Expect(q.QueryCell(coord.Cell{i, j})[0].Id(), Equals, id)
						id++
					}
				}

				var fn func(q quad.Quad)
				fn = func(q quad.Quad) {
					children := q.Children()
					if len(children) == 0 {
						size := len(q.QueryBounds(q.Bounds()))

						c.Expect(size, Satisfies, size <= maxSize)
					} else {
						for _, child := range children {
							fn(child)
						}
					}
				}

				fn(q)
			}
		})
	})
}
