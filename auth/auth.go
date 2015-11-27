package auth

import "github.com/ghthor/filu"

// A Request is a filu.Event that represents an
// authorization request sent by a client/user.
// It is consumed by a Processor that will output
// a PostAuthEvent.
type Request struct {
	filu.Time
	Username, Password string

	// The public interface for the user to receive the
	// result of the authorization request.
	InvalidPassword <-chan InvalidPassword
	CreatedUser     <-chan CreatedUser
	AuthorizedUser  <-chan AuthorizedUser

	// The private interface used by the stream terminator
	// to respond with the result of the Request.
	sendInvalidPassword chan<- InvalidPassword
	sendCreatedUser     chan<- CreatedUser
	sendAuthorizedUser  chan<- AuthorizedUser
}

func NewRequest(username, password string) Request {
	invalidCh := make(chan InvalidPassword)
	createdCh := make(chan CreatedUser)
	authorizedCh := make(chan AuthorizedUser)

	return Request{
		Username: username,
		Password: password,

		InvalidPassword: invalidCh,
		CreatedUser:     createdCh,
		AuthorizedUser:  authorizedCh,

		sendInvalidPassword: invalidCh,
		sendCreatedUser:     createdCh,
		sendAuthorizedUser:  authorizedCh,
	}
}

// A Result of a Request after it was processed by a Processor.
type Result interface {
	filu.Event

	respondToRequestor()
}

// An InvalidPassword is the result of a Request with an invalid password.
type InvalidPassword struct {
	filu.Time
	Request
}

// A CreatedUser is the result of a Request where the user doesn't already exist.
type CreatedUser struct {
	filu.Time
	Request
}

// An AuthorizedUser is the result of a correct Username & Password combonation.
type AuthorizedUser struct {
	filu.Time
	Request
}

// A Stream consumes Request's.
type Stream interface {
	RequestAuthorization() chan<- Request
}

// A memoryProcessor stores all registered Username/Password
// combonations in a go map. The map is a materialized view of
// the Request stream.
type memoryProcessor struct {
	requests chan<- Request
}

func newMemoryProcessor(results chan<- Result) memoryProcessor {
	var in <-chan Request

	var memProc memoryProcessor

	func() {
		requestCh := make(chan Request)

		in = requestCh

		memProc = memoryProcessor{
			requests: requestCh,
		}
	}()

	go func() {
		users := make(map[string]string)

		for r := range in {
			password := users[r.Username]
			switch {
			case password == "":
				users[r.Username] = r.Password
				results <- CreatedUser{Request: r}

			case password == r.Password:
				results <- AuthorizedUser{Request: r}

			default:
				results <- InvalidPassword{Request: r}
			}
		}
	}()

	return memProc
}

func (p memoryProcessor) RequestAuthorization() chan<- Request {
	return p.requests
}

// A NewStream creates a authorization that is terminated.
func NewStream() Stream {
	return newMemoryProcessor(newTerminator())
}

// A terminator comsumes Result's and will terminate an auth Stream.
// The Stream is terminated by sending the Result to the Request sender.
// A terminator has no outputs.
type terminator chan<- Result

func newTerminator() terminator {
	resultCh := make(chan Result)

	go func() {
		for r := range resultCh {
			r.respondToRequestor()
		}
	}()

	return resultCh
}

func (e InvalidPassword) respondToRequestor() {
	e.Request.sendInvalidPassword <- e
}

func (e CreatedUser) respondToRequestor() {
	e.Request.sendCreatedUser <- e
}

func (e AuthorizedUser) respondToRequestor() {
	e.Request.sendAuthorizedUser <- e
}
