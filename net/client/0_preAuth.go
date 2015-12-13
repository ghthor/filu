package client

import (
	"fmt"

	"github.com/ghthor/filu/net"
)

// An UnauthenticatedConn will require the client to authenticated
// as a registered user. If the requested user doesn't exist it will
// be created.
type UnauthenticatedConn interface {
	// Non-blocking login request
	AttemptLogin(name, password string) LoginRoundTrip
}

// An implementation of UnauthenticatedConn
type preAuthConn struct {
	conn net.Conn
}

// NewUnauthenticatedConn will create a new connection
// that can be used to login or create an user.
func NewUnauthenticatedConn(conn net.Conn) UnauthenticatedConn {
	return &preAuthConn{
		conn: conn,
	}
}

func (c *preAuthConn) AttemptLogin(name, password string) LoginRoundTrip {
	return LoginRoundTrip{conn: c.conn}.run(net.UserLoginRequest{name, password})
}

// A LoginRoundTrip represents a login request -> response transaction.
// All channels should be selected from to receive the response of the
// request.
type LoginRoundTrip struct {
	conn net.Conn

	LoginSuccess  <-chan LoggedInUser
	CreateSuccess <-chan CreatedUser
	LoginFailure  <-chan net.UserLoginFailure
	Error         <-chan error
}

func (trip LoginRoundTrip) run(r net.UserLoginRequest) LoginRoundTrip {
	var (
		loginSuccess  chan<- LoggedInUser
		createSuccess chan<- CreatedUser
		loginFailure  chan<- net.UserLoginFailure
		hadError      chan<- error
	)

	closeChans := func() func() {
		var (
			loginSuccessCh  = make(chan LoggedInUser, 1)
			createSuccessCh = make(chan CreatedUser, 1)
			loginFailureCh  = make(chan net.UserLoginFailure, 1)
			errorCh         = make(chan error, 1)
		)

		trip.LoginSuccess, loginSuccess =
			loginSuccessCh, loginSuccessCh
		trip.CreateSuccess, createSuccess =
			createSuccessCh, createSuccessCh
		trip.LoginFailure, loginFailure =
			loginFailureCh, loginFailureCh
		trip.Error, hadError =
			errorCh, errorCh

		return func() {
			close(loginSuccessCh)
			close(createSuccessCh)
			close(loginFailureCh)
			close(errorCh)
		}
	}()

	go func() {
		defer closeChans()

		err := trip.conn.Encode(r)
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
		case net.ET_USER_CREATE_SUCCESS:
			var r net.UserCreateSuccess
			err := trip.conn.Decode(&r)
			if err != nil {
				hadError <- err
				return
			}

			createSuccess <- CreatedUser{
				Name:           r.Name,
				ActorsListConn: NewActorsListConn(trip.conn),
			}

		case net.ET_USER_LOGIN_FAILED:
			var r net.UserLoginFailure
			err := trip.conn.Decode(&r)
			if err != nil {
				hadError <- err
				return
			}

			loginFailure <- r

		case net.ET_USER_LOGIN_SUCCESS:
			var resp net.UserLoginSuccess
			err := trip.conn.Decode(&resp)
			if err != nil {
				hadError <- err
				return
			}

			loginSuccess <- LoggedInUser{
				Name:           resp.Name,
				ActorsListConn: NewActorsListConn(trip.conn),
			}

		default:
			hadError <- fmt.Errorf("unexpected login request resp type: %v", eType)
		}
	}()

	return trip
}

// A LoggedInUser represents a successful login attempt by the client.
// It contains the Username of the logged in user and a handle to an
// ActorsListConn which allows the client to have more access to
// functionality.
type LoggedInUser struct {
	Name string
	ActorsListConn
}

// A CreatedUser represents a successful login attempt where the
// requested user didn't exist, therefore it was created. It contains
// the Username of the created user and a handle to an ActorsListConn
// which allows the client to have more access to functionality.
type CreatedUser struct {
	Name string
	ActorsListConn
}
