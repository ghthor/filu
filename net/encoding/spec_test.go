package encoding

import (
	"github.com/ghthor/gospec"
	"testing"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribePacket)

	gospec.MainGoTest(r, t)
}
