package filu

import "time"

// An Event is an immutable fact that happened at a single moment in time.
type Event interface {
	HappenedAt() time.Time
}

// A Time is a single moment in time. It is used by Implementors
// of the Event interface to avoid reimplementation of the HappenedAt()
// method.
type Time struct {
	moment time.Time
}

func Now() Time {
	return Time{time.Now()}
}

func (e Time) HappenedAt() time.Time {
	return e.moment
}
