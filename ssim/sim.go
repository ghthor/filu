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

// An EventProcessor is used to listen and
// respond to events as they are added to an
// EventLog.
type EventProcessor interface {
	Process(Event)
}

// An EventLog is an append only log of Event's.
// Any number of EventProcessor's can be subscribed
// to the log. Each processor will be called
// when an Event is appended to the log.
type EventLog interface {
	Append() chan<- Event
	Subscribe(EventProcessor)
}

// A Change is the immutable modification
// to the state of the simulation caused by an Event.
type Change interface {
	// The Event that was the cause of this change.
	Source() Event
}

// A ChangeProcessor is used to listen and
// respond to events as they are added to a
// ChangeLog.
type ChangeProcessor interface {
	Process(Change)
}

// A ChangeLog is an append only log of Change's.
// Any number of ChangeProcessor's can be subscribed
// to the log. Each processor will be called
// when a Change is appended to the log.
type ChangeLog interface {
	Append(Change)
	Subscribe(ChangeProcessor)
}
