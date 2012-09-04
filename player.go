package engine

import (
	"../server/protocol"
	"strconv"
	"strings"
)

// Externaly Accessible Actions
type PlayerEntity interface {
	entity
	SubmitInput(cmd, params string) error
	Disconnect()
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
	// Movement speed in Frames per PathAction
	MovementSpeed uint

	// A Connection to send WorldState too
	Conn protocol.JsonOutputConn

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
	conn protocol.JsonOutputConn

	// Communication channels used inside the muxer
	collectInput    chan InputCmd
	serveMotionInfo chan *motionInfo
	routeWorldState chan WorldStateJson
	killMux         chan bool
}

type PlayerJson struct {
	Id          EntityId         `json:"id"`
	Name        string           `json:"name"`
	Facing      string           `json:"facing"`
	PathActions []PathActionJson `json:"pathActions"`
	Coord       WorldCoord       `json:"coord"`
}

func (p *Player) Id() EntityId {
	return p.entityId
}

func (p *Player) Coord() WorldCoord {
	return p.mi.coord
}

func (p *Player) AABB() (aabb AABB) {
	return p.mi.AABB()
}

func (p *Player) Json() interface{} {
	ps := PlayerJson{
		Id:     p.entityId,
		Name:   p.Name,
		Facing: p.mi.facing.String(),
		Coord:  p.mi.coord,
	}

	if len(p.mi.pathActions) > 0 {
		ps.PathActions = make([]PathActionJson, len(p.mi.pathActions))
		for i, pa := range p.mi.pathActions {
			ps.PathActions[i] = pa.Json()
		}
	}
	return ps
}

func (p *Player) mux() {
	p.collectInput = make(chan InputCmd)
	p.serveMotionInfo = make(chan *motionInfo)
	p.routeWorldState = make(chan WorldStateJson)
	p.killMux = make(chan bool)

	go func() {
		for {
			// Prioritize stopping the mux loop
			select {
			case <-p.killMux:
				return
			default:
			}

			select {
			case input := <-p.collectInput:
				switch input.cmd {
				case "move":
					p.mi.moveRequest = newMoveRequest(input)
				case "moveCancel":
					if p.mi.moveRequest != nil {
						if p.mi.moveRequest.Direction.String() == input.params {
							p.mi.moveRequest = nil
						}
					}
				default:
					panic("Unknown InputCmd: " + input.cmd)
				}

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
						p.conn.SendJson("update", worldState)
						break lockedMotionInfo
					case <-p.killMux:
						return
					}
				}
			case <-p.killMux:
				return
			}
		}
	}()
}

func (p *Player) stopMux() {
	p.killMux <- true
}

// External interface of the muxer presented to the simulation
func (p *Player) motionInfo() *motionInfo             { return <-p.serveMotionInfo }
func (p *Player) SendWorldState(state WorldStateJson) { p.routeWorldState <- state }

// Collision Handlers
func (p *Player) collides(other collidableEntity) (collides bool) {
	switch other.(type) {
	case *Player:
		collides = true
	}
	return
}

func (p *Player) collideWith(other collidableEntity, t WorldTime) {
	switch ce := other.(type) {
	case *Player:
		if p.mi.isMoving() && ce.mi.isMoving() {
			pa, paOther := p.mi.pathActions[0], ce.mi.pathActions[0]

			collision := pa.Collides(*paOther)

			switch collision.Type {
			case CT_FROM_SIDE, CT_HEAD_TO_HEAD:
				// Both started at the same time
				if pa.Start() == t && paOther.Start() == t && collision.T == t {

					// Higher speed means moving  slower
					// Whoever is faster wins
					if p.mi.speed >= ce.mi.speed {
						p.mi.UndoLastApply()
					}

				} else if pa.Start() == collision.T {
					// I started after
					p.mi.UndoLastApply()
				}
			case CT_SWAP:
				p.mi.UndoLastApply()
			}

		} else if p.mi.isMoving() && !ce.mi.isMoving() {
			pa := p.mi.pathActions[0]
			// Attempting to move onto an occupied location
			if pa.Dest == ce.Coord() {
				p.mi.UndoLastApply()
			}
		}
	}
}

// External interface of the muxer presented to the Node
func (p *Player) SubmitInput(cmd, params string) error {
	parts := strings.Split(cmd, "=")

	timeIssued, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}

	p.collectInput <- InputCmd{
		timeIssued: WorldTime(timeIssued),
		cmd:        parts[0],
		params:     params,
	}
	return nil
}

func (p *Player) Disconnect() {
	p.sim.RemovePlayer(p)
}
