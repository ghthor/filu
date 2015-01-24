package quad_test

import (
	"testing"

	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeMockEntities)

	r.AddSpec(DescribeQuad)
	r.AddSpec(DescribeQuadInsert)

	gospec.MainGoTest(r, t)
}
