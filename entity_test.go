package engine

import (
	. "github.com/ghthor/engine/coord"
	. "github.com/ghthor/engine/time"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeMovableEntity(c gospec.Context) {
	c.Specify("a moving entity", func() {
		motionInfo := newMotionInfo(Cell{1, 1}, North, 40)

		c.Specify("knows when it is moving", func() {
			c.Expect(motionInfo.isMoving(), IsFalse)

			motionInfo.pathActions = append(motionInfo.pathActions, &PathAction{
				NewTimeSpan(0, 20),
				motionInfo.cell,
				motionInfo.cell.Neighbor(North),
			})
			c.Expect(motionInfo.isMoving(), IsTrue)
		})

		c.Specify("can describe its bounding box", func() {
			c.Specify("when it isn't moving", func() {
				c.Assume(motionInfo.isMoving(), IsFalse)
				aabb := motionInfo.AABB()
				c.Expect(aabb, Equals, AABB{Cell{1, 1}, Cell{1, 1}})
			})

			c.Specify("when moving north", func() {
				motionInfo.pathActions = append(motionInfo.pathActions, &PathAction{
					NewTimeSpan(0, 20),
					motionInfo.cell,
					motionInfo.cell.Neighbor(North),
				})

				c.Assume(motionInfo.isMoving(), IsTrue)
				aabb := motionInfo.AABB()
				c.Expect(aabb, Equals, AABB{Cell{1, 2}, Cell{1, 1}})
			})

			c.Specify("when moving east", func() {
				motionInfo.pathActions = append(motionInfo.pathActions, &PathAction{
					NewTimeSpan(0, 20),
					motionInfo.cell,
					motionInfo.cell.Neighbor(East),
				})

				c.Assume(motionInfo.isMoving(), IsTrue)
				aabb := motionInfo.AABB()
				c.Expect(aabb, Equals, AABB{Cell{1, 1}, Cell{2, 1}})
			})

			c.Specify("when moving south", func() {
				motionInfo.pathActions = append(motionInfo.pathActions, &PathAction{
					NewTimeSpan(0, 20),
					motionInfo.cell,
					motionInfo.cell.Neighbor(South),
				})

				c.Assume(motionInfo.isMoving(), IsTrue)
				aabb := motionInfo.AABB()
				c.Expect(aabb, Equals, AABB{Cell{1, 1}, Cell{1, 0}})
			})

			c.Specify("when moving west", func() {
				motionInfo.pathActions = append(motionInfo.pathActions, &PathAction{
					NewTimeSpan(0, 20),
					motionInfo.cell,
					motionInfo.cell.Neighbor(West),
				})

				c.Assume(motionInfo.isMoving(), IsTrue)
				aabb := motionInfo.AABB()
				c.Expect(aabb, Equals, AABB{Cell{0, 1}, Cell{1, 1}})
			})
		})

	})
}

func DescribeEntityCollision(c gospec.Context) {
	c.Specify("collisions between the 2 entities", func() {
		entityA := &MockCollidableEntity{id: 0}
		entityB := &MockCollidableEntity{id: 1}

		c.Specify("are the same if they happen at the at the same time", func() {
			c.Expect(entityCollision{0, entityA, entityB}.SameAs(entityCollision{0, entityA, entityB}), IsTrue)
			c.Expect(entityCollision{0, entityA, entityB}.SameAs(entityCollision{0, entityB, entityA}), IsTrue)
		})

		c.Specify("are not the same if they happen at different times", func() {
			c.Expect(entityCollision{0, entityA, entityB}.SameAs(entityCollision{1, entityA, entityB}), IsFalse)
			c.Expect(entityCollision{0, entityA, entityB}.SameAs(entityCollision{1, entityB, entityA}), IsFalse)
		})
	})
}
