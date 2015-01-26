package engine

import (
	"fmt"

	. "github.com/ghthor/engine/rpg2d/coord"
	. "github.com/ghthor/engine/time"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type (
	MockEntityJson struct {
		EntityId EntityId `json:"id"`
		Name     string   `json:"name"`
		Cell     Cell     `json:"cell"`
	}

	MockEntity struct {
		id   EntityId
		cell Cell
	}

	MockMobileEntity struct {
		id EntityId
		mi *motionInfo
	}

	MockCollision struct {
		time Time
		A, B collidableEntity
	}

	MockCollidableEntity struct {
		id         EntityId
		cell       Cell
		collisions []MockCollision
	}

	MockAliveEntity struct {
		id         EntityId
		mi         *motionInfo
		collisions []MockCollision
	}
)

func (e MockEntityJson) Id() EntityId { return e.EntityId }
func (e MockEntityJson) AABB() Bounds { return Bounds{e.Cell, e.Cell} }
func (e MockEntityJson) IsDifferentFrom(other EntityJson) bool {
	o := other.(MockEntityJson)
	if e.Name != o.Name {
		return true
	}
	return false
}

func (e MockEntity) Id() EntityId { return e.id }
func (e MockEntity) Cell() Cell   { return e.cell }
func (e MockEntity) AABB() Bounds { return Bounds{e.cell, e.cell} }
func (e MockEntity) Json() EntityJson {
	return MockEntityJson{
		e.Id(),
		e.String(),
		e.cell,
	}
}

func (e MockEntity) String() string {
	return fmt.Sprintf("MockEntity%v", e.Id())
}

func (e *MockMobileEntity) Id() EntityId { return e.id }
func (e *MockMobileEntity) Cell() Cell   { return e.mi.cell }
func (e *MockMobileEntity) AABB() Bounds { return e.mi.AABB() }
func (e *MockMobileEntity) Json() EntityJson {
	return MockEntityJson{
		e.Id(),
		e.String(),
		e.mi.cell,
	}
}

func (e *MockMobileEntity) motionInfo() *motionInfo { return e.mi }

func (e *MockMobileEntity) String() string {
	return fmt.Sprintf("MockMobileEntity%v", e.Id())
}

func (e *MockCollidableEntity) Id() EntityId { return e.id }
func (e *MockCollidableEntity) Cell() Cell   { return e.cell }
func (e *MockCollidableEntity) AABB() Bounds { return Bounds{e.cell, e.cell} }
func (e *MockCollidableEntity) Json() EntityJson {
	return MockEntityJson{
		e.Id(),
		e.String(),
		e.cell,
	}
}

func (e *MockCollidableEntity) collides(other collidableEntity) bool { return true }
func (e *MockCollidableEntity) collideWith(other collidableEntity, t Time) {
	e.collisions = append(e.collisions, MockCollision{t, e, other})
}

func (e *MockCollidableEntity) String() string {
	return fmt.Sprintf("MockCollidableEntity%v", e.Id())
}

func (e *MockAliveEntity) Id() EntityId { return e.id }
func (e *MockAliveEntity) Cell() Cell   { return e.mi.cell }
func (e *MockAliveEntity) AABB() Bounds { return e.mi.AABB() }
func (e *MockAliveEntity) Json() EntityJson {
	return MockEntityJson{
		e.Id(),
		e.String(),
		e.mi.cell,
	}
}

func (e *MockAliveEntity) motionInfo() *motionInfo { return e.mi }

func (e *MockAliveEntity) collides(other collidableEntity) bool { return true }
func (e *MockAliveEntity) collideWith(other collidableEntity, t Time) {
	e.collisions = append(e.collisions, MockCollision{t, e, other})
}

func (e *MockAliveEntity) String() string {
	return fmt.Sprintf("MockAliveEntity%v", e.Id())
}

func CollidedWith(a, b interface{}) (collided bool, pos, neg gospec.Message, err error) {
	var collisionsA, collisionsB []MockCollision

	switch e := a.(type) {
	case *MockCollidableEntity:
		collisionsA = e.collisions
	case *MockAliveEntity:
		collisionsA = e.collisions
	}

	switch e := b.(type) {
	case *MockCollidableEntity:
		collisionsB = e.collisions
	case *MockAliveEntity:
		collisionsB = e.collisions
	}

	var time Time
outer:
	for _, collision1 := range collisionsA {
		for _, collision2 := range collisionsB {
			if collision1.time == collision2.time {
				if collision1.A == collision2.B && collision1.B == collision2.A {
					time = collision1.time
					collided = true
					break outer
				}
			}
		}
	}

	pos = gospec.Messagef(a, "had a collision with %v @%d", b, time)
	neg = gospec.Messagef(a, "did not collide with %v", b)
	return
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

	c.Specify("mock collidable entity", func() {
		e := entity(&MockCollidableEntity{})

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity)
			c.Expect(isAnEntity, IsTrue)
		})

		c.Specify("is not a movable entity", func() {
			_, isAMovableEntity := e.(movableEntity)
			c.Expect(isAMovableEntity, IsFalse)
		})

		c.Specify("is a collidable entity", func() {
			_, isACollidableEntity := e.(collidableEntity)
			c.Expect(isACollidableEntity, IsTrue)
		})
	})

	c.Specify("mock alive entity", func() {
		e := entity(&MockAliveEntity{})

		c.Specify("is an entity", func() {
			_, isAnEntity := e.(entity)
			c.Expect(isAnEntity, IsTrue)
		})

		c.Specify("is a movable entity", func() {
			_, isAMovableEntity := e.(movableEntity)
			c.Expect(isAMovableEntity, IsTrue)
		})

		c.Specify("is a collidable entity", func() {
			_, isACollidableEntity := e.(collidableEntity)
			c.Expect(isACollidableEntity, IsTrue)
		})
	})

	c.Specify("CollidedWith matcher", func() {
		entities := [...]*MockCollidableEntity{
			&MockCollidableEntity{collisions: make([]MockCollision, 0, 2)},
			&MockCollidableEntity{collisions: make([]MockCollision, 0, 2)},
		}

		c.Expect(entities[0], Not(CollidedWith), entities[1])
		c.Expect(entities[1], Not(CollidedWith), entities[0])

		entityCollision{0, entities[0], entities[1]}.collide()

		c.Expect(entities[0], CollidedWith, entities[1])
		c.Expect(entities[1], CollidedWith, entities[0])

		entityCollision{0, entities[0], entities[1]}.collide()

		c.Expect(entities[0], CollidedWith, entities[1])
		c.Expect(entities[1], CollidedWith, entities[0])
	})
}
