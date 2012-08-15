package world_test

import (
	".././world"
	"github.com/orfjackal/gospec/src/gospec"
	"testing"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(world.DescribeClock)
	r.AddSpec(world.DescribeAction)
	r.AddSpec(world.DescribeDirection)
	r.AddSpec(world.DescribePathAction)
	r.AddSpec(world.DescribeCollision)

	gospec.MainGoTest(r, t)
}
