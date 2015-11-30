package client_test

import (
	"testing"

	"github.com/ghthor/filu/net/prototest"
	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(prototest.DescribeClientServerProtocol)

	gospec.MainGoTest(r, t)
}
