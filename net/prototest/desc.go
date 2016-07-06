package prototest

import (
	"io"

	"github.com/ghthor/filu"
	"github.com/ghthor/filu/actor"
	"github.com/ghthor/filu/auth"
	"github.com/ghthor/filu/net"
	"github.com/ghthor/filu/net/client"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type mockConn struct {
	pr [2]*io.PipeReader
	pw [2]*io.PipeWriter

	server, client net.Conn
}

type readWriter struct {
	io.Reader
	io.Writer
}

func newMockConn() mockConn {
	c := mockConn{}
	c.pr[0], c.pw[0] = io.Pipe()
	c.pr[1], c.pw[1] = io.Pipe()

	c.server = net.NewGobConn(readWriter{c.pr[0], c.pw[1]})
	c.client = net.NewGobConn(readWriter{c.pr[1], c.pw[0]})

	return c
}

func (c mockConn) close() {
	for _, r := range c.pr {
		r.Close()
	}

	for _, w := range c.pw {
		w.Close()
	}
}

type mockActorDB struct {
	Get    chan<- actor.GetActorsRequest
	Select chan<- actor.SelectionRequest
}

func (db mockActorDB) close() {
	close(db.Get)
	close(db.Select)
}

func newMockActorDB(actors map[string][]string) *mockActorDB {
	getCh := make(chan actor.GetActorsRequest)
	selCh := make(chan actor.SelectionRequest)
	db := &mockActorDB{
		Get:    getCh,
		Select: selCh,
	}

	getProc := actor.NewGetActorsRequestProcessor()

	actor.SelectionRequestSource(selCh).
		WriteToProcessor(actor.NewSelectionProcessor()).
		WriteTo(getProc).
		End()

	actor.GetActorsRequestSource(getCh).
		WriteToProcessor(getProc).
		End()

	for username, names := range actors {
		for _, name := range names {
			db.createActor(username, name)
		}
	}

	return db
}

func (db mockActorDB) createActor(username, actorname string) actor.CreatedActor {
	r := actor.NewSelectionRequest(filu.Actor{
		Username: username,
		Name:     actorname,
	})
	db.Select <- r
	return <-r.CreatedActor
}

type loginTripResult struct {
	err          error
	failure      net.UserLoginFailure
	loggedInUser client.LoggedInUser
	createdUser  client.CreatedUser
}

func NewLoginResult(trip client.LoginRoundTrip) loginTripResult {
	var result loginTripResult

	select {
	case result.err = <-trip.Error:
	case result.failure = <-trip.LoginFailure:
	case result.loggedInUser = <-trip.LoginSuccess:
	case result.createdUser = <-trip.CreateSuccess:
	}

	return result
}

func DescribeClientServerProtocol(c gospec.Context) {
	authDB := auth.NewStream(nil, nil, nil)

	createUser := func(conn mockConn, username, password string) (net.AuthenticatedUser, client.CreatedUser) {
		trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin(username, password)
		user, err := net.AuthenticateFrom(conn.server, authDB)
		c.Assume(err, IsNil)

		result := NewLoginResult(trip)

		c.Assume(result.err, IsNil)
		c.Assume(result.failure, Equals, net.UserLoginFailure{})
		c.Assume(result.loggedInUser, Equals, client.LoggedInUser{})

		return user, result.createdUser
	}

	conn := newMockConn()
	defer conn.close()

	c.Specify("an unauthenticated connection", func() {
		c.Specify("can create a new user", func() {
			authUser, createdUser := createUser(conn, "newUser", "password")
			c.Expect(authUser.Username, Equals, createdUser.Name)

			c.Specify("unless the user already exists", func() {
				trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin("newUser", "some other password")
				_, err := net.AuthenticateFrom(conn.server, authDB)
				c.Expect(err, Equals, net.ErrInvalidLoginCredentials)

				result := NewLoginResult(trip)
				c.Assume(result.err, IsNil)
				c.Assume(result.loggedInUser, Equals, client.LoggedInUser{})
				c.Assume(result.createdUser, Equals, client.CreatedUser{})
				c.Expect(result.failure.Name, Equals, "newUser")
			})
		})

		c.Specify("can log a user in", func() {
			createUser(conn, "username", "password")
			trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin("username", "password")
			authedUser, err := net.AuthenticateFrom(conn.server, authDB)
			c.Assume(err, IsNil)

			result := NewLoginResult(trip)
			c.Assume(result.err, IsNil)
			c.Assume(result.failure, Equals, net.UserLoginFailure{})
			c.Assume(result.createdUser, Equals, client.CreatedUser{})
			c.Expect(result.loggedInUser.Name, Equals, authedUser.Username)

			c.Specify("unless the password is invalid", func() {
				trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin("username", "invalid")
				_, err := net.AuthenticateFrom(conn.server, authDB)
				c.Expect(err, Equals, net.ErrInvalidLoginCredentials)

				result := NewLoginResult(trip)
				c.Assume(result.err, IsNil)
				c.Assume(result.loggedInUser, Equals, client.LoggedInUser{})
				c.Assume(result.createdUser, Equals, client.CreatedUser{})
				c.Expect(result.failure.Name, Equals, "username")
			})
		})
	})

	authenticatedUser, createdUser := createUser(conn, "jim", "jimisthebest11!")

	actorDB := newMockActorDB(map[string][]string{
		"jim": {
			"jim the slayer",
			"jim the destroyer",
			"jimmy shrimp steamer",
		},
	})
	defer actorDB.close()

	c.Specify("an authenticated connection", func() {
		c.Specify("receives a list of actors", func() {
			trip := createdUser.GetActors()
			c.Expect(net.SendActors(conn.server, actorDB.Get, authenticatedUser), IsNil)

			var err error = nil
			var selectConn client.SelectActorConn
			select {
			case err = <-trip.Error:
			case selectConn = <-trip.SelectActorConn:
			}

			c.Assume(err, IsNil)
			c.Expect(selectConn.Actors(), ContainsAll, []string{
				"jim the slayer",
				"jim the destroyer",
				"jimmy shrimp steamer",
			})
		})

		trip := createdUser.GetActors()
		c.Assume(net.SendActors(conn.server, actorDB.Get, authenticatedUser), IsNil)

		var err error = nil
		var selectConn client.SelectActorConn
		select {
		case err = <-trip.Error:
		case selectConn = <-trip.SelectActorConn:
		}
		c.Assume(err, IsNil)

		c.Specify("can create a new actor", func() {
			trip := selectConn.SelectActor("jay")

			actor, err := net.SelectActorFrom(conn.server, actorDB.Select, authenticatedUser)
			c.Assume(err, IsNil)

			var selectedActor client.SelectedActorConn
			var createdActor client.SelectedActorConn
			select {
			case err = <-trip.Error:
			case selectedActor = <-trip.SelectedActor:
			case createdActor = <-trip.CreatedActor:
			}
			c.Assume(err, IsNil)
			c.Assume(selectedActor, IsNil)
			c.Assume(createdActor, Not(IsNil))

			expectedActor := filu.Actor{
				Username: "jim",
				Name:     "jay",
			}
			c.Expect(actor, Equals, expectedActor)
			c.Expect(createdActor.Actor(), Equals, expectedActor)
		})

		c.Specify("can select an actor", func() {
			trip := selectConn.SelectActor("jim the slayer")

			actor, err := net.SelectActorFrom(conn.server, actorDB.Select, authenticatedUser)
			c.Assume(err, IsNil)

			var selectedActor client.SelectedActorConn
			var createdActor client.SelectedActorConn
			select {
			case err = <-trip.Error:
			case selectedActor = <-trip.SelectedActor:
			case createdActor = <-trip.CreatedActor:
			}
			c.Assume(err, IsNil)
			c.Assume(selectedActor, Not(IsNil))
			c.Assume(createdActor, IsNil)

			expectedActor := filu.Actor{
				Username: "jim",
				Name:     "jim the slayer",
			}
			c.Expect(actor, Equals, expectedActor)
			c.Expect(selectedActor.Actor(), Equals, expectedActor)
		})
	})
}
