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
	Cell Cell
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
	// TODO When go updates to 1.1? this can be converted to an embedded type
	// The current json marshaller doesn't marshall embedded types
	EntityId    EntityId         `json:"id"`
	Name        string           `json:"name"`
	Facing      string           `json:"facing"`
	PathActions []PathActionJson `json:"pathActions"`
	Cell        Cell             `json:"cell"`
}

func (p *Player) Id() EntityId {
	return p.entityId
}

func (p *Player) Cell() Cell {
	return p.mi.cell
}

func (p *Player) AABB() (aabb AABB) {
	return p.mi.AABB()
}

func (p *Player) Json() EntityJson {
	ps := PlayerJson{
		EntityId: p.entityId,
		Name:     p.Name,
		Facing:   p.mi.facing.String(),
		Cell:     p.mi.cell,
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

			collision := pathCollision(*pa, *paOther)

			switch collision.CollisionType {
			case CT_FROM_SIDE:
				// If my movement starts the collision
				if pa.Start() == collision.Start() {
					if paOther.Start() < collision.Start() {
						// and the other is already moving, they've claimed the destination already
						p.mi.UndoLastApply()

					} else if paOther.Start() == collision.Start() && paOther.End() < collision.End() {
						// and the other moves faster then us
						p.mi.UndoLastApply()
					} else if pa.Start() == paOther.Start() && pa.End() == paOther.End() {
						// Same Speed, Same Start, I lose due to order of collision execution
						p.mi.UndoLastApply()
					}
				}

			case CT_HEAD_TO_HEAD:
				// other started before us and therefore has claimed the position
				if pa.Start() == t && paOther.Start() < t {
					p.mi.UndoLastApply()
				} else if pa.Start() == t && paOther.Start() == t {
					// we've started at the same time, whoever finishs first wins the destination
					if pa.End() == collision.End() {
						// Side effect of pa.End() == paOther.End() is paOther wins the random faceoff
						p.mi.UndoLastApply()
					}
				}

			case CT_SWAP:
				p.mi.UndoLastApply()

			case CT_A_INTO_B:
				// If I'm A then I don't move
				if collision.A == *pa {
					p.mi.UndoLastApply()
				}

			case CT_A_INTO_B_FROM_SIDE:
				// If I'm A then I can't move until my movement ends with or after collision ending
				if collision.A == *pa {
					if pa.End() < collision.End() {
						p.mi.UndoLastApply()
					}
				}
			}

		} else if p.mi.isMoving() && !ce.mi.isMoving() {
			pa := p.mi.pathActions[0]
			// Attempting to move onto an occupied location
			if pa.Dest == ce.Cell() {
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

func (p PlayerJson) Id() EntityId { return p.EntityId }
func (p PlayerJson) AABB() AABB   { return AABB{p.Cell, p.Cell} }
func (p PlayerJson) IsDifferentFrom(other EntityJson) (different bool) {
	o := other.(PlayerJson)

	switch {
	case p.Facing != o.Facing:
		fallthrough
	case len(p.PathActions) != len(o.PathActions):
		different = true
	case len(p.PathActions) == len(o.PathActions):
		for i, _ := range o.PathActions {
			different = different || (p.PathActions[i] != o.PathActions[i])
		}
	}
	return
}
