package engine

import (
	"fmt"

	. "github.com/ghthor/engine/coord"
	. "github.com/ghthor/engine/time"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeQuad(c gospec.Context) {
	c.Specify("quad can be queried with an AABB to return all entities that are inside", func() {
		qt, err := newQuadTree(AABB{
			Cell{-1000, 1000},
			Cell{1000, -1000},
		}, nil, 1)
		c.Assume(err, IsNil)

		qt = qt.Insert(MockEntity{0, Cell{0, 0}})
		qt = qt.Insert(MockEntity{1, Cell{10, 10}})
		qt = qt.Insert(MockEntity{2, Cell{-10, -10}})
		qt = qt.Insert(MockEntity{3, Cell{999, -1000}})

		entities := qt.QueryAll(AABB{
			Cell{0, 0},
			Cell{10, -10},
		})
		c.Expect(len(entities), Equals, 1)

		entities = qt.QueryAll(AABB{
			Cell{-90, 90},
			Cell{-1, -9},
		})
		c.Expect(len(entities), Equals, 0)

		entities = qt.QueryAll(qt.AABB())
		c.Expect(len(entities), Equals, 4)
	})

	c.Specify("quad can be queried with a cell for any collidable entities", func() {
		world, err := newQuadTree(AABB{
			Cell{-20, 20},
			Cell{20, -20},
		}, nil, 20)
		c.Assume(err, IsNil)

		entity := &MockAliveEntity{mi: newMotionInfo(Cell{0, 0}, North, 20)}
		path := &PathAction{
			NewSpan(0, 20),
			Cell{0, 0},
			Cell{0, 1},
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
			Cell{-1000, 1000},
			Cell{1000, -1000},
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
			Cell{-1000, 1000},
			Cell{1000, -1000},
		}, nil, 1)
		c.Assume(err, IsNil)

		// root tree divides
		nwEntity := MockEntity{0, Cell{-1, 1}}
		neEntity := MockEntity{1, Cell{0, 1}}
		qt = qt.Insert(nwEntity)
		qt = qt.Insert(neEntity)

		treeNode, isAQuadTree := qt.(*quadTree)
		c.Assume(isAQuadTree, IsTrue)

		c.Expect(treeNode.quads[QUAD_NW].Contains(nwEntity), IsTrue)
		c.Expect(treeNode.quads[QUAD_NE].Contains(neEntity), IsTrue)

		// Subtree divides
		entity := MockEntity{3, Cell{999, 2}}
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
			Cell{-20, 20},
			Cell{20, -20},
		}, nil, 20)
		c.Assume(err, IsNil)

		worldTime := Time(0)
		step := func() {
			worldTime++
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
			entity := &MockMobileEntity{nextId(), newMotionInfo(Cell{0, 0}, North, entitySpeed)}

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
			c.Expect(pathAction.Orig, Equals, Cell{0, 0})
			c.Expect(pathAction.Dest, Equals, Cell{0, 0}.Neighbor(North))
			c.Expect(pathAction.Duration, Equals, int64(entitySpeed))
		})

		c.Specify("consume movement requests and apply appropiate turn facing actions", func() {
			entity := &MockMobileEntity{nextId(), newMotionInfo(Cell{0, 0}, North, 20)}

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
				NewSpan(0, 20),
				Cell{0, 0},
				Cell{-1, 0},
			}
			entity := &MockMobileEntity{
				nextId(),
				&motionInfo{
					path.Orig,
					path.Direction(),
					uint(path.Duration),
					nil,
					[]*PathAction{path},
					nil,
					nil,
				},
			}

			c.Specify("within the same leaf", func() {
				world = world.Insert(entity)

				world.updatePositions(path.End() - 1)
				c.Expect(len(entity.mi.pathActions), Equals, 1)
				c.Expect(entity.Cell(), Equals, path.Orig)

				world.updatePositions(path.End())
				c.Expect(len(entity.mi.pathActions), Equals, 0)
				c.Expect(entity.Cell(), Equals, path.Dest)
			})

			c.Specify("and entity is transfered to another leaf", func() {
				tree := divideLeafIntoTree()

				world = world.Insert(entity)
				c.Assume(tree.quads[QUAD_SE].Contains(entity), IsTrue)

				world.updatePositions(path.End() - 1)
				c.Expect(len(entity.mi.pathActions), Equals, 1)
				c.Expect(entity.Cell(), Equals, path.Orig)
				c.Expect(tree.quads[QUAD_SE].Contains(entity), IsTrue)

				world.updatePositions(path.End())
				c.Expect(len(entity.mi.pathActions), Equals, 0)
				c.Expect(entity.Cell(), Equals, path.Dest)
				c.Expect(tree.quads[QUAD_SW].Contains(entity), IsTrue)
			})
		})

		c.Specify("bound movement inside the world's bounds", func() {
			cases := []struct {
				position    Cell
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
			c.Specify("when 2 alive entities contest a cell", func() {
				contestedPoint := Cell{0, 0}

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

			c.Specify("when 3 alive entities contest a cell", func() {
				contestedPoint := Cell{0, 0}

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

			c.Specify("when 4 alive entities contest a cell", func() {
				contestedPoint := Cell{0, 0}

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
