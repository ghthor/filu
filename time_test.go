package engine

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeClock(c gospec.Context) {
	clk := Clock(0)

	c.Specify("Tick moves Clock forward in time", func() {
		for i := 0; i < 100; i++ {
			c.Expect(clk.Now(), Equals, WorldTime(i))
			clk = clk.Tick()
		}
	})

	c.Specify("Future produces WorldTime offset's that are in the future", func() {
		future := clk.Future(10)
		c.Expect(future, Equals, WorldTime(10))

		clk = clk.Tick()
		future = clk.Future(10)
		c.Expect(future, Equals, WorldTime(11))
	})
}

func DescribeTimeSpan(c gospec.Context) {
	clk, duration := Clock(50), int64(100)
	a := NewTimeSpan(clk.Now(), clk.Future(duration))

	c.Specify("TimeLeft reports the full duration when WorldTime is the start of the Action", func() {
		c.Expect(a.Remaining(clk.Now()), Equals, duration)
	})

	c.Specify("Timeleft reports 0 when the WorldTime is the end of the Action", func() {
		clk = Clock(clk.Future(duration))
		c.Expect(a.Remaining(clk.Now()), Equals, int64(0))
	})
}
