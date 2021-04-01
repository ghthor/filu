package quad_test

import (
	"testing"

	"github.com/ghthor/filu/rpg2d/entity/entitytest"
	"github.com/ghthor/filu/rpg2d/quad"
)

func TestCollisionGroupPool(t *testing.T) {
	pool := &quad.CollisionGroupPool{}

	cg1 := pool.NewGroup()

	pool.Reset()

	cg2 := pool.NewGroup()

	if cg1 != cg2 {
		t.Fail()
	}

	cg3 := pool.NewGroup()

	if cg1 == cg3 {
		t.Fail()
	}

	cg3.AddCollision(entitytest.MockEntity{EntityId: 0}, entitytest.MockEntity{EntityId: 1})

	if len(cg3.CollisionsById) != 1 {
		t.Fail()
	}

	pool.Reset()

	if len(cg3.CollisionsById) != 0 {
		t.Fail()
	}
}
