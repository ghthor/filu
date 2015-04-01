package ssim_test

import (
	"testing"

	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeMemEventLog)

	r.AddSpec(DescribePipelines)
	r.AddSpec(DescribeSyncedStream)

	gospec.MainGoTest(r, t)
}
