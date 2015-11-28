// Package auth provides a stream processing pattern
// to supply user authentication to a filu application
package auth

import "github.com/ghthor/filu"

// A Request is a filu.Event that represents an
// authentication request sent by a client/user.
// It is consumed by a Processor that will output
// a PostAuthEvent.
type Request struct {
	filu.Time
	Username, Password string

	// The public interface for the user to receive the
	// result of the authentication request.
	InvalidPassword   <-chan InvalidPassword
	CreatedUser       <-chan CreatedUser
	AuthenticatedUser <-chan AuthenticatedUser

	// The private interface used by the stream terminator
	// to respond with the result of the Request.
	sendInvalidPassword   chan<- InvalidPassword
	sendCreatedUser       chan<- CreatedUser
	sendAuthenticatedUser chan<- AuthenticatedUser
}

// NewRequest will construct a Request suitible for use with
// Stream.RequestAuthentication() <- Request.
func NewRequest(username, password string) Request {
	invalidCh := make(chan InvalidPassword)
	createdCh := make(chan CreatedUser)
	authenticatedCh := make(chan AuthenticatedUser)

	return Request{
		Time:     filu.Now(),
		Username: username,
		Password: password,

		InvalidPassword:   invalidCh,
		CreatedUser:       createdCh,
		AuthenticatedUser: authenticatedCh,

		sendInvalidPassword:   invalidCh,
		sendCreatedUser:       createdCh,
		sendAuthenticatedUser: authenticatedCh,
	}
}

// A RequestConsumer is used as the consumption end of a RequestStream.
type RequestConsumer interface {
	// The implementation of Write can assume in will never be called in parallel.
	Write(Request)
}

// A RequestProducer is used as the production end of a RequestStream.
type RequestProducer interface {
	Read() <-chan Request
}

// A RequestStream represents a function that when given a Request
// will produce a Request.
type RequestStream interface {
	RequestConsumer
	RequestProducer
}

func linkRequest(source RequestProducer, destination RequestConsumer) {
	go func() {
		for r := range source.Read() {
			destination.Write(r)
		}
	}()
}

// NewRequestStream will concatenate a series of RequestStreams into
// a single RequestStream. The Consumer entry point will be the first parameter,
// the Producer endpoint will be the last parameter.
func NewRequestStream(streams ...RequestStream) RequestStream {
	switch len(streams) {
	case 0:
		return nil
	case 1:
		return streams[0]
	default:
	}

	linkRequest(streams[0], streams[1])

	return struct {
		RequestConsumer
		RequestProducer
	}{
		streams[0],
		NewRequestStream(streams[1:]...),
	}
}

// A Result of a Request after it was processed by a Processor.
type Result interface {
	filu.Event

	respondToRequestor()
}

// A ResultProducer is the source of a stream of Results.
type ResultProducer interface {
	Read() <-chan Result
}

// A ResultConsumer is a sink of a stream of Results.
type ResultConsumer interface {
	// The implementation of Write can assume in will never be called in parallel.
	Write(Result)
}

// A ResultStream is sink & source of Results. It is implemented
// and used when constructing a Stream to hook into the post-auth
// Result stream for user defined processing.
type ResultStream interface {
	ResultProducer
	ResultConsumer
}

func linkResult(source ResultProducer, destination ResultConsumer) {
	if source == nil || destination == nil {
		return
	}

	go func() {
		for r := range source.Read() {
			destination.Write(r)
		}
	}()
}

// NewResultStream will concatenate a series of ResultStreams into
// a single ResultStream. The Consumer entry point will be the first parameter,
// the Producer endpoint will be the last parameter.
func NewResultStream(streams ...ResultStream) ResultStream {
	switch len(streams) {
	case 0:
		return nil
	case 1:
		return streams[0]
	default:
	}

	linkResult(streams[0], streams[1])

	return struct {
		ResultConsumer
		ResultProducer
	}{
		streams[0],
		NewResultStream(streams[1:]...),
	}
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

// An AuthenticatedUser is the result of a correct Username & Password combonation.
type AuthenticatedUser struct {
	filu.Time
	Request
}

// A Processor is the step in a Stream when a Request is transformed into a Result.
// This is where a Username/Password pair would be compared against what exists in
// a database to determine if the pair is a valid.
type Processor interface {
	RequestConsumer
	ResultProducer
}

// A Stream consumes Request's.
type Stream interface {
	RequestAuthentication() chan<- Request
}

// A memoryProcessor stores all registered Username/Password
// combonations in a go map. The map is a materialized view of
// the Request stream.
type memoryProcessor struct {
	users   map[string]string
	results chan Result
}

func newMemoryProcessor() memoryProcessor {
	return memoryProcessor{
		users:   make(map[string]string),
		results: make(chan Result),
	}
}

func (p memoryProcessor) Write(r Request) {
	password := p.users[r.Username]
	switch {
	case password == "":
		p.users[r.Username] = r.Password
		p.results <- CreatedUser{
			Time:    filu.Now(),
			Request: r,
		}

	case password == r.Password:
		p.results <- AuthenticatedUser{
			Time:    filu.Now(),
			Request: r,
		}

	default:
		p.results <- InvalidPassword{
			Time:    filu.Now(),
			Request: r,
		}
	}
}

func (p memoryProcessor) Read() <-chan Result {
	return p.results
}

type streamHead struct {
	requests chan<- Request
}

func (s streamHead) RequestAuthentication() chan<- Request {
	return s.requests
}

func newStreamHead(consumer RequestConsumer) Stream {
	var requests <-chan Request
	var head streamHead

	func() {
		requestsCh := make(chan Request)

		requests = requestsCh

		head = streamHead{
			requests: requestsCh,
		}
	}()

	go func() {
		for r := range requests {
			consumer.Write(r)
		}
	}()

	return head
}

// NewStream creates an auth processor and connect the Result output
// into the provided ResultStream's and returns a terminated Stream
// that will return the Result of a Request back to the Requestor.
func NewStream(preAuth RequestStream, postAuth ResultStream) Stream {
	proc := newMemoryProcessor()

	if preAuth != nil {
		linkRequest(preAuth, proc)
	}

	if postAuth != nil {
		linkResult(proc, postAuth)
		linkResult(postAuth, terminator{})
	} else {
		linkResult(proc, terminator{})
	}

	if preAuth != nil {
		return newStreamHead(preAuth)
	}

	return newStreamHead(proc)
}

// A terminator comsumes Result's and will terminate an auth Stream.
// The Stream is terminated by sending the Result to the Request sender.
// A terminator has no outputs.
type terminator struct{}

func (terminator) Write(r Result) {
	r.respondToRequestor()
}

func (e InvalidPassword) respondToRequestor() {
	e.Request.sendInvalidPassword <- e
}

func (e CreatedUser) respondToRequestor() {
	e.Request.sendCreatedUser <- e
}

func (e AuthenticatedUser) respondToRequestor() {
	e.Request.sendAuthenticatedUser <- e
}
