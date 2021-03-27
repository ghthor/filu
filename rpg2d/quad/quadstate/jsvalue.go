// +build js,wasm

package quadstate

import (
	"syscall/js"
)

func EntitiesJSValue(s []Entity) js.Value {
	a := js.Global().Get("Array").New(len(s))
	for i, _ := range s {
		a.SetIndex(i, s[i].State)
	}
	return a
}
