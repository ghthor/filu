package quadstate

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
)

type QueryFlag uint

const (
	QueryInstant QueryFlag = 1 << iota
	QueryRemoved
	QueryNew
	QueryChanged
	QueryUnchanged
)

const (
	QueryAll  QueryFlag = QueryInstant | QueryRemoved | QueryNew | QueryChanged | QueryUnchanged
	QueryDiff QueryFlag = QueryInstant | QueryRemoved | QueryNew | QueryChanged
)

type Type uint

const (
	TypeRemoved Type = iota
	TypeInstant
	TypeNew
	TypeChanged
	TypeUnchanged
	SizeType
)

var allTypes = [SizeType]struct {
	QueryFlag
	Type
}{
	{QueryInstant, TypeInstant},
	{QueryRemoved, TypeRemoved},
	{QueryNew, TypeNew},
	{QueryChanged, TypeChanged},
	{QueryUnchanged, TypeUnchanged},
}

// TODO Factor out a file split for wasm/other build targets
type Entity struct {
	Type
	entity.Id
	coord.Cell
	entity.State
	cachedEncoding []byte
}

func NewEntity(state entity.State, t Type) *Entity {
	return &Entity{
		t,
		state.EntityId(),
		state.EntityCell(),
		state,
		nil,
	}
}

type ByType [SizeType][]*Entity

type Entities struct {
	ByType
}

func NewEntities(size int) *Entities {
	var arr ByType
	for i := range arr {
		arr[i] = make([]*Entity, 0, size)
	}
	return &Entities{arr}
}

// TODO Merge this with Add implementation
func (entities *Entities) Insert(e *Entity) {
	entities.ByType[e.Type] = append(entities.ByType[e.Type], e)
}

func (e *Entities) Clear() {
	for t := range e.ByType {
		e.ByType[t] = e.ByType[t][:0]
	}
}

var _ Accumulator = &Entities{}

func (entities *Entities) Add(e *Entity) {
	entities.ByType[e.Type] = append(entities.ByType[e.Type], e)
}

func (entities *Entities) AddSlice(others []*Entity, t Type) {
	entities.ByType[t] = append(entities.ByType[t], others...)
}

type EncodingRequest struct {
	Slices    [][]*Entity
	Completed chan<- [][]*Entity
}

// TODO Fix memory leak for entities that no longer exist
//      The QuadTree needs to keep track per clock cycle and
//      provide a list of Entity Id's that no longer exist
type EntityEncoder struct {
	cache    map[entity.Id]*bytes.Buffer
	freelist []*bytes.Buffer
}

func NewEntityEncoder() *EntityEncoder {
	return &EntityEncoder{
		make(map[entity.Id]*bytes.Buffer),
		make([]*bytes.Buffer, 0, 30),
	}
}

func (encoder *EntityEncoder) FreeBufferFor(id entity.Id) {
	if b, exists := encoder.cache[id]; exists {
		if b != nil {
			b.Reset()
			encoder.freelist = append(encoder.freelist, b)
			delete(encoder.cache, id)
		}
	}
}

func (e *EntityEncoder) newBuffer() *bytes.Buffer {
	if len(e.freelist) == 0 {
		return bytes.NewBuffer(make([]byte, 0, 10))
	}

	b := e.freelist[len(e.freelist)-1]
	e.freelist = e.freelist[:len(e.freelist)-1]
	return b
}

var StateEncode = func(entity.State, io.Writer) error {
	panic("quadstate.StateEncode must be set during initialization")
}

var StateDecode = func([]byte) (entity.State, error) {
	panic("quadstate.StateDecode must be set during initialization")
}

func (encoder *EntityEncoder) Start(ctx context.Context, next <-chan EncodingRequest) {
	go func(ctx context.Context, next <-chan EncodingRequest) {
		for {
			select {
			case r := <-next:
				for _, entities := range r.Slices {
					for _, e := range entities {
						encodedBytes, exists := encoder.cache[e.EntityId()]
						if exists {
							// log.Println("Encoding Cache Hit")
							// fmt.Printf("%#v\n", e.State)
							// fmt.Printf("%s\n", encodedBytes.String())
							e.cachedEncoding = encodedBytes.Bytes()
							continue
						} else {
							encodedBytes = encoder.newBuffer()
						}

						err := StateEncode(e.State, encodedBytes)
						if err != nil {
							panic(fmt.Sprint("encoding error", err))
						}

						encoder.cache[e.EntityId()] = encodedBytes
						e.cachedEncoding = encodedBytes.Bytes()

						// // Debugging
						// log.Printf("\n%#v", e.State)
						// fmt.Printf("%s\n", encodedBytes.String())
					}
				}
				r.Completed <- r.Slices

			case <-ctx.Done():
				return
			}
		}
	}(ctx, next)
}

func (e *Entity) MarshalBinary() ([]byte, error) {
	if e.cachedEncoding != nil {
		return e.cachedEncoding, nil
	}

	var buf bytes.Buffer
	err := StateEncode(e.State, &buf)
	if err != nil {
		return nil, err
	}

	e.cachedEncoding = buf.Bytes()
	return e.cachedEncoding, nil
}

func (e *Entity) UnmarshalBinary(data []byte) error {
	state, err := StateDecode(data)
	if err != nil {
		return err
	}

	e.State = state
	return nil
}

func (e *Entity) EncodingIsCached() bool {
	return e.cachedEncoding != nil && len(e.cachedEncoding) > 0
}
