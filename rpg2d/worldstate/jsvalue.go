// +build js,wasm

package worldstate

import (
	"syscall/js"

	"github.com/ghthor/filu/rpg2d/quad/quadstate"
)

func (s *Snapshot) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Time", int64(s.Time))
	v.Set("Bounds", s.Bounds.JSValue())
	v.Set("EntitiesRemoved", quadstate.EntitiesJSValue(s.Entities.Removed))
	v.Set("EntitiesNew", quadstate.EntitiesJSValue(s.Entities.New))
	v.Set("EntitiesChanged", quadstate.EntitiesJSValue(s.Entities.Changed))
	v.Set("EntitiesUnchanged", quadstate.EntitiesJSValue(s.Entities.Unchanged))
	v.Set("TerrainMap", s.TerrainMap.JSValue())
	return v
}

func (s *Update) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Time", int64(s.Time))
	v.Set("Bounds", s.Bounds.JSValue())

	v.Set("Entities", quadstate.EntitiesJSValue(s.Entities))
	v.Set("Removed", quadstate.EntitiesJSValue(s.Removed))

	a := js.Global().Get("Array").New(len(s.RemovedIds))
	for i, id := range s.RemovedIds {
		a.SetIndex(i, int64(id))
	}
	v.Set("RemovedIds", a)

	if s.TerrainMapSlices == nil || len(s.TerrainMapSlices.Slices) <= 0 {
		v.Set("TerrainMapSlices", js.Null())
	} else {
		a := js.Global().Get("Array").New(len(s.TerrainMapSlices.Slices))
		for i, slice := range s.TerrainMapSlices.Slices {
			a.SetIndex(i, slice)
		}
		vv := js.Global().Get("Object").New()
		vv.Set("Bounds", s.TerrainMapSlices.Bounds)
		vv.Set("Slices", a)
		v.Set("TerrainMapSlices", vv)
	}

	return v
}
