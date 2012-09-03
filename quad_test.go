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

	c.Specify("expand AABB by a magnitude", func() {
		c.Expect(aabb.Expand(1), Equals, AABB{
			WorldCoord{-1, 1},
			WorldCoord{1, -1},
		})

		aabb = AABB{
			WorldCoord{5, 6},
			WorldCoord{5, -6},
		}

		c.Expect(aabb.Expand(2), Equals, AABB{
			WorldCoord{3, 8},
			WorldCoord{7, -8},
		})
	})
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

	c.Specify("quad can be queried with a coord for any collidable entities", func() {
		world, err := newQuadTree(AABB{
			WorldCoord{-20, 20},
			WorldCoord{20, -20},
		}, nil, 20)
		c.Assume(err, IsNil)

		entity := &MockAliveEntity{mi: newMotionInfo(WorldCoord{0, 0}, North, 20)}
		path := &PathAction{
			NewTimeSpan(0, 20),
			WorldCoord{0, 0},
			WorldCoord{0, 1},
		}
		entity.mi.pathActions = append(entity.mi.pathActions, path)

		world = world.Insert(entity)

		divideLeafIntoTree := func() *quadTree {
			world = world.(*quadLeaf).divide()
			quadTree, isAQuadTree := world.(*quadTree)
			c.Assume(isAQuadTree, IsTrue)
			return quadTree
		}

		c.Specify("if the quad is a leaf", func() {
			matches := world.QueryCollidables(path.Orig)
			c.Expect(len(matches), Equals, 1)

			matches = world.QueryCollidables(path.Dest)
			c.Expect(len(matches), Equals, 1)
		})

		c.Specify("if the quad is a tree", func() {
			tree := divideLeafIntoTree()
			c.Assume(tree.quads[QUAD_SE].AABB().Contains(path.Orig), IsTrue)
			c.Assume(tree.quads[QUAD_SE].AABB().Contains(path.Dest), IsFalse)

			matches := world.QueryCollidables(path.Orig)
			c.Expect(len(matches), Equals, 1)

			matches = world.QueryCollidables(path.Dest)
			c.Expect(len(matches), Equals, 1)
		})
	})

	c.Specify("entity is inserted into quadtree", func() {
		qt, err := newQuadTree(AABB{
			WorldCoord{-1000, 1000},
			WorldCoord{1000, -1000},
		}, nil, 1)
		c.Assume(err, IsNil)

		c.Specify("as an entity", func() {
			entity := MockEntity{}
			qt = qt.Insert(entity)

			leaf, isAQuadLeaf := qt.(*quadLeaf)
			c.Assume(isAQuadLeaf, IsTrue)

			c.Expect(len(leaf.entities), Equals, 1)
			c.Expect(len(leaf.movable), Equals, 0)
			c.Expect(len(leaf.collidable), Equals, 0)

			c.Specify("and removed", func() {
				qt.Remove(entity)
				c.Expect(len(leaf.entities), Equals, 0)
			})
		})

		c.Specify("as an movableEntity", func() {
			entity := &MockMobileEntity{}
			qt = qt.Insert(entity)

			leaf, isAQuadLeaf := qt.(*quadLeaf)
			c.Assume(isAQuadLeaf, IsTrue)

			c.Expect(len(leaf.entities), Equals, 1)
			c.Expect(len(leaf.movable), Equals, 1)
			c.Expect(len(leaf.collidable), Equals, 0)

			c.Specify("and removed", func() {
				qt.Remove(entity)
				c.Expect(len(leaf.entities), Equals, 0)
				c.Expect(len(leaf.movable), Equals, 0)
			})
		})

		c.Specify("as an collidable entity", func() {
			entity := &MockCollidableEntity{}
			qt = qt.Insert(entity)

			leaf, isAQuadLeaf := qt.(*quadLeaf)
			c.Assume(isAQuadLeaf, IsTrue)

			c.Expect(len(leaf.entities), Equals, 1)
			c.Expect(len(leaf.movable), Equals, 0)
			c.Expect(len(leaf.collidable), Equals, 1)

			c.Specify("and removed", func() {
				qt.Remove(entity)
				c.Expect(len(leaf.entities), Equals, 0)
				c.Expect(len(leaf.collidable), Equals, 0)
			})
		})

		c.Specify("as an alive entity", func() {
			entity := &MockAliveEntity{}
			qt = qt.Insert(entity)

			leaf, isAQuadLeaf := qt.(*quadLeaf)
			c.Assume(isAQuadLeaf, IsTrue)

			c.Expect(len(leaf.entities), Equals, 1)
			c.Expect(len(leaf.movable), Equals, 1)
			c.Expect(len(leaf.collidable), Equals, 1)

			c.Specify("and removed", func() {
				qt.Remove(entity)
				c.Expect(len(leaf.entities), Equals, 0)
				c.Expect(len(leaf.movable), Equals, 0)
				c.Expect(len(leaf.collidable), Equals, 0)
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
		}, nil, 20)
		c.Assume(err, IsNil)

		worldTime := WorldTime(0)
		step := func() {
			worldTime++
			world.AdjustPositions(worldTime)
			world.StepTo(worldTime)
		}

		entityCounter := EntityId(0)
		nextId := func() EntityId {
			defer func() { entityCounter += 1 }()
			return entityCounter
		}

		divideLeafIntoTree := func() *quadTree {
			world = world.(*quadLeaf).divide()
			quadTree, isAQuadTree := world.(*quadTree)
			c.Assume(isAQuadTree, IsTrue)
			return quadTree
		}

		c.Specify("consume movement requests and apply appropiate path actions", func() {
			entitySpeed := uint(20)
			entity := &MockMobileEntity{nextId(), newMotionInfo(WorldCoord{0, 0}, North, entitySpeed)}

			world = world.Insert(entity)

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
			entity := &MockMobileEntity{nextId(), newMotionInfo(WorldCoord{0, 0}, North, 20)}

			world = world.Insert(entity)

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
				nextId(),
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

			c.Specify("within the same leaf", func() {
				world = world.Insert(entity)

				world.AdjustPositions(path.End() - 1)
				c.Expect(len(entity.mi.pathActions), Equals, 1)
				c.Expect(entity.Coord(), Equals, path.Orig)

				world.AdjustPositions(path.End())
				c.Expect(len(entity.mi.pathActions), Equals, 0)
				c.Expect(entity.Coord(), Equals, path.Dest)
			})

			c.Specify("and entity is transfered to another leaf", func() {
				tree := divideLeafIntoTree()

				world = world.Insert(entity)
				c.Assume(tree.quads[QUAD_SE].Contains(entity), IsTrue)

				world.AdjustPositions(path.End() - 1)
				c.Expect(len(entity.mi.pathActions), Equals, 1)
				c.Expect(entity.Coord(), Equals, path.Orig)
				c.Expect(tree.quads[QUAD_SE].Contains(entity), IsTrue)

				world.AdjustPositions(path.End())
				c.Expect(len(entity.mi.pathActions), Equals, 0)
				c.Expect(entity.Coord(), Equals, path.Dest)
				c.Expect(tree.quads[QUAD_SW].Contains(entity), IsTrue)
			})
		})

		c.Specify("bound movement inside the world's bounds", func() {
			cases := []struct {
				position    WorldCoord
				directions  []Direction
				description string
			}{{
				world.AABB().TopL,
				[]Direction{West, North},
				"top left",
			}, {
				world.AABB().TopR(),
				[]Direction{North, East},
				"top right",
			}, {
				world.AABB().BotR,
				[]Direction{East, South},
				"bot right",
			}, {
				world.AABB().BotL(),
				[]Direction{South, West},
				"bot left",
			}}

			for _, specCase := range cases {
				for _, direction := range specCase.directions {
					c.Specify(fmt.Sprintf("from %v moving %v", specCase.description, direction), func() {
						entity := &MockMobileEntity{nextId(), newMotionInfo(specCase.position, direction, 20)}
						entity.mi.moveRequest = &moveRequest{0, direction}

						world = world.Insert(entity)
						c.Specify("as a quadLeaf", func() {
							step()
							c.Expect(entity.mi.moveRequest, Not(IsNil))
							c.Expect(len(entity.mi.pathActions), Equals, 0)
						})

						c.Specify("as a quadTree", func() {
							divideLeafIntoTree()
							step()
							c.Expect(entity.mi.moveRequest, Not(IsNil))
							c.Expect(len(entity.mi.pathActions), Equals, 0)
						})
					})
				}
			}
		})

		c.Specify("identifies a collision", func() {
			c.Specify("when 2 alive entities contest a coordinate", func() {
				contestedPoint := WorldCoord{0, 0}

				entityA := &MockAliveEntity{
					nextId(),
					newMotionInfo(contestedPoint.Neighbor(South), North, 20),
					make([]MockCollision, 0, 2),
				}
				entityA.mi.moveRequest = &moveRequest{worldTime, North}

				entityB := &MockAliveEntity{
					nextId(),
					newMotionInfo(contestedPoint.Neighbor(North), South, 20),
					make([]MockCollision, 0, 2),
				}
				entityB.mi.moveRequest = &moveRequest{worldTime, South}

				world = world.Insert(entityA)
				world = world.Insert(entityB)

				c.Specify("within the same leaf", func() {
					_, isALeaf := world.(*quadLeaf)
					c.Assume(isALeaf, IsTrue)

					step()

					c.Expect(entityA, CollidedWith, entityB)
					c.Expect(len(entityA.collisions), Equals, 1)
					c.Expect(len(entityB.collisions), Equals, 1)
				})

				c.Specify("across seperate leaves", func() {
					divideLeafIntoTree()

					step()

					c.Expect(entityA, CollidedWith, entityB)
					c.Expect(len(entityA.collisions), Equals, 1)
					c.Expect(len(entityB.collisions), Equals, 1)
				})
			})

			c.Specify("when 3 alive entities contest a coordinate", func() {
				contestedPoint := WorldCoord{0, 0}

				entityA := &MockAliveEntity{
					nextId(),
					newMotionInfo(contestedPoint.Neighbor(South), North, 20),
					make([]MockCollision, 0, 2),
				}
				entityA.mi.moveRequest = &moveRequest{worldTime, North}

				entityB := &MockAliveEntity{
					nextId(),
					newMotionInfo(contestedPoint.Neighbor(North), South, 20),
					make([]MockCollision, 0, 2),
				}
				entityB.mi.moveRequest = &moveRequest{worldTime, South}

				entityC := &MockAliveEntity{
					nextId(),
					newMotionInfo(contestedPoint.Neighbor(West), East, 20),
					make([]MockCollision, 0, 2),
				}
				entityC.mi.moveRequest = &moveRequest{worldTime, East}

				world = world.Insert(entityA)
				world = world.Insert(entityB)
				world = world.Insert(entityC)

				c.Specify("within the same leaf", func() {
					_, isALeaf := world.(*quadLeaf)
					c.Assume(isALeaf, IsTrue)

					step()

					c.Expect(entityA, CollidedWith, entityB)
					c.Expect(entityA, CollidedWith, entityC)
					c.Expect(entityB, CollidedWith, entityC)

					c.Expect(len(entityA.collisions), Equals, 2)
					c.Expect(len(entityB.collisions), Equals, 2)
					c.Expect(len(entityC.collisions), Equals, 2)
				})

				c.Specify("across seperate leaves", func() {
					divideLeafIntoTree()

					step()

					c.Expect(entityA, CollidedWith, entityB)
					c.Expect(entityA, CollidedWith, entityC)
					c.Expect(entityB, CollidedWith, entityC)

					c.Expect(len(entityA.collisions), Equals, 2)
					c.Expect(len(entityB.collisions), Equals, 2)
					c.Expect(len(entityC.collisions), Equals, 2)
				})
			})

			c.Specify("when 4 alive entities contest a coordinate", func() {
				contestedPoint := WorldCoord{0, 0}

				entityA := &MockAliveEntity{
					nextId(),
					newMotionInfo(contestedPoint.Neighbor(South), North, 20),
					make([]MockCollision, 0, 2),
				}
				entityA.mi.moveRequest = &moveRequest{worldTime, North}

				entityB := &MockAliveEntity{
					nextId(),
					newMotionInfo(contestedPoint.Neighbor(North), South, 20),
					make([]MockCollision, 0, 2),
				}
				entityB.mi.moveRequest = &moveRequest{worldTime, South}

				entityC := &MockAliveEntity{
					nextId(),
					newMotionInfo(contestedPoint.Neighbor(West), East, 20),
					make([]MockCollision, 0, 2),
				}
				entityC.mi.moveRequest = &moveRequest{worldTime, East}

				entityD := &MockAliveEntity{
					nextId(),
					newMotionInfo(contestedPoint.Neighbor(East), West, 20),
					make([]MockCollision, 0, 2),
				}
				entityD.mi.moveRequest = &moveRequest{worldTime, West}

				world = world.Insert(entityA)
				world = world.Insert(entityB)
				world = world.Insert(entityC)
				world = world.Insert(entityD)

				c.Specify("within the same leaf", func() {
					_, isALeaf := world.(*quadLeaf)
					c.Assume(isALeaf, IsTrue)

					step()

					c.Expect(entityA, CollidedWith, entityB)
					c.Expect(entityA, CollidedWith, entityC)
					c.Expect(entityA, CollidedWith, entityD)
					c.Expect(entityB, CollidedWith, entityC)
					c.Expect(entityB, CollidedWith, entityD)
					c.Expect(entityC, CollidedWith, entityD)

					c.Expect(len(entityA.collisions), Equals, 3)
					c.Expect(len(entityB.collisions), Equals, 3)
					c.Expect(len(entityC.collisions), Equals, 3)
					c.Expect(len(entityD.collisions), Equals, 3)
				})

				c.Specify("across seperate leaves", func() {
					divideLeafIntoTree()

					step()

					c.Expect(entityA, CollidedWith, entityB)
					c.Expect(entityA, CollidedWith, entityC)
					c.Expect(entityA, CollidedWith, entityD)
					c.Expect(entityB, CollidedWith, entityC)
					c.Expect(entityB, CollidedWith, entityD)
					c.Expect(entityC, CollidedWith, entityD)

					c.Expect(len(entityA.collisions), Equals, 3)
					c.Expect(len(entityB.collisions), Equals, 3)
					c.Expect(len(entityC.collisions), Equals, 3)
					c.Expect(len(entityD.collisions), Equals, 3)
				})
			})
		})
	})
}
