// +build js,wasm

package entity

import (
	"fmt"
	"syscall/js"
)

func (s StateSlice) JSValue() js.Value {
	a := js.Global().Get("Array").New(len(s))
	for i, e := range s {
		switch e := e.(type) {
		case js.Wrapper:
			a.SetIndex(i, e.JSValue())
		default:
			panic(fmt.Sprintf("JSValue not implemented for %#v", e))
		}
	}
	return a
}

func (e RemovedState) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Id", int64(e.Id))
	v.Set("Type", "removed")
	v.Set("Cell", e.Cell)
	return v
}
