// +build js,wasm

package coord

import "syscall/js"

func (c Cell) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("X", c.X)
	v.Set("Y", c.Y)
	return v
}

func (b Bounds) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("TopL", b.TopL.JSValue())
	v.Set("BotR", b.BotR.JSValue())
	return v
}

func (a PathActionState) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Start", int64(a.Start))
	v.Set("End", int64(a.End))
	v.Set("Orig", a.Orig.JSValue())
	v.Set("Dest", a.Dest.JSValue())
	return v
}
