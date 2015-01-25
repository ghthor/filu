package rpg2d

import (
	"time"

	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/quad"
	"github.com/ghthor/engine/sim"
)

// An interface the user will implement to resolve
// an entity from an actor. This is user defined
// because the user is creating the entities and actors
// that will be added to the simulation. This allows the
// user to define how the actors state is stored,
// aka database design/interaction.
type EntityResolver interface {
	EntityForActor(sim.Actor) entity.Entity
}

// A SimulationDef used to configure a simulation
// to define the how the simulation will behave.
type SimulationDef struct {
	// The target FPS for the simulation to calculate at
	FPS int

	// Initial World State
	QuadTree quad.Quad

	// User defined to resolve an entity from and actor
	EntityResolver EntityResolver

	// User defined input application phase
	InputPhaseHandler quad.InputPhaseHandler

	// User defined the narrow phase
	NarrowPhaseHandler quad.NarrowPhaseHandler
}

// An implementation of engine/sim.RunningSimulation
type runningSimulation struct {
	fps int

	//---- World state
	quadTree quad.Quad

	//---- User Settings
	EntityResolver

	//---- Communication

	// These channels are used by the public api
	// to add and remove actors. They are 1way
	// send only channels. The requests contains
	// a send only and recieve only channel so the
	// the public api call will be atomic action
	// and the caller can assume without a doubt
	// that the actor is now added or removed
	// from the simulation.
	addActor    chan<- addActorReq
	removeActor chan<- removeActorReq

	// This channel is used by the public api to request
	// that the simulation is halted. You send a 1way
	// send only channel into the game loop so the public
	// api can wait and be notified that the go routine
	// has returned and is no longer running.
	requestHalt chan<- chan<- struct{}
}

// Communication object used to atomicly add a new actor to the sim
type addActorReq struct {
	toBeAdded chan sim.Actor
	wasAdded  chan sim.Actor
}

// Implement engine/sim.RunningSimulation
func (s runningSimulation) ConnectActor(a sim.Actor) error {
	ch := make(chan sim.Actor)

	// Create an add request
	actor := addActorReq{ch, ch}

	// Send the add request to the game loop
	s.addActor <- actor

	// Send the actor to be added to the game loop
	actor.toBeAdded <- a

	// Wait for the add request to be successfully completed
	a = <-actor.wasAdded
	return nil
}

// Communication object used to atomicly remove an actor from the sim
type removeActorReq struct {
	toBeRemoved chan sim.Actor
	wasRemoved  chan sim.Actor
}

// Implement engine/sim.RunningSimulation
func (s runningSimulation) RemoveActor(a sim.Actor) error {
	ch := make(chan sim.Actor)

	// Create a remove request
	actor := removeActorReq{ch, ch}

	// Send the remove request to the game loop
	s.removeActor <- actor

	// Send the actor to be removed to the game loop
	actor.toBeRemoved <- a

	// Wait for the remove request to be successfully completed
	a = <-actor.wasRemoved
	return nil
}

// Implement engine/sim.RunningSimulation
func (s runningSimulation) Halt() (sim.HaltedSimulation, error) {
	wasHalted := make(chan struct{})

	// Send a request to the game loop to halt
	s.requestHalt <- wasHalted

	// Wait for the halt request to be successfully completed
	<-wasHalted

	return haltedSimulation{}, nil
}

// Implement engine/sim.UnstartedSimulation
func (s SimulationDef) Begin() (sim.RunningSimulation, error) {
	rs := &runningSimulation{
		fps: s.FPS,

		//---- World State
		quadTree: s.QuadTree,

		//---- User Settings
		EntityResolver:     s.EntityResolver,

		//---- Communication
		// All initialized within startLoop()
	}

	// Starts 2 go routines and returns
	// The Ticker and the Communication Game Loop
	rs.startLoop()

	return rs, nil
}

// Prepares a closure and executes it as a go routine
// Calling this function will create 2 infinte looping
// go routines. One for the clock ticker and one for
// calulating the next world state and adding/removing actors.
// This method has a pointer recv because it MUST set the
// addActor and removeActor communication channels used
// by the public api to request adding & removing actors
func (s *runningSimulation) startLoop() {
	ticker := time.NewTicker(time.Duration(1000/s.fps) * time.Millisecond)

	// Make the 2way channels that will be used to make
	// add and remove actor requests to the go routine game loop
	addCh := make(chan addActorReq)
	removeCh := make(chan removeActorReq)

	// Set the 1way send chanels used by the public api
	s.addActor = addCh
	s.removeActor = removeCh

	// Set the 1way recieve channels used by the game loop
	var addReq <-chan addActorReq
	var removeReq <-chan removeActorReq

	addReq = addCh
	removeReq = removeCh

	// Make channel to be used to by the public api to
	// request that the simulation be halted
	haltCh := make(chan chan<- struct{})

	// Set the 1way send channel used by the public api
	s.requestHalt = haltCh

	// Set the 1way recieve channel used by the game loop
	var haltReq <-chan chan<- struct{}
	haltReq = haltCh

	go func() {
		var hasHalted chan<- struct{}

	gameLoop:
		for {
			// Prioritized select for ticker.C and haltReq
			select {
			case <-ticker.C:
				// TODO Step the simulation forward in time

			case hasHalted = <-haltReq:
				break gameLoop
			default:
			}

			select {
			case <-ticker.C:
				// TODO Step the simulation forward in time

			case actor := <-addReq:
				// a is the new sim.Actor{} to be inserted into the sim
				a := <-actor.toBeAdded

				e := s.EntityForActor(a)
				s.quadTree = s.quadTree.Insert(e)

				// signal that the operation was a success
				actor.wasAdded <- a

			case actor := <-removeReq:
				// a is the new sim.Actor{} to be removed from the sim
				a := <-actor.toBeRemoved

				// TODO removed the actor from the simulation

				// signal that the operation was a success
				actor.wasRemoved <- a

			case hasHalted = <-haltReq:
				break gameLoop
			}
		}

		// We're done with cleanup and going to exit
		hasHalted <- struct{}{}
	}()
}

type haltedSimulation struct{}
