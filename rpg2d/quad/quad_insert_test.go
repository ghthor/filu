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
			TopL: coord.Cell{-1000, 1000},
			BotR: coord.Cell{1000, -1000},
		}, 10, nil)
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
		})

	})
}
