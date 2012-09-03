package engine

import (
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
)

func DescribeMovableEntity(c gospec.Context) {
	c.Specify("a moving entity", func() {
		motionInfo := newMotionInfo(WorldCoord{1, 1}, North, 40)

		c.Specify("knows when it is moving", func() {
			c.Expect(motionInfo.isMoving(), IsFalse)

			motionInfo.pathActions = append(motionInfo.pathActions, &PathAction{
				NewTimeSpan(0, 20),
				motionInfo.coord,
				motionInfo.coord.Neighbor(North),
			})
			c.Expect(motionInfo.isMoving(), IsTrue)
		})

		c.Specify("can describe its bounding box", func() {
			c.Specify("when it isn't moving", func() {
				c.Assume(motionInfo.isMoving(), IsFalse)
				aabb := motionInfo.AABB()
				c.Expect(aabb, Equals, AABB{WorldCoord{1, 1}, WorldCoord{1, 1}})
			})

			c.Specify("when moving north", func() {
				motionInfo.pathActions = append(motionInfo.pathActions, &PathAction{
					NewTimeSpan(0, 20),
					motionInfo.coord,
					motionInfo.coord.Neighbor(North),
				})

				c.Assume(motionInfo.isMoving(), IsTrue)
				aabb := motionInfo.AABB()
				c.Expect(aabb, Equals, AABB{WorldCoord{1, 2}, WorldCoord{1, 1}})
			})

			c.Specify("when moving east", func() {
				motionInfo.pathActions = append(motionInfo.pathActions, &PathAction{
					NewTimeSpan(0, 20),
					motionInfo.coord,
					motionInfo.coord.Neighbor(East),
				})

				c.Assume(motionInfo.isMoving(), IsTrue)
				aabb := motionInfo.AABB()
				c.Expect(aabb, Equals, AABB{WorldCoord{1, 1}, WorldCoord{2, 1}})
			})

			c.Specify("when moving south", func() {
				motionInfo.pathActions = append(motionInfo.pathActions, &PathAction{
					NewTimeSpan(0, 20),
					motionInfo.coord,
					motionInfo.coord.Neighbor(South),
				})

				c.Assume(motionInfo.isMoving(), IsTrue)
				aabb := motionInfo.AABB()
				c.Expect(aabb, Equals, AABB{WorldCoord{1, 1}, WorldCoord{1, 0}})
			})

			c.Specify("when moving west", func() {
				motionInfo.pathActions = append(motionInfo.pathActions, &PathAction{
					NewTimeSpan(0, 20),
					motionInfo.coord,
					motionInfo.coord.Neighbor(West),
				})

				c.Assume(motionInfo.isMoving(), IsTrue)
				aabb := motionInfo.AABB()
				c.Expect(aabb, Equals, AABB{WorldCoord{0, 1}, WorldCoord{1, 1}})
			})
		})

	})
}
