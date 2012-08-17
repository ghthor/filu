package engine_test

import (
	".././engine"
	"github.com/ghthor/gospec/src/gospec"
	"testing"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(engine.DescribeClock)
	r.AddSpec(engine.DescribeAction)
	r.AddSpec(engine.DescribeDirection)
	r.AddSpec(engine.DescribePathAction)
	r.AddSpec(engine.DescribeCollision)

	gospec.MainGoTest(r, t)
}
