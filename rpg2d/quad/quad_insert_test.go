package quad_test

import (
	"github.com/ghthor/engine/coord"
	"github.com/ghthor/engine/rpg2d/quad"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeQuadInsert(c gospec.Context) {
	c.Specify("a quad", func() {
		q, err := quad.New(coord.Bounds{
			TopL: coord.Cell{-10, 10},
			BotR: coord.Cell{10, -10},
		}, 1, nil)
		c.Assume(err, IsNil)

		c.Specify("can insert", func() {
			c.Specify("an entity", func() {
				entity := MockEntity{}
				q = q.Insert(entity)
				chunk := q.Chunk()

				c.Expect(len(chunk.Entities), Equals, 1)
				c.Expect(len(chunk.Moveables), Equals, 0)
				c.Expect(len(chunk.Collidables), Equals, 0)

				c.Specify("and remove it", func() {
					q = q.Remove(entity)
					c.Expect(len(q.Chunk().Entities), Equals, 0)
				})
			})

			c.Specify("a movableEntity", func() {
				entity := &MockMobileEntity{}
				q = q.Insert(entity)
				chunk := q.Chunk()

				c.Expect(len(chunk.Entities), Equals, 1)
				c.Expect(len(chunk.Moveables), Equals, 1)
				c.Expect(len(chunk.Collidables), Equals, 0)

				c.Specify("and remove it", func() {
					q = q.Remove(entity)
					chunk = q.Chunk()

					c.Expect(len(chunk.Entities), Equals, 0)
					c.Expect(len(chunk.Moveables), Equals, 0)
				})
			})

			c.Specify("a collidable entity", func() {
				entity := &MockCollidableEntity{}
				q = q.Insert(entity)
				chunk := q.Chunk()

				c.Expect(len(chunk.Entities), Equals, 1)
				c.Expect(len(chunk.Moveables), Equals, 0)
				c.Expect(len(chunk.Collidables), Equals, 1)

				c.Specify("and remove it", func() {
					q = q.Remove(entity)
					chunk = q.Chunk()

					c.Expect(len(chunk.Entities), Equals, 0)
					c.Expect(len(chunk.Collidables), Equals, 0)
				})
			})

			c.Specify("some entities", func() {
				e1 := MockEntity{0, coord.Cell{-10, 10}}
				e3 := MockCollidableEntity{1, coord.Cell{10, -10}}
				e2 := MockMobileEntity{2, coord.Cell{5, -1}}

				q = q.Insert(e1)
				q = q.Insert(e2)
				q = q.Insert(e3)
				c.Assume(len(q.QueryBounds(q.Bounds())), Equals, 3)

				c.Specify("and remove them", func() {
					q = q.Remove(e1)

					c.Expect(len(q.QueryBounds(q.Bounds())), Equals, 2)
					c.Expect(q.QueryCell(coord.Cell{10, -10})[0].Id(), Equals, int64(1))
					c.Expect(q.QueryCell(coord.Cell{5, -1})[0].Id(), Equals, int64(2))
				})
			})
		})

	})
}
