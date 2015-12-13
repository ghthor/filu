// Package actor provides a stream processing pattern
// to supply actor selection/creation to a filu application.
package actor

import (
	"sync"

	"github.com/ghthor/filu"
)

// A SelectionRequest is a filu.Event that represents an
// Actor selection request sent by an authenticated User.
// It is consumed by a Processor that will output a
// SelectionResult.
type SelectionRequest struct {
	filu.Time
	filu.Actor

	// The public interface for the user to receive the
	// result of their selection request.
	SelectedActor <-chan SelectedActor
	CreatedActor  <-chan CreatedActor

	// The private interface used by the stream terminator
	// to respond with the result of the Request.
	sendSelectedActor chan<- SelectedActor
	sendCreatedActor  chan<- CreatedActor
}

func NewSelectionRequest(actor filu.Actor) SelectionRequest {
	selectedActorCh := make(chan SelectedActor, 1)
	createdActorCh := make(chan CreatedActor, 1)
	return SelectionRequest{
		Time:  filu.Now(),
		Actor: actor,

		SelectedActor: selectedActorCh,
		CreatedActor:  createdActorCh,

		sendSelectedActor: selectedActorCh,
		sendCreatedActor:  createdActorCh,
	}
}

// A SelectionResult is a filu.Event that contains the Actor
// that was the target of a selection request.
type SelectionResult interface {
	filu.Event

	Source() SelectionRequest

	respondToRequestor()
}

type SelectionRequestSource <-chan SelectionRequest

func (s SelectionRequestSource) WriteTo(sink SelectionRequestSink) SelectionRequestSource {
	return sink.ReadSelectionRequestsFrom(s)
}

func (s SelectionRequestSource) WriteToProcessor(proc SelectionProcessor) SelectionResultSource {
	return proc.ProcessSelectionRequestsFrom(s)
}

type SelectionRequestSink interface {
	ReadSelectionRequestsFrom(<-chan SelectionRequest) SelectionRequestSource
}

type SelectionResultSource <-chan SelectionResult

func (s SelectionResultSource) WriteTo(sink SelectionResultSink) SelectionResultSource {
	return sink.ReadSelectionResultsFrom(s)
}

func (s SelectionResultSource) End() {
	go func() {
		for r := range s {
			r.respondToRequestor()
		}
	}()
}

type SelectionResultSink interface {
	ReadSelectionResultsFrom(<-chan SelectionResult) SelectionResultSource
}

type SelectionProcessor interface {
	ProcessSelectionRequestsFrom(<-chan SelectionRequest) SelectionResultSource
}

// Dictionary of [Username:Actor]
type actors []filu.Actor

func (actors actors) contains(target filu.Actor) bool {
	for _, a := range actors {
		if a == target {
			return true
		}
	}

	return false
}

type actorsHashmapDB map[string]actors

func (db actorsHashmapDB) ProcessSelectionRequestsFrom(requests <-chan SelectionRequest) SelectionResultSource {
	out := make(chan SelectionResult)

	go func(out chan<- SelectionResult) {
		for r := range requests {
			actors := db[r.Username]
			actor := filu.Actor{
				Username: r.Username,
				Name:     r.Name,
			}

			if actors.contains(actor) {
				out <- SelectedActor{
					Time:             filu.Now(),
					Actor:            actor,
					SelectionRequest: r,
				}
				continue
			}

			db[r.Username] = append(actors, actor)

			out <- CreatedActor{
				Time:             filu.Now(),
				Actor:            actor,
				SelectionRequest: r,
			}
		}

		close(out)
	}(out)

	return out
}

func NewSelectionProcessor() SelectionProcessor {
	return make(actorsHashmapDB, 100)
}

// A CreatedActor is the result of a selection request where
// the actor's name didn't exist for the user.
type CreatedActor struct {
	filu.Time
	filu.Actor

	SelectionRequest
}

// A SelectedActor is the result of a selection request where
// the actor's name did exist for the user.
type SelectedActor struct {
	filu.Time
	filu.Actor

	SelectionRequest
}

func (a CreatedActor) Source() SelectionRequest {
	return a.SelectionRequest
}

func (a SelectedActor) Source() SelectionRequest {
	return a.SelectionRequest
}

// A terminator comsumes Result's and will terminate an auth Stream.
// The Stream is terminated by sending the Result to the Request sender.
// A terminator has no outputs.
type terminator struct{}

func (terminator) Write(r SelectionResult) {
	r.respondToRequestor()
}

func (e CreatedActor) respondToRequestor() {
	e.SelectionRequest.sendCreatedActor <- e
}

func (e SelectedActor) respondToRequestor() {
	e.SelectionRequest.sendSelectedActor <- e
}

// A GetActorsRequest is created for a specific Username and will
// return all the actors belonging to that username.
type GetActorsRequest struct {
	Username string

	Actors     <-chan []filu.Actor
	sendActors chan<- []filu.Actor
}

func NewGetActorsRequest(username string) GetActorsRequest {
	resultCh := make(chan []filu.Actor, 1)
	return GetActorsRequest{
		Username:   username,
		Actors:     resultCh,
		sendActors: resultCh,
	}
}

func (r GetActorsRequest) result(actors []filu.Actor) {
	r.sendActors <- actors
}

// GetActorsResult will supply the list of actors for a given Username.
// It is the result of creating and sending a GetActorsRequest.
type ExistingActors struct {
	filu.Time
	Actors []filu.Actor
	GetActorsRequest
}

// A GetActorsRequestSource is created to build a GetActors stream.
// Make a channel that will represent the public API to the stream
// and chain .WriteTo(Sink). Once the stream of processors is constructed
// the stream must be .End()'ed.
type GetActorsRequestSource <-chan GetActorsRequest

// A GetActorsRequestSink should be implemented as a concurrent process.
// If the source channel is closed, it is expected that the process will
// close the channel it writes it's input too, to ensure the pipeline
// will shutdown.
type GetActorsRequestSink interface {
	ReadGetActorsRequestsFrom(<-chan GetActorsRequest) GetActorsRequestSource
}

func (s GetActorsRequestSource) WriteTo(sink GetActorsRequestSink) GetActorsRequestSource {
	return sink.ReadGetActorsRequestsFrom(s)
}

func (s GetActorsRequestSource) WriteToProcessor(sink GetActorsRequestProcessor) ExistingActorsSource {
	return sink.ProcessGetActorRequestsFrom(s)
}

type ExistingActorsSource <-chan ExistingActors

// An ExistingActorsSink should be implemented as a concurrent process.
// If the source channel is closed, it is expected that the process will
// close the channel it writes it's input too, to ensure the pipeline
// will shutdown.
type ExistingActorsSink interface {
	ReadExistingActorsFrom(<-chan ExistingActors) ExistingActorsSource
}

func (s ExistingActorsSource) WriteTo(sink ExistingActorsSink) ExistingActorsSource {
	return sink.ReadExistingActorsFrom(s)
}

func (s ExistingActorsSource) End() {
	go func() {
		for r := range s {
			r.GetActorsRequest.result(r.Actors)
		}
	}()
}

type GetActorsRequestProcessor interface {
	SelectionResultSink
	ProcessGetActorRequestsFrom(<-chan GetActorsRequest) ExistingActorsSource
}

func NewGetActorsRequestProcessor() GetActorsRequestProcessor {
	return &existingActorsDB{
		actors: make(map[string][]filu.Actor, 100),
	}
}

type existingActorsDB struct {
	sync.RWMutex
	actors map[string][]filu.Actor
}

func (db *existingActorsDB) ReadSelectionResultsFrom(results <-chan SelectionResult) SelectionResultSource {
	out := make(chan SelectionResult)

	go func(out chan<- SelectionResult) {
		defer close(out)

		for result := range results {
			db.Lock()

			request := result.Source()
			actors := db.actors[request.Username]
			actors = append(actors, filu.Actor{
				Username: request.Username,
				Name:     request.Name,
			})
			db.actors[request.Username] = actors

			out <- result

			db.Unlock()
		}
	}(out)

	return out
}

func (db *existingActorsDB) ProcessGetActorRequestsFrom(requests <-chan GetActorsRequest) ExistingActorsSource {
	out := make(chan ExistingActors)

	go func(out chan<- ExistingActors) {
		defer close(out)

		for r := range requests {
			db.RLock()

			out <- ExistingActors{
				Time:   filu.Now(),
				Actors: db.actors[r.Username],

				GetActorsRequest: r,
			}

			db.RUnlock()
		}
	}(out)

	return out
}
