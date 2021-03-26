package worldterrain_test

import (
	"testing"

	"github.com/ghthor/filu/rpg2d/worldterrain"

	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(worldterrain.DescribeTerrainMap)

	gospec.MainGoTest(r, t)
}
