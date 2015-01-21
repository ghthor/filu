package rpg2d

import "github.com/ghthor/engine/sim"

// A SimulationDef used to configure a simulation
// to define the how the simulation will behave.
type SimulationDef struct {
}

// Implement engine/sim.UnstartedSimulation
func (s SimulationDef) Begin() (sim.RunningSimulation, error) {
	return runningSimulation{}, nil
}

// An implementation of engine/sim.RunningSimulation
type runningSimulation struct{}

// Implement engine/sim.RunningSimulation
func (s runningSimulation) ConnectActor(sim.Actor) error {
	return nil
}

// Implement engine/sim.RunningSimulation
func (s runningSimulation) RemoveActor(sim.Actor) error {
	return nil
}

// Implement engine/sim.RunningSimulation
func (s runningSimulation) Halt() (sim.HaltedSimulation, error) {
	return haltedSimulation{}, nil
}

type haltedSimulation struct{}
