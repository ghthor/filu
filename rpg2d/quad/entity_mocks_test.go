package quad_test

import (
	"fmt"

	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/rpg2d/entity"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type (
	MockEntity struct {
		id   int64
		cell coord.Cell
	}

	MockMobileEntity struct {
		id   int64
		cell coord.Cell
	}

	MockCollidableEntity struct {
		id   int64
		cell coord.Cell
	}
)

func (e MockEntity) String() string       { return fmt.Sprintf("MockEntity%v", e.Id()) }
func (e MockEntity) Id() int64            { return e.id }
func (e MockEntity) Cell() coord.Cell     { return e.cell }
func (e MockEntity) Bounds() coord.Bounds { return coord.Bounds{e.cell, e.cell} }

func (e MockMobileEntity) String() string       { return fmt.Sprintf("MockMobileEntity%v", e.Id()) }
func (e MockMobileEntity) Id() int64            { return e.id }
func (e MockMobileEntity) Cell() coord.Cell     { return e.cell }
func (e MockMobileEntity) Bounds() coord.Bounds { return coord.Bounds{e.cell, e.cell} }

func (e MockMobileEntity) MotionState() entity.MotionState { return entity.MotionState{} }

func (e MockCollidableEntity) String() string       { return fmt.Sprintf("MockCollidableEntity%v", e.Id()) }
func (e MockCollidableEntity) Id() int64            { return e.id }
func (e MockCollidableEntity) Cell() coord.Cell     { return e.cell }
func (e MockCollidableEntity) Bounds() coord.Bounds { return coord.Bounds{e.cell, e.cell} }

func (e MockCollidableEntity) CanCollideWith(other entity.Collidable) bool { return true }
func (e MockCollidableEntity) HadCollisionWith(other entity.Collidable)    {}

func DescribeMockEntities(c gospec.Context) {
	c.Specify("mock entity", func() {
		e := entity.Entity(&MockEntity{})

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity.Entity)
			c.Expect(isAnEntity, IsTrue)
		})

		c.Specify("is not a movable entity", func() {
			_, isAMovableEntity := e.(entity.Movable)
			c.Expect(isAMovableEntity, IsFalse)
		})

		c.Specify("is not a collidable entity", func() {
			_, isACollidableEntity := e.(entity.Collidable)
			c.Expect(isACollidableEntity, IsFalse)
		})
	})

	c.Specify("mock movable entity", func() {
		e := entity.Entity(&MockMobileEntity{})

		var _ entity.Movable = &MockMobileEntity{}

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity.Entity)
			c.Expect(isAnEntity, IsTrue)
		})

		c.Specify("is a movable entity", func() {
			_, isAMovableEntity := e.(entity.Movable)
			c.Expect(isAMovableEntity, IsTrue)
		})

		c.Specify("is not a collidable entity", func() {
			_, isACollidableEntity := e.(entity.Collidable)
			c.Expect(isACollidableEntity, IsFalse)
		})
	})

	c.Specify("mock collidable entity", func() {
		e := entity.Entity(&MockCollidableEntity{})

		var _ entity.Collidable = &MockCollidableEntity{}

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity.Entity)
			c.Expect(isAnEntity, IsTrue)
		})

		c.Specify("is not a movable entity", func() {
			_, isAMovableEntity := e.(entity.Movable)
			c.Expect(isAMovableEntity, IsFalse)
		})

		c.Specify("is a collidable entity", func() {
			_, isACollidableEntity := e.(entity.Collidable)
			c.Expect(isACollidableEntity, IsTrue)
		})
	})
}
