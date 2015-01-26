# A Simulation

A simulation, in engine, is the running kernel that is simulating a world.
It is the go routine that is receiving ticks from the clock, taking input
from the actors and applying it to the simulated world, then calculating 
the world state of the next frame and sending it back out to the actors.
It is the game loop.

A simulation has a life cycle. It begins as an
[UnstartedSimulation](http://godoc.org/github.com/ghthor/engine/sim#UnstartedSimulation).
An unstarted simulation has only one behavior, it can Begin.

Once the simulation has begun, it is a
[RunningSimulation](http://godoc.org/github.com/ghthor/engine/sim#RunningSimulation).
A running simulation can connect and remove actors. It can also halt
which stops the simulation.

A [HaltedSimulation](http://godoc.org/github.com/ghthor/engine/sim#HaltedSimulation)
has no defined behavior yet. But it could perhaps contain stats and analytics.
Maybe all the Actors that remained connected through the Halt.

[rpg2d](rpg2d) is an implementation of the Simulation interface and life cycle.

### Process & Go routine's of a running Simulation

![dia diagram of the process architecture](work-journal/design-notes/process-architecture.png)

# Work journal

A journal to store work log of [entries](work-journal/entry/)
and [ideas](work-journal/idea/).  Reading the latest entry will give
some context to what's currently being worked on, what's working,
what's not working.
