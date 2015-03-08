package entity_test

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/ghthor/engine/rpg2d/entity"
	"github.com/ghthor/engine/rpg2d/entity/entitytest"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func init() {
	gob.Register(entity.StateSlice{})
	gob.Register(entitytest.MockEntityState{})
}

func DescribeStateSlice(c gospec.Context) {
	c.Specify("a slice of entity states", func() {
		slice := entity.StateSlice{
			entitytest.MockEntity{EntityId: 0}.ToState(),
			entitytest.MockEntity{EntityId: 1}.ToState(),
		}

		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		c.Specify("can be binary marshaled", func() {
			enc := gob.NewEncoder(buf)
			c.Expect(enc.Encode(slice), IsNil)
		})

		c.Specify("can be binary unmarshaled", func() {
			enc := gob.NewEncoder(buf)
			c.Assume(enc.Encode(slice), IsNil)

			dec := gob.NewDecoder(buf)
			var decodeSlice entity.StateSlice
			c.Assume(dec.Decode(&decodeSlice), IsNil)

			c.Expect(decodeSlice, ContainsExactly, slice)
		})
	})
}

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeStateSlice)

	gospec.MainGoTest(r, t)
}
