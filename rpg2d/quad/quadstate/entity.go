package quadstate

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/ghthor/filu/rpg2d/entity"
)

type Type uint

const (
	TypeRemoved Type = 1 << iota
	TypeNew
	TypeChanged
	TypeUnchanged
)

type Entity struct {
	entity.State
	Type
	cachedEncoding []byte
}

type EntityEncodedType struct {
	entity.State
}

type Entities struct {
	Removed,
	New,
	Unchanged,
	Changed []*Entity
}

func NewEntities(size int) *Entities {
	return &Entities{
		Removed:   make([]*Entity, 0, size),
		New:       make([]*Entity, 0, size),
		Changed:   make([]*Entity, 0, size),
		Unchanged: make([]*Entity, 0, size),
	}
}

// TODO Merge this is Add implementation
func (entities *Entities) Insert(e *Entity) {
	switch {
	case e.Type&TypeRemoved != 0:
		entities.Removed = append(entities.Removed, e)
	case e.Type&TypeNew != 0:
		entities.New = append(entities.New, e)
	case e.Type&TypeChanged != 0:
		entities.Changed = append(entities.Changed, e)
	case e.Type&TypeUnchanged != 0:
		entities.Unchanged = append(entities.Unchanged, e)
	default:
		panic(fmt.Sprintf("unknown entity state type %#v", e))
	}
}

func (e *Entities) Clear() {
	e.Removed =
		e.Removed[:0]
	e.New =
		e.New[:0]
	e.Changed =
		e.Changed[:0]
	e.Unchanged =
		e.Unchanged[:0]
}

var _ Accumulator = &Entities{}

func (entities *Entities) Add(e *Entity) {
	switch {
	case e.Type&TypeRemoved != 0:
		entities.Removed = append(entities.Removed, e)
	case e.Type&TypeNew != 0:
		entities.New = append(entities.New, e)
	case e.Type&TypeChanged != 0:
		entities.Changed = append(entities.Changed, e)
	case e.Type&TypeUnchanged != 0:
		entities.Unchanged = append(entities.Unchanged, e)
	default:
		panic(fmt.Sprintf("unknown entity state type %#v", e))
	}
}

func (entities *Entities) AddSlice(others []*Entity, t Type) {
	switch {
	case t&TypeRemoved != 0:
		entities.Removed = append(entities.Removed, others...)
	case t&TypeNew != 0:
		entities.New = append(entities.New, others...)
	case t&TypeChanged != 0:
		entities.Changed = append(entities.Changed, others...)
	case t&TypeUnchanged != 0:
		entities.Unchanged = append(entities.Unchanged, others...)
	default:
		panic(fmt.Sprintf("unknown entity state type %#v", t))
	}
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
