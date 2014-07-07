package perlin2d

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"log"
	"testing"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeStaticGeneration)

	gospec.MainGoTest(r, t)
}

func DescribeStaticGeneration(c gospec.Context) {
	seed := int64(4859)
	static := NewStatic(5, 6, seed)
	c.Assume(static, Not(IsNil))
	c.Assume(len(static), Equals, 6)
	c.Assume(len(static[0]), Equals, 5)

	c.Specify("static with the same height, width, and seed should be equal", func() {
		sameSeed := NewStatic(5, 6, seed)
		c.Assume(sameSeed, Not(IsNil))
		c.Assume(len(sameSeed), Equals, len(static))
		c.Assume(len(sameSeed[0]), Equals, len(static[0]))

		for y, row := range static {
			for x, _ := range row {
				c.Expect(static[y][x], IsWithin(1e-100), sameSeed[y][x])
			}
		}
	})

	c.Specify("testing", func() {
		for freq := 1; freq < 10; freq++ {
			for octaves := freq; octaves < freq*40; octaves++ {
				min, max := 1e3, -1e3
				total, totalCells := 0.0, 0

				for i := int64(0); i < 1000; i += 1 {
					noise := NewNoise(NewStatic(100, 100, i), float64(freq), float64(octaves))

					for _, row := range noise {
						for _, val := range row {
							if val < min {
								min = val
							}

							if val > max {
								max = val
							}

							total += val
							totalCells++
						}
					}
				}

				log.Printf("\nFreq:\t%v\nOctaves:\t%v\nAverage:\t%6.5f\nMinimum:\t%6.5f\nMaximum:\t%6.5f", freq, octaves, total/float64(totalCells), min, max)
			}
		}
	})
}
