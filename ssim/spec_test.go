package ssim_test

import (
	"testing"

	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribePipelines)
	r.AddSpec(DescribeMemEventLog)

	gospec.MainGoTest(r, t)
}
