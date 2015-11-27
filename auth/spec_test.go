package auth_test

import (
	"testing"

	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeStream)

	gospec.MainGoTest(r, t)
}
