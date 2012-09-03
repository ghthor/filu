package engine

import (
	"fmt"
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
)

type (
	MockEntity struct {
		id    EntityId
		coord WorldCoord
	}

	MockMobileEntity struct {
		id EntityId
		mi *motionInfo
	}
)

func (e MockEntity) Id() EntityId      { return e.id }
func (e MockEntity) Coord() WorldCoord { return e.coord }
func (e MockEntity) AABB() AABB        { return AABB{e.coord, e.coord} }
func (e MockEntity) Json() interface{} {
	return struct {
		Id   EntityId `json:"id"`
		Name string   `json:"name"`
	}{
		e.Id(),
		e.String(),
	}
}

func (e MockEntity) String() string {
	return fmt.Sprintf("MockEntity%v", e.Id())
}

func (e *MockMobileEntity) Id() EntityId      { return e.id }
func (e *MockMobileEntity) Coord() WorldCoord { return e.mi.coord }
func (e *MockMobileEntity) AABB() AABB        { return e.mi.AABB() }
func (e *MockMobileEntity) Json() interface{} {
	return struct {
		Id   EntityId `json:"id"`
		Name string   `json:"name"`
	}{
		e.Id(),
		e.String(),
	}
}

func (e *MockMobileEntity) motionInfo() *motionInfo { return e.mi }

func (e *MockMobileEntity) String() string {
	return fmt.Sprintf("MockMobileEntity%v", e.Id())
}

func DescribeMockEntities(c gospec.Context) {
	c.Specify("mock entity", func() {
		e := entity(&MockEntity{})

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity)
			c.Expect(isAnEntity, IsTrue)
		})

		c.Specify("is not a movable entity", func() {
			_, isAMovableEntity := e.(movableEntity)
			c.Expect(isAMovableEntity, IsFalse)
		})

		c.Specify("is not a collidable entity", func() {
			_, isACollidableEntity := e.(collidableEntity)
			c.Expect(isACollidableEntity, IsFalse)
		})
	})

	c.Specify("mock movable entity", func() {
		e := entity(&MockMobileEntity{})

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity)
			c.Expect(isAnEntity, IsTrue)
		})

		c.Specify("is a movable entity", func() {
			_, isAMovableEntity := e.(movableEntity)
			c.Expect(isAMovableEntity, IsTrue)
		})

		c.Specify("is not a collidable entity", func() {
			_, isACollidableEntity := e.(collidableEntity)
			c.Expect(isACollidableEntity, IsFalse)
		})
	})
}
