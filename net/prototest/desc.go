package prototest

import (
	"io"

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

func DescribeClientServerProtocol(c gospec.Context) {
	authStream := auth.NewStream(nil, nil, nil)

	createUser := func(conn mockConn, username, password string) (net.AuthenticatedUser, client.CreatedUser) {
		trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin(username, password)
		user, err := net.AuthenticateFrom(conn.server, authStream)
		c.Assume(err, IsNil)
		c.Assume(<-trip.Error, IsNil)
		return user, <-trip.CreateSuccess
	}

	c.Specify("a connection", func() {
		conn := newMockConn()
		defer conn.close()

		c.Specify("can create a new user", func() {
			authUser, createdUser := createUser(conn, "newUser", "password")
			c.Expect(authUser.Username, Equals, createdUser.Name)

			c.Specify("unless the user already exists", func() {
				trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin("newUser", "some other password")
				_, err := net.AuthenticateFrom(conn.server, authStream)
				c.Expect(err, Equals, net.ErrInvalidLoginCredentials)

				authFailure := <-trip.LoginFailure
				c.Assume(<-trip.Error, IsNil)
				c.Expect(authFailure.Name, Equals, "newUser")
			})
		})

		c.Specify("can log a user in", func() {
			createUser(conn, "username", "password")

			trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin("username", "password")
			authedUser, err := net.AuthenticateFrom(conn.server, authStream)
			c.Assume(err, IsNil)

			loggedInUser := <-trip.LoginSuccess
			c.Assume(<-trip.Error, IsNil)
			c.Expect(authedUser.Username, Equals, loggedInUser.Name)

			c.Specify("unless the password is invalid", func() {
				trip := client.NewUnauthenticatedConn(conn.client).AttemptLogin("username", "invalid")
				_, err := net.AuthenticateFrom(conn.server, authStream)
				c.Expect(err, Equals, net.ErrInvalidLoginCredentials)

				loginFailure := <-trip.LoginFailure
				c.Assume(<-trip.Error, IsNil)
				c.Expect(loginFailure.Name, Equals, "username")
			})
		})
	})
}
