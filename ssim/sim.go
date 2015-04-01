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

// NewEventPipeline will chain `n` EventStreams together with streams[0]
// as the entry point and streams[len(streams)-1] being the exit point.
func NewEventPipeline(streams ...EventStream) EventStream {
	switch len(streams) {
	case 0:
		return nil
	case 1:
		return streams[0]
	default:
	}

	streams[0].Subscribe(streams[1])

	return struct {
		EventWriter
		EventEmitter
	}{
		streams[0],
		NewEventPipeline(streams[1:]...),
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

// NewChangePipeline will chain `n` ChangeStreams together with streams[0]
// as the entry point and streams[len(streams)-1] being the exit point.
func NewChangePipeline(streams ...ChangeStream) ChangeStream {
	switch len(streams) {
	case 0:
		return nil
	case 1:
		return streams[0]
	default:
	}

	streams[0].Subscribe(streams[1])

	return struct {
		ChangeWriter
		ChangeEmitter
	}{
		streams[0],
		NewChangePipeline(streams[1:]...),
	}
}
