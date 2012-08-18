package engine

import (
	"../server/protocol"
	"../server/protocol/encoding"
	"fmt"
)

type (
	EntityId int64

	entity interface {
		Id() EntityId
	}

	moveRequest struct {
		t WorldTime
		Direction
	}

	motionInfo struct {
		coord  WorldCoord
		facing Direction

		moveRequest *moveRequest

		// fifo
		pathActions []*PathAction
	}

	movableEntity interface {
		entity
		motionInfo() *motionInfo
	}
)

func newMotionInfo(c WorldCoord, f Direction) *motionInfo {
	return &motionInfo{
		c,
		f,
		nil,
		make([]*PathAction, 0, 2),
	}
}

func (mi motionInfo) isMoving() bool {
	return len(mi.pathActions) == 0
}

// This object is used to create a player
// All the Fields must be provided
type PlayerDef struct {

	// The Players Name
	Name string

	// Where the Player is locationed
	Coord WorldCoord
	// Which Direction the Player is Facing
	Facing Direction

	// A Connection to send WorldState too
	Conn protocol.MessageOutputConn

	// This is Used internally by the simulation to return the new
	// Player object after is has been created
	newPlayer chan *Player
}

type Player struct {
	// The Players Name
	Name     string
	entityId EntityId
	mi       *motionInfo

	// A handle to the simulation this player is in
	sim Simulation

	// A Connection to send WorldState too
	conn protocol.MessageOutputConn

	// Communication channels used inside the muxer
	collectInput    chan encoding.Packet
	serveMotionInfo chan *motionInfo
	routeWorldState chan *WorldState
}

func (p *Player) Id() EntityId {
	return p.entityId
}

func (p *Player) mux() {
	p.collectInput = make(chan encoding.Packet)
	p.serveMotionInfo = make(chan *motionInfo)
	p.routeWorldState = make(chan *WorldState)

	go func() {
		for {
			select {
			case input := <-p.collectInput:
				// Process Input
				// Create ActionReq's, like a MoveRequest
				fmt.Println(input)

			// The simulation has requested access to the motionInfo
			// This 'locks' the motionInfo until the server publishs a WorldState
			case p.serveMotionInfo <- p.mi:
			lockedMotionInfo:
				for {
					select {
					//case input := <-p.collectInput:
					// Buffer all input to be processed after WorldState is published
					case p.serveMotionInfo <- p.mi:
					case worldState := <-p.routeWorldState:
						// Take the worldState and cut out anything the client shouldn't know
						// Package up this localized WorldState and send it over the wire
						p.conn.SendMessage("worldState", worldState.String())
						break lockedMotionInfo
					}
				}
			}
		}
	}()
}

// External interface of the muxer presented to the simulation
func (p *Player) motionInfo() *motionInfo          { return <-p.serveMotionInfo }
func (p *Player) SendWorldState(state *WorldState) { p.routeWorldState <- state }

// External interface of the muxer presented to the Node
func (p *Player) SubmitInput(packet encoding.Packet) { p.collectInput <- packet }