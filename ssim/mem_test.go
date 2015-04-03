package ssim_test

import (
	"time"

	"github.com/ghthor/engine/ssim"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type mockEvent struct {
	recv time.Time
}

func (e mockEvent) Source() ssim.ActorID { return 0 }
func (e mockEvent) IssuedAt() time.Time  { return time.Time{} }
func (e mockEvent) AcceptAt(t time.Time) ssim.Event {
	e.recv = t
	return e
}

type mockEventWriter struct {
	lastWrite ssim.Event
}

func (w *mockEventWriter) Write(e ssim.Event) {
	w.lastWrite = e
}

func DescribeMemEventLog(c gospec.Context) {
	c.Specify("an event log", func() {
		now := time.Now()

		l := ssim.NewMemEventLog(ssim.NowProvider(func() time.Time {
			return now
		}))

		outStream := mockEventWriter{}

		l.Subscribe(&outStream)

		c.Specify("can receive events", func() {
			l.Write(mockEvent{})
			c.Expect(outStream.lastWrite, Equals, mockEvent{recv: now})

			c.Specify("and will set the time received on the event", func() {
				var err error
				now, err = time.Parse("2006-Jan-01", "2015-Apr-01")
				c.Assume(err, IsNil)

				l.Write(mockEvent{})
				c.Expect(outStream.lastWrite, Equals, mockEvent{recv: now})
			})

			c.Specify("and will publish the event to subscribers", func() {
				outStreams := []*mockEventWriter{
					&outStream,
					&mockEventWriter{},
					&mockEventWriter{},
				}

				for _, w := range outStreams[1:] {
					l.Subscribe(w)
				}

				l.Write(mockEvent{})
				for _, w := range outStreams {
					c.Expect(w.lastWrite, Equals, mockEvent{recv: now})
				}
			})
		})
	})
}
