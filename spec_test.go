package engine_test

import (
	".././engine"
	"github.com/ghthor/gospec/src/gospec"
	"testing"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(engine.DescribeClock)
	r.AddSpec(engine.DescribeTimeSpan)

	r.AddSpec(engine.DescribeDirection)
	r.AddSpec(engine.DescribeWorldCoord)
	r.AddSpec(engine.DescribePathCollision)
	r.AddSpec(engine.DescribeCoordCollision)
	r.AddSpec(engine.DescribeAABB)

	r.AddSpec(engine.DescribePathAction)
	r.AddSpec(engine.DescribeMoveAction)
	r.AddSpec(engine.DescribeMovableEntity)
	r.AddSpec(engine.DescribeEntityCollision)
	r.AddSpec(engine.DescribeMockEntities)

	r.AddSpec(engine.DescribeQuad)
	r.AddSpec(engine.DescribeWorldState)
	r.AddSpec(engine.DescribeSimulation)

	r.AddSpec(engine.DescribePlayer)
	r.AddSpec(engine.DescribePlayerCollisions)
	r.AddSpec(engine.DescribeInputCommands)

	gospec.MainGoTest(r, t)
}
