package engine

import (
	"fmt"
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
)

func DescribeAABB(c gospec.Context) {
	aabb := AABB{
		WorldCoord{0, 0},
		WorldCoord{0, 0},
	}

	c.Specify("the width, height and area of an aabb", func() {
		c.Expect(aabb.Width(), Equals, 1)
		c.Expect(aabb.Height(), Equals, 1)
		c.Expect(aabb.Area(), Equals, 1)

		aabb = AABB{
			WorldCoord{0, 0},
			WorldCoord{1, -1},
		}
		c.Expect(aabb.Width(), Equals, 2)
		c.Expect(aabb.Height(), Equals, 2)
		c.Expect(aabb.Area(), Equals, 4)
	})

	c.Specify("aabb contains a coord inside of itself", func() {
		c.Expect(aabb.Contains(WorldCoord{0, 0}), IsTrue)
		containsCheck := func(aabb AABB) {
			for i := aabb.TopL.X; i <= aabb.BotR.X; i++ {
				for j := aabb.TopL.Y; j >= aabb.BotR.Y; j-- {
					c.Expect(aabb.Contains(WorldCoord{i, j}), IsTrue)
				}
			}
		}

		containsCheck(AABB{
			WorldCoord{0, 0},
			WorldCoord{1, -1},
		})
		containsCheck(AABB{
			WorldCoord{1, 1},
			WorldCoord{2, -10},
		})
	})

	c.Specify("can identify coords that lay on it's edges", func() {
		edgeCheck := func(aabb AABB) {
			c.Assume(aabb.IsInverted(), IsFalse)

			// Horizontal Edges
			for _, y := range [...]int{aabb.TopL.Y, aabb.BotR.Y} {
				for x := aabb.TopL.X; x <= aabb.BotR.X; x++ {
					c.Expect(aabb.HasOnEdge(WorldCoord{x, y}), IsTrue)
				}
			}

			// Vertical Edges
			for _, x := range [...]int{aabb.TopL.X, aabb.BotR.X} {
				for y := aabb.TopL.Y - 1; y > aabb.BotR.Y; y-- {
					c.Expect(aabb.HasOnEdge(WorldCoord{x, y}), IsTrue)
				}
			}

			outside := AABB{
				aabb.TopL.Add(-1, 1),
				aabb.BotR.Add(1, -1),
			}

			// Horizontal Edges
			for _, y := range [...]int{outside.TopL.Y, outside.BotR.Y} {
				for x := outside.TopL.X; x <= outside.BotR.X; x++ {
					c.Expect(aabb.HasOnEdge(WorldCoord{x, y}), IsFalse)
				}
			}

			// Vertical Edges
			for _, x := range [...]int{outside.TopL.X, outside.BotR.X} {
				for y := outside.TopL.Y - 1; y > outside.BotR.Y; y-- {
					c.Expect(aabb.HasOnEdge(WorldCoord{x, y}), IsFalse)
				}
			}
		}

		edgeCheck(aabb)

		edgeCheck(AABB{
			WorldCoord{0, 0},
			WorldCoord{1, -1},
		})

		edgeCheck(AABB{
			WorldCoord{1, 1},
			WorldCoord{1, -1},
		})

		edgeCheck(AABB{
			WorldCoord{-10, 10},
			WorldCoord{10, -10},
		})

		edgeCheck(AABB{
			WorldCoord{-10, 10},
			WorldCoord{-10, -10},
		})

		edgeCheck(AABB{
			WorldCoord{-10, -10},
			WorldCoord{10, -10},
		})
	})

	c.Specify("can calulate the intersection of 2 AABBs", func() {
		aabb := AABB{
			WorldCoord{0, 0},
			WorldCoord{10, -10},
		}

		c.Specify("when they overlap", func() {
			other := AABB{
				WorldCoord{5, -5},
				WorldCoord{15, -15},
			}

			intersection := AABB{
				WorldCoord{5, -5},
				WorldCoord{10, -10},
			}

			intersectionResult, err := aabb.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)

			intersectionResult, err = other.Intersection(aabb)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)

			other = AABB{
				WorldCoord{-5, 5},
				WorldCoord{5, -5},
			}

			intersection = AABB{
				WorldCoord{0, 0},
				WorldCoord{5, -5},
			}

			intersectionResult, err = aabb.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)

			intersectionResult, err = other.Intersection(aabb)
			c.Assume(err, IsNil)
			c.Expect(intersectionResult, Equals, intersection)
		})

		c.Specify("when one is contained inside the other", func() {
			// aabb Contains other
			other := AABB{
				WorldCoord{5, -5},
				WorldCoord{6, -6},
			}

			intersection, err := aabb.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, other)

			intersection, err = other.Intersection(aabb)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, other)

			// other Contains aabb
			other = AABB{
				WorldCoord{-1, 1},
				WorldCoord{11, -11},
			}

			intersection, err = aabb.Intersection(other)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, aabb)

			intersection, err = other.Intersection(aabb)
			c.Assume(err, IsNil)
			c.Expect(intersection, Equals, aabb)
		})

		c.Specify("and an error is returned if the rectangles do not overlap", func() {
			other := AABB{
				WorldCoord{11, -11},
				WorldCoord{11, -11},
			}

			_, err := aabb.Intersection(other)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "no overlap")
		})
	})

	c.Specify("flip topleft and bottomright if they are inverted", func() {
		aabb = AABB{
			WorldCoord{0, 0},
			WorldCoord{-1, 1},
		}

		c.Expect(aabb.IsInverted(), IsTrue)
		c.Expect(aabb.Invert().IsInverted(), IsFalse)
	})
}

type MockMobileEntity struct {
	id EntityId
	mi *motionInfo
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

func DescribeQuad(c gospec.Context) {
	c.Specify("AABB can be split into 4 quads", func() {
		aabb := AABB{
			WorldCoord{0, 0},
			WorldCoord{10, -9},
		}

		quads := [4]AABB{{
			WorldCoord{0, 0},
			WorldCoord{4, -4},
		}, {
			WorldCoord{5, 0},
			WorldCoord{10, -4},
		}, {
			WorldCoord{5, -5},
			WorldCoord{10, -9},
		}, {
			WorldCoord{0, -5},
			WorldCoord{4, -9},
		}}

		quadsResult, err := splitAABBToQuads(aabb)
		c.Assume(err, IsNil)

		for i, quad := range quadsResult {
			c.Expect(quad, Equals, quads[i])
		}

		// Width == Height == 2
		aabb = AABB{
			WorldCoord{2, -2},
			WorldCoord{3, -3},
		}

		quads = [4]AABB{{
			WorldCoord{2, -2},
			WorldCoord{2, -2},
		}, {
			WorldCoord{3, -2},
			WorldCoord{3, -2},
		}, {
			WorldCoord{3, -3},
			WorldCoord{3, -3},
		}, {
			WorldCoord{2, -3},
			WorldCoord{2, -3},
		}}

		quadsResult, err = splitAABBToQuads(aabb)
		c.Assume(err, IsNil)

		for i, quad := range quadsResult {
			c.Expect(quad, Equals, quads[i])
		}

		c.Specify("only if the height is greater than 1", func() {
			aabb = AABB{
				WorldCoord{1, 1},
				WorldCoord{2, 1},
			}

			_, err := splitAABBToQuads(aabb)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "aabb is too small to split")
		})

		c.Specify("only if the width is greater than 1", func() {
			aabb = AABB{
				WorldCoord{1, 1},
				WorldCoord{1, 0},
			}

			_, err := splitAABBToQuads(aabb)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "aabb is too small to split")
		})

		c.Specify("only if it isn't inverted", func() {
			aabb = AABB{
				WorldCoord{0, 0},
				WorldCoord{-1, 1},
			}

			_, err := splitAABBToQuads(aabb)
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "aabb is inverted")
		})
	})

	c.Specify("quad can be queried with an AABB to return all entities that are inside", func() {
		qt, err := newQuadTree(AABB{
			WorldCoord{-1000, 1000},
			WorldCoord{1000, -1000},
		}, nil, 1)
		c.Assume(err, IsNil)

		qt = qt.Insert(MockEntity{0, WorldCoord{0, 0}})
		qt = qt.Insert(MockEntity{1, WorldCoord{10, 10}})
		qt = qt.Insert(MockEntity{2, WorldCoord{-10, -10}})
		qt = qt.Insert(MockEntity{3, WorldCoord{999, -1000}})

		entities := qt.QueryAll(AABB{
			WorldCoord{0, 0},
			WorldCoord{10, -10},
		})
		c.Expect(len(entities), Equals, 1)

		entities = qt.QueryAll(AABB{
			WorldCoord{-90, 90},
			WorldCoord{-1, -9},
		})
		c.Expect(len(entities), Equals, 0)

		entities = qt.QueryAll(qt.AABB())
		c.Expect(len(entities), Equals, 4)
	})

	c.Specify("entity is inserted into quadtree", func() {
		qt, err := newQuadTree(AABB{
			WorldCoord{-1000, 1000},
			WorldCoord{1000, -1000},
		}, nil, 1)
		c.Assume(err, IsNil)

		c.Specify("as an entity", func() {
			qt = qt.Insert(MockEntity{})

			leaf, isAQuadLeaf := qt.(*quadLeaf)
			c.Assume(isAQuadLeaf, IsTrue)

			c.Expect(len(leaf.entities), Equals, 1)
			c.Expect(len(leaf.movableEntities), Equals, 0)

			c.Specify("and removed", func() {
				qt.Remove(MockEntity{})
				c.Expect(len(leaf.entities), Equals, 0)
			})
		})

		c.Specify("as an movableEntity", func() {
			qt = qt.Insert(&MockMobileEntity{mi: newMotionInfo(WorldCoord{0, 0}, North, 20)})

			leaf, isAQuadLeaf := qt.(*quadLeaf)
			c.Assume(isAQuadLeaf, IsTrue)

			c.Expect(len(leaf.entities), Equals, 1)
			c.Expect(len(leaf.movableEntities), Equals, 1)

			c.Specify("and removed", func() {
				qt.Remove(&MockMobileEntity{})
				c.Expect(len(leaf.entities), Equals, 0)
				c.Expect(len(leaf.movableEntities), Equals, 0)
			})
		})
	})

	c.Specify("quadTree divides when per quad entity limit is reached", func() {
		qt, err := newQuadTree(AABB{
			WorldCoord{-1000, 1000},
			WorldCoord{1000, -1000},
		}, nil, 1)
		c.Assume(err, IsNil)

		// root tree divides
		nwEntity := MockEntity{0, WorldCoord{-1, 1}}
		neEntity := MockEntity{1, WorldCoord{0, 1}}
		qt = qt.Insert(nwEntity)
		qt = qt.Insert(neEntity)

		treeNode, isAQuadTree := qt.(*quadTree)
		c.Assume(isAQuadTree, IsTrue)

		c.Expect(treeNode.quads[QUAD_NW].Contains(nwEntity), IsTrue)
		c.Expect(treeNode.quads[QUAD_NE].Contains(neEntity), IsTrue)

		// Subtree divides
		entity := MockEntity{3, WorldCoord{999, 2}}
		qt = qt.Insert(entity)

		treeNode, isAQuadTree = qt.(*quadTree)
		c.Assume(isAQuadTree, IsTrue)
		c.Expect(treeNode.quads[QUAD_NE].Contains(neEntity), IsTrue)
		c.Expect(treeNode.quads[QUAD_NE].Contains(entity), IsTrue)

		treeNode, isAQuadTree = treeNode.quads[QUAD_NE].(*quadTree)
		c.Assume(isAQuadTree, IsTrue)
		c.Expect(treeNode.quads[QUAD_SW].Contains(neEntity), IsTrue)
		c.Expect(treeNode.quads[QUAD_SE].Contains(entity), IsTrue)

		c.Specify("and removes entities from children", func() {
			treeNode, isAQuadTree := qt.(*quadTree)
			c.Assume(isAQuadTree, IsTrue)

			qt.Remove(nwEntity)
			c.Expect(treeNode.quads[QUAD_NW].Contains(nwEntity), IsFalse)
			c.Expect(qt.Contains(nwEntity), IsFalse)

			qt.Remove(neEntity)
			qt.Remove(entity)

			treeNode, isAQuadTree = qt.(*quadTree)
			c.Assume(isAQuadTree, IsTrue)

			c.Expect(treeNode.quads[QUAD_NE].Contains(neEntity), IsFalse)
			c.Expect(treeNode.quads[QUAD_NE].Contains(entity), IsFalse)

			treeNode, isAQuadTree = treeNode.quads[QUAD_NE].(*quadTree)
			c.Assume(isAQuadTree, IsTrue)

			c.Expect(treeNode.quads[QUAD_SW].Contains(neEntity), IsFalse)
			c.Expect(treeNode.quads[QUAD_SE].Contains(entity), IsFalse)
		})
	})

	c.Specify("stepping forward in time", func() {
		world, err := newQuadTree(AABB{
			WorldCoord{-20, 20},
			WorldCoord{20, -20},
		}, nil, 2)
		c.Assume(err, IsNil)

		world = world.Insert(MockEntity{0, WorldCoord{-10, 10}})
		world = world.Insert(MockEntity{1, WorldCoord{10, 10}})
		world = world.Insert(MockEntity{2, WorldCoord{10, -10}})

		quadTree, isAQuadTree := world.(*quadTree)
		c.Assume(isAQuadTree, IsTrue)

		worldTime := WorldTime(0)
		step := func() {
			worldTime++
			world.AdjustPositions(worldTime)
			world.StepTo(worldTime)
		}

		c.Specify("consume movement requests and apply appropiate path actions", func() {
			entitySpeed := uint(20)
			entity := &MockMobileEntity{2, newMotionInfo(WorldCoord{0, 0}, North, entitySpeed)}

			world = world.Insert(entity)
			c.Assume(quadTree.quads[QUAD_SE].Contains(entity), IsTrue)

			entity.mi.moveRequest = &moveRequest{0, North}

			step()

			// Movement Info state expectations
			c.Expect(entity.mi.moveRequest, IsNil)
			c.Expect(entity.mi.facing, Equals, North)
			c.Expect(len(entity.mi.pathActions), Equals, 1)
			c.Expect(entity.mi.isMoving(), IsTrue)

			// PathAction expectations
			pathAction := entity.mi.pathActions[0]
			c.Expect(pathAction.Orig, Equals, WorldCoord{0, 0})
			c.Expect(pathAction.Dest, Equals, WorldCoord{0, 0}.Neighbor(North))
			c.Expect(pathAction.duration, Equals, int64(entitySpeed))
		})

		c.Specify("consume movement requests and apply appropiate turn facing actions", func() {
			entity := &MockMobileEntity{2, newMotionInfo(WorldCoord{0, 0}, North, 20)}

			world = world.Insert(entity)
			c.Assume(quadTree.quads[QUAD_SE].Contains(entity), IsTrue)

			entity.mi.moveRequest = &moveRequest{0, South}

			step()

			c.Expect(entity.mi.moveRequest, IsNil)
			c.Expect(entity.mi.facing, Equals, South)
			c.Expect(len(entity.mi.pathActions), Equals, 0)

			c.Specify("only after TurnActionDelay ticks have passed", func() {
				facingWillChangeAt := worldTime + TurnActionDelay

				entity.mi.moveRequest = &moveRequest{0, North}

				for step(); worldTime <= facingWillChangeAt; step() {
					c.Expect(entity.mi.moveRequest, Not(IsNil))
					c.Expect(entity.mi.facing, Equals, South)
					c.Expect(len(entity.mi.pathActions), Equals, 0)
				}

				c.Expect(entity.mi.moveRequest, IsNil)
				c.Expect(entity.mi.facing, Equals, North)
				c.Expect(len(entity.mi.pathActions), Equals, 0)
			})
		})

		c.Specify("adjusts entity's position when pathActions have completed", func() {

			path := &PathAction{
				NewTimeSpan(0, 20),
				WorldCoord{0, 0},
				WorldCoord{-1, 0},
			}
			entity := &MockMobileEntity{
				2,
				&motionInfo{
					path.Orig,
					path.Direction(),
					uint(path.duration),
					nil,
					[]*PathAction{path},
					nil,
					nil,
				},
			}

			world = world.Insert(entity)
			c.Assume(quadTree.quads[QUAD_SE].Contains(entity), IsTrue)

			world.AdjustPositions(path.End() - 1)
			c.Expect(len(entity.mi.pathActions), Equals, 1)
			c.Expect(entity.Coord(), Equals, path.Orig)
			c.Expect(quadTree.quads[QUAD_SE].Contains(entity), IsTrue)

			world.AdjustPositions(path.End())
			c.Expect(len(entity.mi.pathActions), Equals, 0)
			c.Expect(entity.Coord(), Equals, path.Dest)

			c.Specify("and moves entity into the proper quad", func() {
				c.Expect(quadTree.quads[QUAD_SW].Contains(entity), IsTrue)
			})
		})

		c.Specify("bound movement inside the world's bounds", func() {
			entity := &MockMobileEntity{2, newMotionInfo(world.AABB().TopL, North, 20)}
			entity.mi.moveRequest = &moveRequest{0, North}

			world = world.Insert(entity)

			step()

			c.Expect(entity.mi.moveRequest, Not(IsNil))
			c.Expect(len(entity.mi.pathActions), Equals, 0)

			// Check that quadLeaf performs the exact same way as quadTree
			worldTime = WorldTime(0)
			world, err = newQuadTree(AABB{
				WorldCoord{-20, 20},
				WorldCoord{20, -20},
			}, nil, 2)
			c.Assume(err, IsNil)

			entity = &MockMobileEntity{0, newMotionInfo(world.AABB().TopL, North, 20)}
			entity.mi.moveRequest = &moveRequest{0, North}

			world = world.Insert(entity)

			_, isQuadLeaf := world.(*quadLeaf)
			c.Assume(isQuadLeaf, IsTrue)

			step()

			c.Expect(entity.mi.moveRequest, Not(IsNil))
			c.Expect(len(entity.mi.pathActions), Equals, 0)
		})
	})
}
