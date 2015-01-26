package protocol

import (
	"github.com/ghthor/gospec"
	"testing"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeWebsocketConn)

	gospec.MainGoTest(r, t)
}
