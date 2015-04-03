package ssim

// A SelectableEventWriter is an EventWriter that
// is easier to write to from a select statement.
type SelectableEventWriter interface {
	Write() chan<- Event
}

// A SelectableEventEmitter is an EventEmitter that
// is easier to subscribe to from a select statement.
type SelectableEventEmitter interface {
	Subscribe() chan<- EventWriter
}

// A SelectableEventStream is an EventStream that
// is easier to use from a select statement.
type SelectableEventStream interface {
	SelectableEventWriter
	SelectableEventEmitter
	HaltStream() EventStream
}

type eventStreamHaltReq struct {
	hasHalted chan<- EventStream
}

type syncedEventStream struct {
	in   chan<- Event
	subs chan<- EventWriter

	halt chan<- eventStreamHaltReq
}

func (s syncedEventStream) Write() chan<- Event {
	return s.in
}

func (s syncedEventStream) Subscribe() chan<- EventWriter {
	return s.subs
}

func (s syncedEventStream) HaltStream() EventStream {
	stream := make(chan EventStream)
	s.halt <- eventStreamHaltReq{stream}
	return <-stream
}

// NewSyncedEventStream starts a go routine that synchronizes
// access to an EventStream. It returns an interface that is
// friendly to usage from a select statement.
func NewSyncedEventStream(stream EventStream) SelectableEventStream {
	var (
		sync syncedEventStream

		in   <-chan Event
		subs <-chan EventWriter
		halt <-chan eventStreamHaltReq
	)

	closeChans := func() func() {
		var (
			inCh   = make(chan Event)
			subsCh = make(chan EventWriter)
			haltCh = make(chan eventStreamHaltReq)
		)

		sync.in, in = inCh, inCh
		sync.subs, subs = subsCh, subsCh
		sync.halt, halt = haltCh, haltCh

		return func() {
			close(inCh)
			close(subsCh)
			close(haltCh)
		}
	}()

	go func() {
		var haltReq eventStreamHaltReq

	communication:
		for {
			select {
			case e := <-in:
				stream.Write(e)

			case w := <-subs:
				stream.Subscribe(w)

			case haltReq = <-halt:
				break communication
			}
		}

		closeChans()
		haltReq.hasHalted <- stream
	}()

	return sync
}
