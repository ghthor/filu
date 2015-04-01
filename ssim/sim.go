// Package ssim is an experimental implementation
// of an append only log of immutable Event's.
package ssim

import "time"

// A ActorID is a unique ID assigned to an actor.
type ActorID int

// An Event is an immutable fact that an actor
// emits to interact with a simulation.
type Event interface {
	// The actor that produced the event.
	Source() ActorID

	// The time(according to the client)
	// when the event was issued.
	IssuedAt() time.Time
}

// An EventEmitter will emit Event's to all
// subscribed EventWriter's.
type EventEmitter interface {
	Subscribe(EventWriter)
}

// An EventWriter can receive Event's.
type EventWriter interface {
	Write(Event)
}

// An EventStream can receive Event's and will
// emit them to all subscriber's.
type EventStream interface {
	EventWriter
	EventEmitter
}

func NewEventPipeline(nodes ...EventStream) EventStream {
	switch len(nodes) {
	case 0:
		return nil
	case 1:
		return nodes[0]
	default:
	}

	nodes[0].Subscribe(nodes[1])

	return struct {
		EventWriter
		EventEmitter
	}{
		nodes[0],
		NewEventPipeline(nodes[1:]...),
	}

}

// A Change is the immutable modification
// to the state of the simulation caused by an Event.
type Change interface {
	// The Event that was the cause of this change.
	Source() Event
}

// A ChangeEmitter will emit Change's to all
// subscribed ChangeWriter's.
type ChangeEmitter interface {
	Subscribe(ChangeWriter)
}

// A ChangeWriter can receive Change's.
type ChangeWriter interface {
	Write(Change)
}

// A ChangeStream can receive Change's and will
// emit them to all subscriber's.
type ChangeStream interface {
	ChangeWriter
	ChangeEmitter
}

func NewChangePipeline(nodes ...ChangeStream) ChangeStream {
	switch len(nodes) {
	case 0:
		return nil
	case 1:
		return nodes[0]
	default:
	}

	nodes[0].Subscribe(nodes[1])

	return struct {
		ChangeWriter
		ChangeEmitter
	}{
		nodes[0],
		NewChangePipeline(nodes[1:]...),
	}
}
