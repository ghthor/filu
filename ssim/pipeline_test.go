package ssim_test

import (
	"github.com/ghthor/filu/ssim"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type eventEmitter struct {
	listeners []ssim.EventWriter
}

func (em eventEmitter) emit(e ssim.Event) {
	for _, l := range em.listeners {
		l.Write(e)
	}
}

func (em *eventEmitter) Subscribe(w ssim.EventWriter) {
	em.listeners = append(em.listeners, w)
}

type eventProcessor struct {
	fn func(ssim.Event)
	*eventEmitter
}

func newEventProc(fn func(ssim.Event)) eventProcessor {
	return eventProcessor{fn: fn, eventEmitter: &eventEmitter{}}
}

func (p eventProcessor) Write(e ssim.Event) {
	p.fn(e)
	p.emit(e)
}

func DescribePipelines(c gospec.Context) {
	c.Specify("a pipeline", func() {
		c.Specify("can be made from event streams", func() {
			var output string
			pipe := ssim.NewEventPipeline(
				newEventProc(func(e ssim.Event) {
					output += "1"
				}),
				newEventProc(func(e ssim.Event) {
					output += "2"
				}),
				newEventProc(func(e ssim.Event) {
					output += "3"
				}),
			)

			pipe.Write(nil)
			c.Expect(output, Equals, "123")
		})
	})
}
