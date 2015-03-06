package protocol

import (
	"testing"

	"github.com/ghthor/gospec"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeConn)

	gospec.MainGoTest(r, t)
}
