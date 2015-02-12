package quad_test

import (
	"testing"

	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeChunkMatching)

	r.AddSpec(DescribeCollision)
	r.AddSpec(DescribeCollisionIndex)
	r.AddSpec(DescribeCollisionGroup)
	r.AddSpec(DescribeCollisionGroupIndex)

	r.AddSpec(DescribeQuad)
	r.AddSpec(DescribeQuadInsert)

	r.AddSpec(DescribePhase)

	gospec.MainGoTest(r, t)
}
