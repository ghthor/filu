package engine_test

import (
	"testing"

	"bitbucket.org/ghthor/ages-of-dark/engine"

	"github.com/ghthor/gospec"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(engine.DescribeClock)
	r.AddSpec(engine.DescribeTimeSpan)

	r.AddSpec(engine.DescribeDirection)
	r.AddSpec(engine.DescribeCell)
	r.AddSpec(engine.DescribePathCollision)
	r.AddSpec(engine.DescribeCellCollision)
	r.AddSpec(engine.DescribeAABB)

	r.AddSpec(engine.DescribePathAction)
	r.AddSpec(engine.DescribeMoveAction)
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
	r.AddSpec(engine.DescribePlayerCollisions)
	r.AddSpec(engine.DescribePlayerJson)

	r.AddSpec(engine.DescribeViewPortCulling)

	gospec.MainGoTest(r, t)
}
