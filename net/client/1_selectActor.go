package client

import (
	"fmt"

	"github.com/ghthor/filu"
	"github.com/ghthor/filu/net"
)

// ActorsListConn is used to retrieve the list of actors
// available to a user. The result is a SelectActorConn.
type ActorsListConn interface {
	GetActors() GetActorsRoundTrip
}

type actorsList struct {
	conn net.Conn
}

func NewActorsListConn(conn net.Conn) ActorsListConn {
	return actorsList{conn: conn}
}

func (c actorsList) GetActors() GetActorsRoundTrip {
	return GetActorsRoundTrip{conn: c.conn}.run()
}

// GetActorsRoundTrip is used to retrieve the result of a
// GetActors() request. The result is a connection that
// expects an actor to be selected from those available
// or created.
type GetActorsRoundTrip struct {
	conn net.Conn

	SelectActorConn <-chan SelectActorConn
	Error           <-chan error
}

func (trip GetActorsRoundTrip) run() GetActorsRoundTrip {
	var (
		selectActor chan<- SelectActorConn
		hadError    chan<- error
	)

	closeChans := func() func() {
		var (
			selectActorCh = make(chan SelectActorConn)
			errorCh       = make(chan error)
		)

		trip.SelectActorConn, selectActor =
			selectActorCh, selectActorCh
		trip.Error, hadError =
			errorCh, errorCh

		return func() {
			close(errorCh)
		}
	}()

	go func() {
		defer closeChans()

		eType, err := trip.conn.NextType()
		if err != nil {
			hadError <- err
			return
		}

		switch eType {
		case net.ET_ACTORS:
			var actors net.ActorsList
			err := trip.conn.Decode(&actors)
			if err != nil {
				hadError <- err
				return
			}

			selectActor <- selectActorConn{
				conn:   trip.conn,
				actors: actors,
			}

		default:
			hadError <- fmt.Errorf("expected %v got: %v", net.ET_ACTORS, eType)
		}
	}()

	return trip
}

// SelectActorConn is used to select or create an actor
// for the user to use during this session. If the name
// passed to SelectActor doesn't exist it will be created.
type SelectActorConn interface {
	Actors() []string

	// Will select the actor associated with this connection.
	// If the selected actor name doesn't exist, it will be created.
	SelectActor(name string) SelectActorRoundTrip
}

type selectActorConn struct {
	conn   net.Conn
	actors []string
}

func (c selectActorConn) Actors() []string { return c.actors }

func (c selectActorConn) SelectActor(name string) SelectActorRoundTrip {
	return SelectActorRoundTrip{conn: c.conn}.run(name)
}

// SelectActorRoundTrip is a Request->Response transaction to select the
// Actor associated with the User's connection.
type SelectActorRoundTrip struct {
	conn net.Conn

	CreatedActor  <-chan SelectedActorConn
	SelectedActor <-chan SelectedActorConn
	Error         <-chan error
}

func (trip SelectActorRoundTrip) run(actorName string) SelectActorRoundTrip {
	var (
		createdActor  chan<- SelectedActorConn
		selectedActor chan<- SelectedActorConn
		hadError      chan<- error
	)

	closeChans := func() func() {
		var (
			createdActorCh  = make(chan SelectedActorConn)
			selectedActorCh = make(chan SelectedActorConn)
			errorCh         = make(chan error)
		)

		trip.CreatedActor, createdActor =
			createdActorCh, createdActorCh
		trip.SelectedActor, selectedActor =
			selectedActorCh, selectedActorCh
		trip.Error, hadError =
			errorCh, errorCh

		return func() {
			close(createdActorCh)
			close(selectedActorCh)
			close(errorCh)
		}
	}()

	go func() {
		defer closeChans()

		err := trip.conn.Encode(net.SelectActorRequest{
			Name: actorName,
		})
		if err != nil {
			hadError <- err
			return
		}

		eType, err := trip.conn.NextType()
		if err != nil {
			hadError <- err
			return
		}

		switch eType {
		case net.ET_CREATE_ACTOR_SUCCESS:
			var r net.CreateActorSuccess
			err = trip.conn.Decode(&r)
			if err != nil {
				hadError <- err
				return
			}

			createdActor <- selectedActorConn{
				conn:  trip.conn,
				actor: r.Actor,
			}

		case net.ET_SELECT_ACTOR_SUCCESS:
			var r net.SelectActorSuccess
			err = trip.conn.Decode(&r)
			if err != nil {
				hadError <- err
				return
			}

			selectedActor <- selectedActorConn{
				conn:  trip.conn,
				actor: r.Actor,
			}

		default:
			hadError <- fmt.Errorf("expected %v or %v got: %v", net.ET_CREATE_ACTOR_SUCCESS, net.ET_SELECT_ACTOR_SUCCESS, eType)
		}
	}()

	return trip
}

type SelectedActorConn interface {
	Actor() filu.Actor
}

type selectedActorConn struct {
	conn  net.Conn
	actor filu.Actor
}

func (c selectedActorConn) Actor() filu.Actor { return c.actor }
