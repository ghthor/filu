package ssim_test

import (
	"github.com/ghthor/filu/ssim"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func newEventProc(fn func(ssim.Event)) ssim.EventStream {
	return ssim.NewEventProccessorFn(fn)
}

func DescribePipelines(c gospec.Context) {
	c.Specify("a pipeline", func() {
		c.Specify("can be made from", func() {
			c.Specify("event streams", func() {
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

			var streams []ssim.EventStream
			defer func() {
				for _, s := range streams {
					switch s := s.(type) {
					case ssim.SelectableEventStream:
						s.HaltStream()
					default:
					}
				}
			}()

			c.Specify("synced event streams", func() {
				var output string
				streams = []ssim.EventStream{
					ssim.NewSyncedEventStream(newEventProc(func(e ssim.Event) {
						output += "1"
					})),
					ssim.NewSyncedEventStream(newEventProc(func(e ssim.Event) {
						output += "2"
					})),
					ssim.NewSyncedEventStream(newEventProc(func(e ssim.Event) {
						output += "3"
					})),
				}

				pipe := ssim.NewEventPipeline(streams...)

				out := newMockSyncedEventWriter()
				pipe.Subscribe(out)
				pipe.Write(mockEvent{})
				<-out.lastWrite

				c.Expect(output, Equals, "123")
			})

			c.Specify("a mixture", func() {
				var output string
				streams = []ssim.EventStream{
					ssim.NewSyncedEventStream(newEventProc(func(e ssim.Event) {
						output += "1"
					})),
					newEventProc(func(e ssim.Event) {
						output += "2"
					}),
					ssim.NewSyncedEventStream(newEventProc(func(e ssim.Event) {
						output += "3"
					})),
				}

				pipe := ssim.NewEventPipeline(streams...)

				out := newMockSyncedEventWriter()
				pipe.Subscribe(out)
				pipe.Write(mockEvent{})
				<-out.lastWrite

				c.Expect(output, Equals, "123")
			})
		})
	})
}
