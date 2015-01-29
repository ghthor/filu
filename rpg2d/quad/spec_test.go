package quad_test

import (
	"testing"

	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeMockEntities)
	r.AddSpec(DescribeChunkMatching)

	r.AddSpec(DescribeCollision)
	r.AddSpec(DescribeCollisionGroup)

	r.AddSpec(DescribeQuad)
	r.AddSpec(DescribeQuadInsert)

	r.AddSpec(DescribePhase)

	gospec.MainGoTest(r, t)
}
