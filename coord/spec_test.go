package coord

import (
	"testing"

	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeCell)
	r.AddSpec(DescribeCellCollision)

	r.AddSpec(DescribeDirection)

	r.AddSpec(DescribeMoveAction)
	r.AddSpec(DescribePathAction)
	r.AddSpec(DescribePathCollision)

	r.AddSpec(DescribeAABB)

	gospec.MainGoTest(r, t)
}
