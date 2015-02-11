package rpg2d

import (
	"errors"
	"time"

	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/quad"
	"github.com/ghthor/engine/sim"
	"github.com/ghthor/engine/sim/stime"
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
	Now      stime.Time
	QuadTree quad.Quad

	// User defined to resolve an entity from and actor
	EntityResolver EntityResolver

	// User defined input application phase
	InputPhaseHandler quad.InputPhaseHandler

	// User defined the narrow phase
	NarrowPhaseHandler quad.NarrowPhaseHandler
}

type initialWorldState struct {
	now      stime.Time
	quadTree quad.Quad
}

type simSettings struct {
	fps int

	EntityResolver
	quad.InputPhaseHandler
	quad.NarrowPhaseHandler
}

// An implementation of engine/sim.RunningSimulation
type runningSimulation struct {
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

var ErrMustProvideAQuadtree = errors.New("user must provide a quad tree to a simulation defination")

// Implement engine/sim.UnstartedSimulation
func (s SimulationDef) Begin() (sim.RunningSimulation, error) {
	if s.QuadTree == nil {
		return nil, ErrMustProvideAQuadtree
	}

	initialState := initialWorldState{
		now:      s.Now,
		quadTree: s.QuadTree,
	}

	settings := simSettings{
		s.FPS,

		s.EntityResolver,
		s.InputPhaseHandler,
		s.NarrowPhaseHandler,
	}

	rs := &runningSimulation{}

	// Starts 2 go routines and returns
	// The ticker and the engine communication kernel
	rs.startLoop(initialState, settings)

	return rs, nil
}

// Prepares a closure and executes it as a go routine
// Calling this function will create 2 infinte looping
// go routines. One for the clock ticker and one for
// calulating the next world state and adding/removing actors.
// This method has a pointer recv because it MUST set the
// addActor and removeActor communication channels used
// by the public api to request adding & removing actors
func (s *runningSimulation) startLoop(initialState initialWorldState, settings simSettings) {
	//---- Create all the communication channels

	// Make the 2way channels that will be used to make
	// add and remove actor requests to the go routine game loop
	addCh := make(chan addActorReq)
	removeCh := make(chan removeActorReq)

	// Set the 1way send channels used by the public api
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

	quadTree := initialState.quadTree
	clock := stime.Clock(initialState.now)

	//---- User provided actor to entity resolver
	entityResolver := settings.EntityResolver

	//---- User provided input application phase
	inputPhase := settings.InputPhaseHandler

	//---- User provided narrow phase
	narrowPhase := settings.NarrowPhaseHandler

	runTick := func(q quad.Quad, t stime.Time) quad.Quad {
		return quad.RunPhasesOn(q, inputPhase, narrowPhase, t)
	}

	// Start the Clock
	ticker := time.NewTicker(time.Duration(1000/settings.fps) * time.Millisecond)

	// Start the simulation server
	go func() {
		var hasHalted chan<- struct{}

	communicationLoop:
		// # This select prioritizes the following communication events
		// ## 2 potential events to respond to
		// 1. Trigger a simulation tick
		// 2. Halt() method has requested halting
		select {
		case <-ticker.C:
			goto tick

		case hasHalted = <-haltReq:
			goto exit
		default:
		}

		// ## 2 potential events to respond to
		// 1. Trigger a simulation tick
		// 2. ConnectActor() method has requested to connect an actor
		// 3. RemoveActor() method has requested to remove an actor
		// 4. Halt() method has requested halting
		select {
		case <-ticker.C:
			goto tick

		case actor := <-addReq:
			// a is the new sim.Actor{} to be inserted into the sim
			a := <-actor.toBeAdded

			e := entityResolver.EntityForActor(a)
			quadTree = quadTree.Insert(e)

			// signal that the operation was a success
			actor.wasAdded <- a

			goto communicationLoop

		case actor := <-removeReq:
			// a is the new sim.Actor{} to be removed from the sim
			a := <-actor.toBeRemoved

			// TODO removed the actor from the simulation

			// signal that the operation was a success
			actor.wasRemoved <- a

			goto communicationLoop

		case hasHalted = <-haltReq:
			goto exit
		}

		panic("unclosed case in simulation communication loop select case")

	tick:
		clock = clock.Tick()
		quadTree = runTick(quadTree, clock.Now())
		// TODO quadTree to state
		// TODO send state
		goto communicationLoop

	exit:
		// TODO pre halt cleanup
		// Signal to Halt() caller that we've finished cleanup
		hasHalted <- struct{}{}
	}()
}

type haltedSimulation struct{}
