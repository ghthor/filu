package quadstate

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"testing"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/entity/entitytest"
)

func init() {
	StateEncode = func(s entity.State, w io.Writer) error {
		enc := json.NewEncoder(w)
		return enc.Encode(s)
	}

	StateDecode = func(data []byte) (entity.State, error) {
		var e entitytest.MockEntityState
		err := json.Unmarshal(data, &e)
		if err != nil {
			return nil, err
		}

		return e, nil
	}
}

func TestCachedGobEncodingEntity(t *testing.T) {
	input := &Entity{
		State: entitytest.MockEntityState{
			Id:   entity.Id(100),
			Name: "testEntity",
			Cell: coord.Cell{-27, 27},
		},
	}

	if input.EncodingIsCached() {
		t.Fatalf("assumption failed: cache should be empty")
	}

	func() {
		var b bytes.Buffer
		enc := gob.NewEncoder(&b)
		err := enc.Encode(input)
		if err != nil {
			t.Fatalf("encoding error: %v", err)
		}

		if !input.EncodingIsCached() {
			t.Fatalf("encoding failed to cache")
		}

		var output Entity
		dec := gob.NewDecoder(&b)
		err = dec.Decode(&output)
		if err != nil {
			t.Fatalf("decoding error: %v", err)
		}

		if input.State.(entity.State) != output.State.(entity.State) {
			t.Fatalf("%#v != %#v", input, output)
		}
	}()

	func() {
		if !input.EncodingIsCached() {
			t.Fatalf("encoding failed to cache")
		}

		var b bytes.Buffer
		enc := gob.NewEncoder(&b)
		err := enc.Encode(input)
		if err != nil {
			t.Fatalf("encoding error: %v", err)
		}

		var output Entity
		dec := gob.NewDecoder(&b)
		err = dec.Decode(&output)
		if err != nil {
			t.Fatalf("decoding error: %v", err)
		}

		if input.State.(entity.State) != output.State.(entity.State) {
			t.Fatalf("%#v != %#v", input, output)
		}
	}()
}

func TestEntityEncoderCache(t *testing.T) {
	e := NewEntityEncoder()
	e.cache[0] = &bytes.Buffer{}
	e.FreeBufferFor(0)

	if len(e.freelist) != 1 {
		t.Fatalf("expected freelist length of 1")
	}

	e.newBuffer()

	if len(e.freelist) != 0 {
		t.Fatalf("expected freelist length of 0")
	}

	if _, exists := e.cache[0]; exists {
		t.Fatalf("expected cached buffer to not exist")
	}
}
