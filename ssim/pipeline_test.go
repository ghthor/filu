package ssim_test

import (
	"testing"

	"github.com/ghthor/engine/ssim"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type eventEmitter struct {
	listeners []ssim.EventWriter
}

type changeEmitter struct {
	listeners []ssim.ChangeWriter
}

func (em eventEmitter) emit(e ssim.Event) {
	for _, l := range em.listeners {
		l.Write(e)
	}
}

func (em changeEmitter) emit(e ssim.Change) {
	for _, l := range em.listeners {
		l.Write(e)
	}
}

func (em *eventEmitter) Subscribe(w ssim.EventWriter) {
	em.listeners = append(em.listeners, w)
}

func (em *changeEmitter) Subscribe(w ssim.ChangeWriter) {
	em.listeners = append(em.listeners, w)
}

type eventProcessor struct {
	fn func(ssim.Event)
	*eventEmitter
}

type changeProcessor struct {
	fn func(ssim.Change)
	*changeEmitter
}

func newEventProc(fn func(ssim.Event)) eventProcessor {
	return eventProcessor{fn: fn, eventEmitter: &eventEmitter{}}
}

func newChangeProc(fn func(ssim.Change)) changeProcessor {
	return changeProcessor{fn: fn, changeEmitter: &changeEmitter{}}
}

func (p eventProcessor) Write(e ssim.Event) {
	p.fn(e)
	p.emit(e)
}

func (p changeProcessor) Write(e ssim.Change) {
	p.fn(e)
	p.emit(e)
}

func DescribePipelines(c gospec.Context) {
	c.Specify("a pipeline", func() {
		c.Specify("can be made from stream processors", func() {
			c.Specify("for events", func() {
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

			c.Specify("for changes", func() {
				var output string
				pipe := ssim.NewChangePipeline(
					newChangeProc(func(e ssim.Change) {
						output += "1"
					}),
					newChangeProc(func(e ssim.Change) {
						output += "2"
					}),
					newChangeProc(func(e ssim.Change) {
						output += "3"
					}),
				)

				pipe.Write(nil)
				c.Expect(output, Equals, "123")
			})
		})
	})
}

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribePipelines)

	gospec.MainGoTest(r, t)
}
