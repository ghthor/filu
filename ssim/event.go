package ssim

type eventEmitter struct {
	listeners []EventWriter
}

func (em eventEmitter) emit(e Event) {
	for _, l := range em.listeners {
		l.Write(e)
	}
}

func (em *eventEmitter) Subscribe(w EventWriter) {
	em.listeners = append(em.listeners, w)
}

type eventProcessor struct {
	fn func(Event)
	*eventEmitter
}

func (p eventProcessor) Write(e Event) {
	p.fn(e)
	p.emit(e)
}

func NewEventProccessorFn(fn func(Event)) EventStream {
	return eventProcessor{fn: fn, eventEmitter: &eventEmitter{}}
}
