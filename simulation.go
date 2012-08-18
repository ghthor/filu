package engine

import (
	"fmt"
	"sync"
	"time"
)

/*
The Client generates InputEvents and sends them to the server.  The Node processes these input events and packages them up for the simulation. When the simulation runs it gathers all the Clients InputEvent's and Currently executing Actions and steps simulates a step forward in time generating the next WorldState.  This WorldState is then published to all the clients via a channel
*/
type StateConn interface {
	SendWorldState(*WorldState)
}

type (
	WorldState struct {
		processingTime time.Duration
		time           WorldTime
		entities       map[EntityId]Entity
	}

	Simulation struct {
		clock        Clock
		nextEntityId EntityId
		state        *WorldState

		clients   []StateConn
		newPlayer chan PlayerDef

		fps int

		stop    chan bool
		running bool
		sync.Mutex
	}
)

func (ws *WorldState) String() string {
	return fmt.Sprintf(":%v\tEntities:%v", ws.time, len(ws.entities))
}

func NewSimulation(fps int) *Simulation {
	clk := Clock(0)

	s := &Simulation{
		clock: clk,

		state: &WorldState{
			time:     clk.Now(),
			entities: make(map[EntityId]Entity, 10),
		},

		clients:   make([]StateConn, 0, 10),
		newPlayer: make(chan PlayerDef),

		fps: fps,

		stop: make(chan bool),
	}
	return s
}

func (s *Simulation) Start() {
	s.startLoop()
}

func (s *Simulation) startLoop() {

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
				playerDef.NewPlayer <- s.addPlayer(playerDef)
			case <-s.stop:
				break stepLoop

			}
		}

		s.Lock()
		s.running = false
		s.Unlock()
		s.stop <- true
	}()
}

func (s *Simulation) step() {
	s.clock = s.clock.Tick()
	s.state = s.state.StepTo(s.clock.Now())
	for _, c := range s.clients {
		c.SendWorldState(s.state)
	}
}

func (s *Simulation) Stop() {
	s.stop <- true
	<-s.stop
}

func (s *Simulation) IsRunning() bool {
	s.Lock()
	defer s.Unlock()
	return s.running
}

func (s *Simulation) AddPlayer(pd PlayerDef) *Player {
	pd.NewPlayer = make(chan *Player)
	s.newPlayer <- pd
	return <-pd.NewPlayer
}

func (s *Simulation) addPlayer(pd PlayerDef) *Player {
	p := &Player{
		entityId:   s.nextEntityId,
		Name:       pd.Name,
		motionInfo: NewMotionInfo(pd.Coord, pd.Facing),
	}
	s.nextEntityId++

	s.state.entities[p.entityId] = p
	return p
}

func (s *Simulation) AddClient(c StateConn) {
}

func (s *WorldState) StepTo(t WorldTime) *WorldState {
	// Collect Pending Actions

	// Filter out Actions that are impossible - Walking into a wall

	// Filter Conflicting Actions - 2 Characters moving in to the same Coord
	// Resolve Conflicting Actions

	// Apply All remaining Pending Actions as Running Actions
	// Write out new state and return
	return s
}
