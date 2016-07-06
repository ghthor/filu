package client

import (
	"github.com/ghthor/filu"
	"github.com/ghthor/filu/net"
)

/* A SelectedActorConn can take an Actor and connect that
Actor into a simulation over a connection. This extra step after
selection allows the client to control when the Actor's entity will
appear within the simulation. A well behaving client will be prepared
to display the world(all graphical resources loaded) surrounding the
Actor before asking to connect the Actor. In simulations where the
Actor's entity isn't present in the world when the Actor isn't connected,
This ensures that the entity isn't vulnerable before the Actor's
User is able to view and interact with the world. */
type SelectedActorConn interface {
	Actor() filu.Actor
	Connect() ConnectRoundTrip
}

type selectedActorConn struct {
	conn  net.Conn
	actor filu.Actor
}

func NewSelectedActorConn(conn net.Conn, actor filu.Actor) SelectedActorConn {
	return selectedActorConn{conn, actor}
}

func (c selectedActorConn) Actor() filu.Actor { return c.actor }

func (c selectedActorConn) Connect() ConnectRoundTrip {
	return ConnectRoundTrip{}
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
	net.Conn
	filu.Actor
}
