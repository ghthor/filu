package engine

import (
	"strconv"
	"strings"

	. "github.com/ghthor/engine/rpg2d/coord"
	. "github.com/ghthor/engine/time"
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
	Conn JsonOutputConn

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
	conn JsonOutputConn

	// Communication channels used inside the muxer
	collectInput    chan InputCmd
	serveMotionInfo chan *motionInfo
	routeWorldState chan WorldStateJson
	killMux         chan bool
}

func (p *Player) Id() EntityId {
	return p.entityId
}

func (p *Player) Cell() Cell {
	return p.mi.cell
}

func (p *Player) AABB() (aabb Bounds) {
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
						worldState = p.CullStateToView(worldState)
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

const viewPortSize = 52

func (p *Player) CullStateToView(s WorldStateJson) WorldStateJson {
	s = s.Cull(Bounds{
		p.mi.cell.Add(-viewPortSize/2, viewPortSize/2),
		p.mi.cell.Add(viewPortSize/2, -viewPortSize/2),
	})
	return s
}

// Collision Handlers
func (p *Player) collides(other collidableEntity) (collides bool) {
	switch other.(type) {
	case *Player:
		collides = true
	}
	return
}

func (p *Player) collideWith(other collidableEntity, t Time) {
	// switch ce := other.(type) {
	// case *Player:
	// 	if p.mi.isMoving() && ce.mi.isMoving() {
	// 		pa, paOther := p.mi.pathActions[0], ce.mi.pathActions[0]

	// 		collision := pathCollision(*pa, *paOther)

	// 		switch collision.CollisionType {
	// 		case CT_FROM_SIDE:
	// 			// If my movement starts the collision
	// 			if pa.Start() == collision.Start() {
	// 				if paOther.Start() < collision.Start() {
	// 					// and the other is already moving, they've claimed the destination already
	// 					p.mi.UndoLastApply()

	// 				} else if paOther.Start() == collision.Start() && paOther.End() < collision.End() {
	// 					// and the other moves faster then us
	// 					p.mi.UndoLastApply()
	// 				} else if pa.Start() == paOther.Start() && pa.End() == paOther.End() {
	// 					// Same Speed, Same Start, I lose due to order of collision execution
	// 					p.mi.UndoLastApply()
	// 				}
	// 			}

	// 		case CT_HEAD_TO_HEAD:
	// 			// other started before us and therefore has claimed the position
	// 			if pa.Start() == t && paOther.Start() < t {
	// 				p.mi.UndoLastApply()
	// 			} else if pa.Start() == t && paOther.Start() == t {
	// 				// we've started at the same time, whoever finishs first wins the destination
	// 				if pa.End() == collision.End() {
	// 					// Side effect of pa.End() == paOther.End() is paOther wins the random faceoff
	// 					p.mi.UndoLastApply()
	// 				}
	// 			}

	// 		case CT_SWAP:
	// 			p.mi.UndoLastApply()

	// 		case CT_A_INTO_B:
	// 			// If I'm A then I don't move
	// 			if collision.A == *pa {
	// 				p.mi.UndoLastApply()
	// 			}

	// 		case CT_A_INTO_B_FROM_SIDE:
	// 			// If I'm A then I can't move until my movement ends with or after collision ending
	// 			if collision.A == *pa {
	// 				if pa.End() < collision.End() {
	// 					p.mi.UndoLastApply()
	// 				}
	// 			}
	// 		}

	// 	} else if p.mi.isMoving() && !ce.mi.isMoving() {
	// 		pa := p.mi.pathActions[0]
	// 		// Attempting to move onto an occupied location
	// 		if pa.Dest == ce.Cell() {
	// 			p.mi.UndoLastApply()
	// 		}
	// 	}
	// }
}

// External interface of the muxer presented to the Node
func (p *Player) SubmitInput(cmd, params string) error {
	parts := strings.Split(cmd, "=")

	timeIssued, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}

	p.collectInput <- InputCmd{
		timeIssued: Time(timeIssued),
		cmd:        parts[0],
		params:     params,
	}
	return nil
}

func (p *Player) Disconnect() {
	p.sim.RemovePlayer(p)
}
