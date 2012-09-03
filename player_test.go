package engine

import (
	"encoding/json"
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
		mi:       newMotionInfo(WorldCoord{0, 0}, North, 40),
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

		player.SendWorldState(newWorldState(Clock(0)).Json())
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
		c.Expect(string(jsonBytes), Equals, `{"id":0,"name":"thundercleese","facing":"north","pathActions":null,"coord":{"x":0,"y":0}}`)

		player.mi.pathActions = append(player.mi.pathActions, &PathAction{
			NewTimeSpan(0, 10),
			WorldCoord{0, 0},
			WorldCoord{0, 1},
		})

		jsonBytes, err = json.Marshal(player.Json())
		c.Expect(err, IsNil)
		c.Expect(string(jsonBytes), Equals, `{"id":0,"name":"thundercleese","facing":"north","pathActions":[{"start":0,"end":10,"orig":{"x":0,"y":0},"dest":{"x":0,"y":1}}],"coord":{"x":0,"y":0}}`)
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
			mi:           newMotionInfo(WorldCoord{0, 0}, North, 40),
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
