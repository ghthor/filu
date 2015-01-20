package engine_test

import (
	"testing"

	"github.com/ghthor/engine"
	"github.com/ghthor/gospec"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(engine.DescribeMovableEntity)
	r.AddSpec(engine.DescribeEntityCollision)
	r.AddSpec(engine.DescribeMockEntities)

	r.AddSpec(engine.DescribeQuad)
	r.AddSpec(engine.DescribeTerrainMap)
	r.AddSpec(engine.DescribeWorldState)
	r.AddSpec(engine.DescribeDiffConn)
	r.AddSpec(engine.DescribeSimulation)

	r.AddSpec(engine.DescribeInputCommands)
	r.AddSpec(engine.DescribePlayer)
	// r.AddSpec(engine.DescribePlayerCollisions)
	r.AddSpec(engine.DescribePlayerJson)

	r.AddSpec(engine.DescribeViewPortCulling)

	gospec.MainGoTest(r, t)
}
