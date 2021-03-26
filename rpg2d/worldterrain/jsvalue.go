// +build js,wasm

package worldterrain

import "syscall/js"

func (c TypeChange) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Cell", c.Cell.JSValue())
	v.Set("TerrainType", string(c.TerrainType))
	return v
}

func (m MapState) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Bounds", m.Bounds)
	v.Set("Terrain", m.Map.String())
	return v
}

func (m MapStateSlice) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Bounds", m.Bounds)
	v.Set("Terrain", m.Terrain)
	return v
}
