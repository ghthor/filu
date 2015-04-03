package ssim_test

import (
	"time"

	"github.com/ghthor/filu/ssim"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type mockSyncedEventWriter struct {
	lastWrite chan ssim.Event
}

func (w mockSyncedEventWriter) Write(e ssim.Event) {
	w.lastWrite <- e
}

func newMockSyncedEventWriter() mockSyncedEventWriter {
	return mockSyncedEventWriter{make(chan ssim.Event)}
}

func DescribeSyncedStream(c gospec.Context) {
	c.Specify("a synced stream", func() {
		c.Specify("of events", func() {
			now := time.Now()
			log := ssim.NewMemEventLog(ssim.NowProvider(func() time.Time {
				return now
			}))

			syncedLog := ssim.NewSyncedEventStream(log)

			defer func() {
				c.Assume(syncedLog.HaltStream(), Equals, log)
			}()

			c.Specify("can be written to", func() {
				syncedLog.WriteCh() <- mockEvent{}
			})

			c.Specify("can be subscribed to", func() {
				out := newMockSyncedEventWriter()
				syncedLog.SubscribeCh() <- out
				syncedLog.WriteCh() <- mockEvent{}
				c.Expect(<-out.lastWrite, Equals, mockEvent{recv: now})

				go func() {
					log.Write(mockEvent{})
				}()
				c.Expect(<-out.lastWrite, Equals, mockEvent{recv: now})
			})
		})
	})
}
