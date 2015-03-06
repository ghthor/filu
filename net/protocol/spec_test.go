package protocol

import (
	"testing"

	"github.com/ghthor/gospec"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeConn)
	r.AddSpec(DescribeWebsocketConn)

	gospec.MainGoTest(r, t)
}
