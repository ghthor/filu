// The sim package defines the interface used to
// interact with a game world simulation.
// It also defines the IO interface the
// simulation can use to communicate with the actors.
package sim

// An Actor can be connected to the simulation.
type Actor interface {
	// An actor has a unique Id that it is assigned
	// when it is created.
	Id() int64

	// An actor provides a connection the simulation
	// will read input from and write world state to.
	StateWriter() StateWriter
}

// A StateWriter can receive the state of the simulation.
type StateWriter interface {
	// WriteState() takes an empty interface{} because
	// it is expected the implementation will only be
	// encoding the state and writing the it to
	// the network socket.
	WriteState(interface{}) error
}

// A simulation that is prepared to begin simulating a world.
type UnstartedSimulation interface {
	Begin() (RunningSimulation, error)
}

// A simulation that is currently simulating a world.
type RunningSimulation interface {
	// A running simulation will accept connections from
	// actor. The actor will be added into the simulation.
	ConnectActor(Actor) error

	// A Running simulation can remove an actor. The actor
	// will stop receiving world states.
	RemoveActor(Actor) error

	// A Running simulation can be halted
	Halt() (HaltedSimulation, error)
}

// A simulation that has been halted.
type HaltedSimulation interface {
}
