package engine

import (
	"encoding/json"
	"fmt"
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
	"strconv"
)

type spyConn struct {
	packets chan string
}

func (c spyConn) SendJson(msg string, obj interface{}) error {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	c.packets <- msg + ":" + string(jsonBytes)
	return nil
}

func DescribePlayer(c gospec.Context) {
	conn := spyConn{make(chan string)}

	player := &Player{
		Name:     "thundercleese",
		entityId: 0,
		mi:       newMotionInfo(Cell{0, 0}, North, 40),
		conn:     conn,
	}

	player.mux()
	defer player.stopMux()

	c.Specify("motionInfo becomes locked when accessed by the simulation until the worldstate is published", func() {
		_ = player.motionInfo()

		locked := make(chan bool)

		go func() {
			select {
			case player.collectInput <- InputCmd{0, "move", "north"}:
				panic("MotionInfo not locked")
			case <-conn.packets:
				locked <- true
			}
		}()

		player.SendWorldState(newWorldState(Clock(0), AABB{
			Cell{-100, 100},
			Cell{100, -100},
		}).Json())
		c.Expect(<-locked, IsTrue)

		c.Specify("and is unlocked afterwards", func() {
			select {
			case player.collectInput <- InputCmd{0, "move", "north"}:
			default:
				panic("MotionInfo not unlocked")
			}
		})
	})

	c.Specify("a request to move is generated when the user inputs a move cmd", func() {
		player.SubmitInput("move=0", "north")

		moveRequest := player.motionInfo().moveRequest

		c.Expect(moveRequest, Not(IsNil))
	})

	c.Specify("a requst to move is canceled by a moveCancel cmd", func() {
		player.SubmitInput("move=0", "north")
		player.SubmitInput("moveCancel=0", "north")

		c.Expect(player.motionInfo().moveRequest, IsNil)
	})

	c.Specify("a moveCancel cmd is dropped if it doesn't cancel the current move request", func() {
		player.SubmitInput("move=0", "north")
		player.SubmitInput("moveCancel=0", "south")

		c.Expect(player.motionInfo().moveRequest, Not(IsNil))
	})

	c.Specify("generates json compatitable state object", func() {
		jsonBytes, err := json.Marshal(player.Json())
		c.Expect(err, IsNil)
		c.Expect(string(jsonBytes), Equals, `{"id":0,"name":"thundercleese","facing":"north","pathActions":null,"cell":{"x":0,"y":0}}`)

		player.mi.pathActions = append(player.mi.pathActions, &PathAction{
			NewTimeSpan(0, 10),
			Cell{0, 0},
			Cell{0, 1},
		})

		jsonBytes, err = json.Marshal(player.Json())
		c.Expect(err, IsNil)
		c.Expect(string(jsonBytes), Equals, `{"id":0,"name":"thundercleese","facing":"north","pathActions":[{"start":0,"end":10,"orig":{"x":0,"y":0},"dest":{"x":0,"y":1}}],"cell":{"x":0,"y":0}}`)
	})
}

func DescribeViewPortCulling(c gospec.Context) {
	c.Specify("will cull the world state into the viewable area of the player", func() {
		toBeCulled := []EntityJson{
			MockEntity{cell: Cell{-27, 27}}.Json(),
			MockEntity{cell: Cell{27, 27}}.Json(),
			MockEntity{cell: Cell{27, -27}}.Json(),
			MockEntity{cell: Cell{-27, -27}}.Json(),
		}

		wontBeCulled := []EntityJson{
			MockEntity{cell: Cell{-26, 26}}.Json(),
			MockEntity{cell: Cell{26, 26}}.Json(),
			MockEntity{cell: Cell{26, -26}}.Json(),
			MockEntity{cell: Cell{-26, -26}}.Json(),
		}

		player := &Player{mi: &motionInfo{cell: Cell{0, 0}}}
		state := WorldStateJson{}
		state.Entities = append(state.Entities, toBeCulled...)
		state.Entities = append(state.Entities, wontBeCulled...)

		state = player.CullStateToView(state)

		c.Expect(state.Entities, ContainsAll, wontBeCulled)
		c.Expect(state.Entities, Not(ContainsAll), toBeCulled)
	})
}

func DescribePlayerCollisions(c gospec.Context) {
	c.Specify("when a location is contested", func() {
		contested := Cell{0, 0}

		c.Specify("by 2 players", func() {
			playerA := &Player{mi: newMotionInfo(contested.Neighbor(South), North, 20)}
			startA := WorldTime(0)
			pathA := &PathAction{
				NewTimeSpan(startA, startA+WorldTime(playerA.mi.speed)),
				playerA.mi.cell,
				contested,
			}
			playerA.mi.Apply(pathA)
			c.Assume(playerA.mi.isMoving(), IsTrue)

			specs := []struct {
				Direction
				description string
			}{{
				South,
				"Head to Head",
			}, {
				East,
				"From the side [East]",
			}, {
				West,
				"From the side [West]",
			}}

			c.Specify("at the same time", func() {
				startB := startA

				for _, spec := range specs {
					c.Specify(spec.description, func() {
						c.Specify("if the speeds are different the faster player wins", func() {
							playerB := &Player{mi: newMotionInfo(contested.Neighbor(spec.Direction.Reverse()), spec.Direction, 21)}
							pathB := &PathAction{
								NewTimeSpan(startB, startB+WorldTime(playerB.mi.speed)),
								playerB.mi.cell,
								contested,
							}
							playerB.mi.Apply(pathB)

							c.Assume(playerB.mi.isMoving(), IsTrue)
							// Assume PlayerA is faster
							c.Assume(playerA.mi.speed, Satisfies, playerA.mi.speed < playerB.mi.speed)

							c.Specify("player A wins", func() {
								entityCollision{startA, playerA, playerB}.collide()
								c.Expect(playerA.mi.isMoving(), IsTrue)
								c.Expect(playerB.mi.isMoving(), IsFalse)
							})

							c.Specify("player A wins", func() {
								entityCollision{startA, playerB, playerA}.collide()
								c.Expect(playerA.mi.isMoving(), IsTrue)
								c.Expect(playerB.mi.isMoving(), IsFalse)
							})
						})

						c.Specify("if the speeds are the same", func() {
							playerB := &Player{mi: newMotionInfo(contested.Neighbor(spec.Direction.Reverse()),
								spec.Direction,
								playerA.mi.speed)}

							pathB := &PathAction{
								NewTimeSpan(startB, startB+WorldTime(playerB.mi.speed)),
								playerB.mi.cell,
								contested,
							}
							playerB.mi.Apply(pathB)

							c.Assume(playerB.mi.isMoving(), IsTrue)
							c.Assume(playerA.mi.speed, Equals, playerB.mi.speed)

							c.Specify("player A wins", func() {
								entityCollision{startA, playerB, playerA}.collide()

								c.Expect(playerA.mi.isMoving(), IsTrue)
								c.Expect(playerB.mi.isMoving(), IsFalse)
							})

							c.Specify("player B wins", func() {
								entityCollision{startA, playerA, playerB}.collide()

								c.Expect(playerA.mi.isMoving(), IsFalse)
								c.Expect(playerB.mi.isMoving(), IsTrue)
							})
						})
					})
				}

			})

			c.Specify("the player that was already moving wins", func() {
				for _, spec := range specs {
					c.Specify(spec.description, func() {
						playerB := &Player{mi: newMotionInfo(contested.Neighbor(spec.Direction.Reverse()), spec.Direction, 20)}
						startB := startA + 1
						pathB := &PathAction{
							NewTimeSpan(startB, startB+WorldTime(playerB.mi.speed)),
							playerB.mi.cell,
							contested,
						}
						playerB.mi.Apply(pathB)

						c.Assume(playerB.mi.isMoving(), IsTrue)

						c.Specify("player A wins", func() {
							entityCollision{startB, playerA, playerB}.collide()

							c.Expect(playerA.mi.isMoving(), IsTrue)
							c.Expect(playerB.mi.isMoving(), IsFalse)
						})

						c.Specify("player A wins", func() {
							entityCollision{startB, playerB, playerA}.collide()

							c.Expect(playerA.mi.isMoving(), IsTrue)
							c.Expect(playerB.mi.isMoving(), IsFalse)
						})
					})
				}
			})
		})
	})

	c.Specify("when a location is occupied", func() {
		playerNotMoving := &Player{mi: newMotionInfo(Cell{0, 0}, North, 20)}
		// TODO this reads badly
		c.Assume(playerNotMoving.mi.isMoving(), IsFalse)

		directions := [...]Direction{North, South, East, West}

		for _, direction := range directions {
			c.Specify(fmt.Sprintf("and a player is attempting to move there from the %v", direction), func() {
				player := &Player{mi: newMotionInfo(playerNotMoving.Cell().Neighbor(direction), direction.Reverse(), 20)}
				path := &PathAction{
					NewTimeSpan(0, 20),
					player.Cell(),
					playerNotMoving.Cell(),
				}
				player.mi.Apply(path)

				c.Assume(player.mi.isMoving(), IsTrue)

				c.Specify("they are unable to [AB]", func() {
					entityCollision{0, player, playerNotMoving}.collide()
					c.Expect(player.mi.isMoving(), IsFalse)
				})

				c.Specify("they are unable to [BA]", func() {
					entityCollision{0, playerNotMoving, player}.collide()
					c.Expect(player.mi.isMoving(), IsFalse)
				})
			})
		}

		directionPairs := [...]struct {
			A, B Direction
		}{{
			North, South,
		}, {
			North, East,
		}, {
			North, West,
		}, {
			South, East,
		}, {
			South, West,
		}, {
			East, West,
		}}

		for _, direction := range directionPairs {
			c.Specify(fmt.Sprintf("and 2 players are attempting to move there from the %v and the %v", direction.A, direction.B), func() {
				players := [...]*Player{
					{mi: newMotionInfo(playerNotMoving.Cell().Neighbor(direction.A), direction.A.Reverse(), 20)},
					{mi: newMotionInfo(playerNotMoving.Cell().Neighbor(direction.B), direction.B.Reverse(), 20)},
				}

				for _, player := range players {
					player.mi.Apply(&PathAction{
						NewTimeSpan(0, 0+WorldTime(player.mi.speed)),
						player.Cell(),
						playerNotMoving.Cell(),
					})
					c.Assume(player.mi.isMoving(), IsTrue)
				}

				collisions := [...]entityCollision{
					entityCollision{0, players[0], players[1]},
					entityCollision{0, players[0], playerNotMoving},
					entityCollision{0, players[1], playerNotMoving},
				}

				// TODO figure out an algorithm to check all orderings
				// 123
				// 132
				// 213
				// 231
				// 312
				// 321
				for _, c := range collisions {
					c.collide()
				}

				for _, player := range players {
					c.Expect(player.mi.isMoving(), IsFalse)
				}
			})
		}
	})

	c.Specify("when players are attempting to swap positions", func() {
		a, b := Cell{0, 0}, Cell{1, 0}
		playerA := &Player{mi: newMotionInfo(a, a.DirectionTo(b), 20)}
		playerB := &Player{mi: newMotionInfo(b, b.DirectionTo(a), 20)}

		playerA.mi.Apply(&PathAction{
			NewTimeSpan(0, 20),
			a, b,
		})

		playerB.mi.Apply(&PathAction{
			NewTimeSpan(0, 20),
			b, a,
		})

		c.Specify("AB", func() {
			entityCollision{0, playerA, playerB}.collide()
			c.Expect(playerA.mi.isMoving(), IsFalse)
			c.Expect(playerB.mi.isMoving(), IsFalse)
		})

		c.Specify("BA", func() {
			entityCollision{0, playerB, playerA}.collide()
			c.Expect(playerA.mi.isMoving(), IsFalse)
			c.Expect(playerB.mi.isMoving(), IsFalse)
		})
	})

	c.Specify("when a player is following into another players position", func() {
		c.Specify("from directly behind", func() {
			m, n, o := Cell{-1, 0}, Cell{0, 0}, Cell{1, 0}
			playerA := &Player{mi: newMotionInfo(m, m.DirectionTo(n), 15)}
			playerB := &Player{mi: newMotionInfo(n, n.DirectionTo(o), 20)}

			playerA.mi.Apply(&PathAction{NewTimeSpan(20, 35), m, n})
			playerB.mi.Apply(&PathAction{NewTimeSpan(20, 40), n, o})

			c.Specify("player A doesn't move", func() {
				entityCollision{20, playerA, playerB}.collide()
				c.Expect(playerA.mi.isMoving(), IsFalse)
				c.Expect(playerB.mi.isMoving(), IsTrue)
			})

			c.Specify("player A doesn't move", func() {
				entityCollision{20, playerB, playerA}.collide()
				c.Expect(playerA.mi.isMoving(), IsFalse)
				c.Expect(playerB.mi.isMoving(), IsTrue)
			})
		})

		c.Specify("from the side", func() {
			c.Specify("the player is unable to follow until the collision ends before or when their path has ended", func() {
				m, n, o := Cell{1, 0}, Cell{0, 0}, Cell{0, 1}
				playerA := &Player{mi: newMotionInfo(o, o.DirectionTo(n), 19)}
				playerB := &Player{mi: newMotionInfo(n, n.DirectionTo(m), 20)}

				time := WorldTime(10)

				pathA := &PathAction{NewTimeSpan(time, time+WorldTime(playerA.mi.speed)), o, n}
				pathB := &PathAction{NewTimeSpan(time, time+WorldTime(playerB.mi.speed)), n, m}
				c.Assume(pathCollision(*pathA, *pathB).Type(), Equals, CT_A_INTO_B_FROM_SIDE)

				playerA.mi.Apply(pathA)
				playerB.mi.Apply(pathB)

				c.Specify("player A can't move yet", func() {
					entityCollision{time, playerA, playerB}.collide()
					c.Expect(playerA.mi.isMoving(), IsFalse)
					c.Expect(playerB.mi.isMoving(), IsTrue)
				})

				c.Specify("player A can't move yet", func() {
					entityCollision{time, playerB, playerA}.collide()
					c.Expect(playerA.mi.isMoving(), IsFalse)
					c.Expect(playerB.mi.isMoving(), IsTrue)
				})

				c.Specify("player A can move now", func() {
					entityCollision{time, playerA, playerB}.collide()

					time++
					playerA.mi.Apply(&PathAction{NewTimeSpan(time, time+WorldTime(playerA.mi.speed)), o, n})

					c.Specify("and is moving", func() {
						entityCollision{time, playerA, playerB}.collide()
						c.Expect(playerA.mi.isMoving(), IsTrue)
						c.Expect(playerB.mi.isMoving(), IsTrue)
					})

					c.Specify("and is moving", func() {
						entityCollision{time, playerB, playerA}.collide()
						c.Expect(playerA.mi.isMoving(), IsTrue)
						c.Expect(playerB.mi.isMoving(), IsTrue)
					})
				})
			})
		})
	})
}

func DescribePlayerJson(c gospec.Context) {
	player := PlayerJson{
		Facing: North.String(),
		Cell:   Cell{0, 0},
	}

	c.Specify("can identify if nothing has changed", func() {
		c.Specify("when the player is standing still", func() {
			c.Expect(player.IsDifferentFrom(player), IsFalse)
		})

		c.Specify("when the player is moving", func() {
			player.PathActions = []PathActionJson{
				PathAction{
					NewTimeSpan(10, 30),
					player.Cell,
					player.Cell.Neighbor(North),
				}.Json(),
			}
			c.Expect(player.IsDifferentFrom(player), IsFalse)
		})
	})

	c.Specify("can identify when something is different", func() {
		playerChanged := player
		c.Specify("when the player's facing has changed", func() {
			playerChanged.Facing = South.String()
			c.Expect(player.IsDifferentFrom(playerChanged), IsTrue)
		})

		c.Specify("when the player has started moving", func() {
			playerChanged.PathActions = []PathActionJson{
				PathAction{
					NewTimeSpan(10, 30),
					player.Cell,
					player.Cell.Neighbor(North),
				}.Json(),
			}

			c.Expect(player.IsDifferentFrom(playerChanged), IsTrue)
		})

		c.Specify("when the player has stopped moving", func() {
			playerChanged.Cell = player.Cell.Neighbor(North)
			player.PathActions = []PathActionJson{
				PathAction{
					NewTimeSpan(10, 30),
					player.Cell,
					player.Cell.Neighbor(North),
				}.Json(),
			}

			c.Expect(player.IsDifferentFrom(playerChanged), IsTrue)
		})

		c.Specify("when the player has finished a movement and continued moving", func() {
			player.PathActions = []PathActionJson{
				PathAction{
					NewTimeSpan(10, 30),
					player.Cell,
					player.Cell.Neighbor(North),
				}.Json(),
			}

			playerChanged.Cell = player.Cell.Neighbor(North)
			playerChanged.PathActions = []PathActionJson{
				PathAction{
					NewTimeSpan(30, 50),
					playerChanged.Cell,
					playerChanged.Cell.Neighbor(West),
				}.Json(),
			}

			c.Expect(player.IsDifferentFrom(playerChanged), IsTrue)
		})
	})
}

func DescribeInputCommands(c gospec.Context) {
	c.Specify("creating movement requests from InputCmds", func() {
		c.Specify("north", func() {
			moveRequest := newMoveRequest(InputCmd{
				0,
				"move",
				"north",
			})

			c.Expect(moveRequest.t, Equals, WorldTime(0))
			c.Expect(moveRequest.Direction, Equals, North)
		})

		c.Specify("east", func() {
			moveRequest := newMoveRequest(InputCmd{
				0,
				"move",
				"east",
			})

			c.Expect(moveRequest.t, Equals, WorldTime(0))
			c.Expect(moveRequest.Direction, Equals, East)
		})

		c.Specify("south", func() {
			moveRequest := newMoveRequest(InputCmd{
				0,
				"move",
				"south",
			})

			c.Expect(moveRequest.t, Equals, WorldTime(0))
			c.Expect(moveRequest.Direction, Equals, South)
		})

		c.Specify("west", func() {
			moveRequest := newMoveRequest(InputCmd{
				0,
				"move",
				"west",
			})

			c.Expect(moveRequest.t, Equals, WorldTime(0))
			c.Expect(moveRequest.Direction, Equals, West)
		})
	})

	c.Specify("embedding worldtime in the cmd msg", func() {
		player := &Player{
			Name:         "thundercleese",
			entityId:     0,
			mi:           newMotionInfo(Cell{0, 0}, North, 40),
			conn:         noopConn(0),
			collectInput: make(chan InputCmd, 1),
		}

		c.Specify("string splits on = and parses 64bit int", func() {
			player.SubmitInput("move=0", "north")
			input := <-player.collectInput

			c.Expect(input.timeIssued, Equals, WorldTime(0))

			player.SubmitInput("move=1824081", "north")
			input = <-player.collectInput

			c.Expect(input.timeIssued, Equals, WorldTime(1824081))

			player.SubmitInput("move=99", "north")
			input = <-player.collectInput

			c.Expect(input.timeIssued, Equals, WorldTime(99))
		})

		c.Specify("errors with invalid input and doesn't publish the command", func() {
			err := player.SubmitInput("move=a", "north")
			e := err.(*strconv.NumError)

			c.Expect(e, Not(IsNil))
			c.Expect(e.Err, Equals, strconv.ErrSyntax)

			err = player.SubmitInput("move=", "north")
			e = err.(*strconv.NumError)

			c.Expect(e, Not(IsNil))
			c.Expect(e.Err, Equals, strconv.ErrSyntax)
		})
	})
}
