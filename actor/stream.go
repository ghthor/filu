// Package actor provides a stream processing pattern
// to supply actor selection/creation to a filu application.
package actor

import "github.com/ghthor/filu"

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

func (r SelectionRequest) closeResultChannels() {
	close(r.sendSelectedActor)
	close(r.sendCreatedActor)
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

// A terminator comsumes Result's and will terminate an auth Stream.
// The Stream is terminated by sending the Result to the Request sender.
// A terminator has no outputs.
type terminator struct{}

func (terminator) Write(r SelectionResult) {
	r.respondToRequestor()
}

func (e CreatedActor) respondToRequestor() {
	e.SelectionRequest.sendCreatedActor <- e
	e.SelectionRequest.closeResultChannels()
}

func (e SelectedActor) respondToRequestor() {
	e.SelectionRequest.sendSelectedActor <- e
	e.SelectionRequest.closeResultChannels()
}
