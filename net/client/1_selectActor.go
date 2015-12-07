package client

// An AuthenticatedConn is used to select the Actor
// that will be associated with the connection.
type AuthenticatedConn interface {
	// Will retrieve a list of the existing actors associated with a username.
	AvailableActors() ActorsRoundTrip

	// Will select the actor associated with this connection.
	// If the selected actor name doesn't exist, it will be created.
	SelectActor(name string) SelectActorRoundTrip
}

// ActorsRoundTrip is a Request->Response transaction to retrieve a list
// of actors associated with a User.
type ActorsRoundTrip struct{}

// SelectActorRoundTrip is a Request->Response transaction to select the
// Actor associated with the User's connection.
type SelectActorRoundTrip struct{}
