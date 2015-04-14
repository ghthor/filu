package ssim

import "time"

type memEventLog struct {
	events []Event
	now    func() time.Time

	*eventEmitter
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
		events:       make([]Event, 0, 1024),
		eventEmitter: &eventEmitter{},
	}

	for _, o := range options {
		o(l)
	}

	return l
}

type LogEvent struct {
	time.Time

	Source Event
}

func (l *memEventLog) Write(e Event) {
	e = LogEvent{
		Time:   l.now(),
		Source: e,
	}

	l.events = append(l.events, e)

	// TODO add concurrent writes using a waitgroup
	for _, s := range l.eventEmitter.listeners {
		s.Write(e)
	}
}
