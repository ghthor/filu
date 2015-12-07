package client

import (
	"github.com/ghthor/filu"
	"github.com/ghthor/filu/net"
)

// An ActorConn can take a selected Actor and connect that
// Actor into a simulation. This extra step after selection
// allows the client to control when the Actor's entity may
// appear within the simulation. A well behaving client will
// be prepared to display the world(all graphical resources loaded)
// surrounding the Actor before asking to connect the Actor.
// In games where the Actor's entity isn't present in the world
// when the Actor isn't connected, this ensures that the entity
// isn't vulnerable before the Actor's User is able to view
// and interact with the world.
type ActorConn interface {
	ConnectActor(filu.Actor) ConnectRoundTrip
}

// ConnectRountTrip is a Request->Response transaction to connect
// the selected Actor into a running game simulation.
type ConnectRoundTrip struct{}

// A ConnectedActorConn represents an Actor that has been connected
// into a game's simulation. A connected Actor will receive world state
// changes and can interact with the world using input commands. Access
// to a raw net.Conn is provide to enable a game's implementation to
// expand the functionality available to a connected Actor.
type ConnectedActorConn struct {
	filu.Actor

	net.Conn
}
