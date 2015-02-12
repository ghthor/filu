package rpg2d_test

import (
	"testing"

	"github.com/ghthor/engine/rpg2d"

	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(rpg2d.DescribeTerrainMap)
	r.AddSpec(DescribeASimulation)

	gospec.MainGoTest(r, t)
}
