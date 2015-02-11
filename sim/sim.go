// The sim package defines the
// life cycle and interaction protocol used to
// interact with a game world simulation.
package sim

// A simulation that is prepared to begin simulating a world.
type UnstartedSimulation interface {
	Begin() (RunningSimulation, error)
}

// A simulation that is currently simulating a world.
type RunningSimulation interface {
	// A Running simulation can be halted
	Halt() (HaltedSimulation, error)
}

// A simulation that has been halted.
type HaltedSimulation interface {
}
