package quad_test

import (
	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/entity/entitytest"
	"github.com/ghthor/filu/rpg2d/quad"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeQuad(c gospec.Context) {
	c.Specify("a quad tree", func() {
		q, err := quad.New(coord.Bounds{
			TopL: coord.Cell{-1024, 1024},
			BotR: coord.Cell{1023, -1023},
		}, 2, nil)
		c.Assume(err, IsNil)

		c.Specify("must have a bounds with", func() {
			green := []int{2, 4, 8, 16, 32, 64, 128, 256}
			red := []int{6, 10, 12, 20, 24, 28, 36}

			c.Specify("a height that is a power of 2", func() {
				makeQuad := func(height int) error {
					_, err := quad.New(coord.Bounds{
						coord.Cell{0, height - 1},
						coord.Cell{31, 0},
					}, 2, nil)
					return err
				}

				for _, v := range green {
					c.Expect(makeQuad(v), IsNil)
				}

				for _, v := range red {
					c.Expect(makeQuad(v), Equals, quad.ErrBoundsHeightMustBePowerOf2)
				}
			})

			c.Specify("width that is a power of 2", func() {
				makeQuad := func(width int) error {
					_, err := quad.New(coord.Bounds{
						coord.Cell{0, 31},
						coord.Cell{width - 1, 0},
					}, 2, nil)
					return err
				}
				for _, v := range green {
					c.Expect(makeQuad(v), IsNil)
				}

				for _, v := range red {
					c.Expect(makeQuad(v), Equals, quad.ErrBoundsWidthMustBePowerOf2)
				}
			})
		})

		c.Specify("will subdivide", func() {
			c.Assume(len(q.Children()), Equals, 0)
			q = q.Insert(entitytest.MockEntity{0, coord.Cell{0, 0}, 0})
			q = q.Insert(entitytest.MockEntity{1, coord.Cell{5, 5}, 0})
			q = q.Insert(entitytest.MockEntity{3, coord.Cell{6, 6}, 0})
			c.Expect(len(q.Children()), Equals, 4)
		})

		c.Specify("can remove an entity", func() {
			e := entitytest.MockEntity{0, coord.Cell{0, 0}, 0}

			c.Assume(len(q.QueryCell(e.Cell())), Equals, 0)

			q = q.Insert(e)
			c.Assume(len(q.QueryCell(e.Cell())), Equals, 1)

			q = q.Remove(e)
			c.Expect(len(q.QueryCell(e.Cell())), Equals, 0)
		})

		c.Specify("can be queried by cell", func() {
			for maxSize := 2; maxSize < 8*8; maxSize++ {
				q, err := quad.New(coord.Bounds{
					coord.Cell{-4, 4},
					coord.Cell{3, -3},
				}, maxSize, nil)
				c.Assume(err, IsNil)
				c.Assume(q.Bounds().Width(), Equals, 8)
				c.Assume(q.Bounds().Height(), Equals, 8)

				id := entity.Id(0)
				for j := 4; j > -4; j-- {
					for i := -4; i < 4; i++ {
						q = q.Insert(entitytest.MockEntity{id, coord.Cell{i, j}, 0})
						id++
					}
				}

				id = 0
				for j := 4; j > -4; j-- {
					for i := -4; i < 4; i++ {
						entities := q.QueryCell(coord.Cell{i, j})
						c.Assume(len(entities), Equals, 1)

						c.Expect(entities[0].Id(), Equals, id)
						id++
					}
				}

				c.Assume(len(q.QueryBounds(q.Bounds())), Equals, 8*8)

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
