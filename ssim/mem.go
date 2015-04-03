package ssim

import "time"

type memEventLog struct {
	events []Event
	subs   []EventWriter

	now func() time.Time
}

type MemEventLogOption func(*memEventLog)

func NowProvider(now func() time.Time) MemEventLogOption {
	return func(l *memEventLog) {
		l.now = now
	}
}

func NewMemEventLog(options ...MemEventLogOption) EventStream {
	l := &memEventLog{
		// TODO add a package option to adjust default capacities
		events: make([]Event, 0, 1024),
		subs:   make([]EventWriter, 0, 10),
	}

	for _, o := range options {
		o(l)
	}

	return l
}

func (l *memEventLog) Write(e Event) {
	e = e.AcceptAt(l.now())

	l.events = append(l.events, e)

	// TODO add concurrent writes using a waitgroup
	for _, s := range l.subs {
		s.Write(e)
	}
}

func (l *memEventLog) Subscribe(w EventWriter) {
	l.subs = append(l.subs, w)
}
