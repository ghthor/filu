package engine

import (
	"github.com/ghthor/gospec/src/gospec"
	. "github.com/ghthor/gospec/src/gospec"
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
