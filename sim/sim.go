// The sim package defines the interface to a simulation.
// It also defines how the input/output interface the
// simulation has to the actors.
package sim

// An Actor can be connected to the simulation.
type Actor interface {
	// An actor has a unique Id that it is assigned
	// when it is created.
	Id() int64

	// An actor provides a connection the simulation
	// will read input from and write world state to.
	Conn() ActorConn
}

// An ActorConn is the interface to the network
// socket that the simulation will use.
type ActorConn interface {
	// An actor can provide an input. This will
	// be called ONCE per frame per actor.
	// TODO string type may change, but for now it is fine
	ReadInput() (string, error)

	// An actor will receive the state of the world
	// after each calculation.
	StateWriter
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
