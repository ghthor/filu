package engine

import (
	"../server/protocol"
	"fmt"
	"sync"
	"time"
)

/*
The Client generates InputEvents and sends them to the server.  The Node processes these input events and packages them up for the simulation. When the simulation runs it gathers all the Clients InputEvent's and Currently executing Actions and steps simulates a step forward in time generating the next WorldState.  This WorldState is then published to all the clients via a channel
*/
type (
	// Internal format used by the simulation
	WorldState struct {
		processingTime time.Duration
		time           WorldTime
		quadTree       quad
	}

	// External format used to send state to the clients
	WorldStateJson struct {
		Time     WorldTime    `json:"time"`
		Entities []EntityJson `json:"entities"`
		Removed  []EntityJson `json:"removed"`
	}

	StateConn interface {
		SendWorldState(WorldStateJson)
	}

	DiffConn struct {
		protocol.JsonOutputConn
		lastState WorldStateJson
	}
)

func (c *DiffConn) SendJson(msg string, nextState interface{}) error {
	diff := c.lastState.Diff(nextState.(WorldStateJson))
	c.lastState = nextState.(WorldStateJson)

	if len(diff.Entities) > 0 || len(diff.Removed) > 0 {
		c.JsonOutputConn.SendJson(msg, diff)
	}
	return nil
}

type (
	Simulation interface {
		Start()
		IsRunning() bool
		Stop()

		AddPlayer(PlayerDef) PlayerEntity
		RemovePlayer(PlayerEntity)
		AddClient(StateConn)
	}

	simulation struct {
		clock        Clock
		nextEntityId EntityId
		state        *WorldState

		clients []StateConn

		// Comm channels for adding/removing Players
		newPlayer  chan PlayerDef
		dcedPlayer chan dcedPlayer

		fps int

		stop    chan bool
		running bool
		sync.Mutex
	}
)

func (ws *WorldState) String() string {
	return fmt.Sprintf(":%v\tEntities:%v", ws.time, len(ws.quadTree.QueryAll(ws.quadTree.AABB())))
}

func (ws *WorldState) AddMovableEntity(e movableEntity) {
	ws.quadTree = ws.quadTree.Insert(e)
}

func NewSimulation(fps int) Simulation {
	return newSimulation(fps)
}

func newWorldState(clock Clock) *WorldState {
	quadTree, err := newQuadTree(AABB{
		Cell{-1000, 1000},
		Cell{1000, -1000},
	}, nil, 20)

	if err != nil {
		panic("error creating quadTree: " + err.Error())
	}
	return &WorldState{
		time:     clock.Now(),
		quadTree: quadTree,
	}
}

func newSimulation(fps int) *simulation {
	clk := Clock(0)

	s := &simulation{
		clock: clk,

		state: newWorldState(clk),

		clients: make([]StateConn, 0, 10),

		newPlayer:  make(chan PlayerDef),
		dcedPlayer: make(chan dcedPlayer),

		fps: fps,

		stop: make(chan bool),
	}
	return s
}

func (s *simulation) Start() {
	s.startLoop()
}

func (s *simulation) startLoop() {

	ticker := time.NewTicker(time.Duration(1000/s.fps) * time.Millisecond)

	s.Lock()

	go func() {
		s.running = true
		s.Unlock()

	stepLoop:
		for {
			s.step()

			select {
			case <-s.stop:
				break stepLoop
			default:
			}

			select {
			case <-ticker.C:
			case playerDef := <-s.newPlayer:
				playerDef.newPlayer <- s.addPlayer(playerDef)
			case dced := <-s.dcedPlayer:
				dced.removed <- s.removePlayer(dced.player)
			case <-s.stop:
				break stepLoop

			}
		}

		s.Lock()
		s.running = false
		s.Unlock()
		s.stop <- true

		ticker.Stop()
	}()
}

func (s *simulation) step() {
	s.clock = s.clock.Tick()
	s.state = s.state.stepTo(s.clock.Now())
	for _, c := range s.clients {
		c.SendWorldState(s.state.Json())
	}
}

func (s *simulation) Stop() {
	s.stop <- true
	<-s.stop
}

func (s *simulation) IsRunning() bool {
	s.Lock()
	defer s.Unlock()
	return s.running
}

func (s *simulation) AddPlayer(pd PlayerDef) PlayerEntity {
	pd.newPlayer = make(chan *Player)
	s.newPlayer <- pd
	return <-pd.newPlayer
}

func (s *simulation) addPlayer(pd PlayerDef) *Player {
	p := &Player{
		Name: pd.Name,

		entityId: s.nextEntityId,
		mi:       newMotionInfo(pd.Cell, pd.Facing, pd.MovementSpeed),
		sim:      s,
		conn:     &DiffConn{JsonOutputConn: pd.Conn},
	}
	s.nextEntityId++

	// Add the Player entity
	s.state.AddMovableEntity(p)

	// Add the Player as a client that recieves WorldState's
	s.clients = append(s.clients, p)

	// Start the muxer
	p.mux()
	return p
}

type dcedPlayer struct {
	player  *Player
	removed chan *Player
}

func (s *simulation) RemovePlayer(p PlayerEntity) {
	dced := dcedPlayer{
		p.(*Player),
		make(chan *Player),
	}
	s.dcedPlayer <- dced
	<-dced.removed
	return
}

func (s *simulation) removePlayer(p *Player) *Player {
	// Stop the muxer
	p.stopMux()

	// Remove the Player entity
	s.state.quadTree.Remove(p)

	// Remove the client
	for i, c := range s.clients {
		if c == p {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
	return p
}

func (s *simulation) AddClient(c StateConn) {
}

func (s *WorldState) step() *WorldState {
	return s.stepTo(s.time + 1)
}

func (s *WorldState) stepTo(t WorldTime) *WorldState {
	s.quadTree.AdjustPositions(t)
	s.quadTree.StepTo(t)

	s.time = t
	return s
}

func (ws *WorldState) Json() WorldStateJson {
	entities := ws.quadTree.QueryAll(ws.quadTree.AABB())
	s := WorldStateJson{
		ws.time,
		make([]EntityJson, len(entities)),
		nil,
	}

	i := 0
	for _, e := range entities {
		s.Entities[i] = e.Json()
		i++
	}
	return s
}

func (s WorldStateJson) Clone() WorldStateJson {
	clone := WorldStateJson{
		s.Time,
		make([]EntityJson, len(s.Entities)),
		nil,
	}
	copy(clone.Entities, s.Entities)
	return clone
}

func (s WorldStateJson) Diff(ss WorldStateJson) (diff WorldStateJson) {
	diff.Time = ss.Time

	if len(s.Entities) == 0 && len(ss.Entities) > 0 {
		return ss
	}

	// Find the entities that have changed from the old state to the new one
nextEntity:
	for _, entity := range ss.Entities {
		for _, old := range s.Entities {
			if entity.Id() == old.Id() {
				if old.IsDifferentFrom(entity) {
					diff.Entities = append(diff.Entities, entity)
				}
				continue nextEntity
			}
		}
		// This is a new Entity
		diff.Entities = append(diff.Entities, entity)
	}

	// Check if all the entities in old state exist in the new state
entityStillExists:
	for _, old := range s.Entities {
		for _, entity := range ss.Entities {
			if old.Id() == entity.Id() {
				continue entityStillExists
			}
		}
		diff.Removed = append(diff.Removed, old)
	}
	return
}
