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

func DescribeClientServerProtocol(c gospec.Context) {
	authDB := auth.NewStream(nil, nil, nil)

	createUser := func(conn mockConn, username, password string) (net.AuthenticatedUser, client.CreatedUser) {
		trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin(username, password)
		user, err := net.AuthenticateFrom(conn.server, authDB)
		c.Assume(err, IsNil)
		select {
		case err := <-trip.Error:
			panic(err)
		case resp := <-trip.LoginFailure:
			panic(resp)
		case resp := <-trip.LoginSuccess:
			panic(resp)

		case resp := <-trip.CreateSuccess:
			return user, resp
		}
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

				authFailure := <-trip.LoginFailure
				c.Assume(<-trip.Error, IsNil)
				c.Expect(authFailure.Name, Equals, "newUser")
			})
		})

		c.Specify("can log a user in", func() {
			createUser(conn, "username", "password")
			trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin("username", "password")
			authedUser, err := net.AuthenticateFrom(conn.server, authDB)
			c.Assume(err, IsNil)

			loggedInUser := <-trip.LoginSuccess
			c.Assume(<-trip.Error, IsNil)
			c.Expect(authedUser.Username, Equals, loggedInUser.Name)

			c.Specify("unless the password is invalid", func() {
				trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin("username", "invalid")
				_, err := net.AuthenticateFrom(conn.server, authDB)
				c.Expect(err, Equals, net.ErrInvalidLoginCredentials)

				loginFailure := <-trip.LoginFailure
				c.Assume(<-trip.Error, IsNil)
				c.Expect(loginFailure.Name, Equals, "username")
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
			c.Expect((<-trip.SelectActorConn).Actors(), ContainsAll, []string{
				"jim the slayer",
				"jim the destroyer",
				"jimmy shrimp steamer",
			})
			c.Assume(<-trip.Error, IsNil)
		})

		trip := createdUser.GetActors()
		c.Assume(net.SendActors(conn.server, actorDB.Get, authenticatedUser), IsNil)
		selectActorConn := <-trip.SelectActorConn

		c.Specify("can create a new actor", func() {
			trip := selectActorConn.SelectActor("jay")

			actor, err := net.SelectActorFrom(conn.server, actorDB.Select, authenticatedUser)
			c.Assume(err, IsNil)

			selectedActor := <-trip.CreatedActor
			c.Assume(selectedActor, Not(IsNil))
			c.Assume(<-trip.Error, IsNil)

			expectedActor := filu.Actor{
				Username: "jim",
				Name:     "jay",
			}
			c.Expect(actor, Equals, expectedActor)
			c.Expect(selectedActor.Actor(), Equals, expectedActor)
		})

		c.Specify("can select an actor", func() {
			trip := selectActorConn.SelectActor("jim the slayer")

			actor, err := net.SelectActorFrom(conn.server, actorDB.Select, authenticatedUser)
			c.Assume(err, IsNil)

			selectedActor := <-trip.SelectedActor
			c.Assume(<-trip.Error, IsNil)

			expectedActor := filu.Actor{
				Username: "jim",
				Name:     "jim the slayer",
			}
			c.Expect(actor, Equals, expectedActor)
			c.Expect(selectedActor.Actor(), Equals, expectedActor)
		})
	})
}
