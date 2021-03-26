// +build js,wasm

package rpg2d

import "syscall/js"

func (c TerrainTypeChange) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Cell", c.Cell.JSValue())
	v.Set("TerrainType", string(c.TerrainType))
	return v
}

func (m TerrainMapState) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Bounds", m.Bounds)
	v.Set("Terrain", m.TerrainMap.String())
	return v
}

func (m TerrainMapStateSlice) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Bounds", m.Bounds)
	v.Set("Terrain", m.Terrain)
	return v
}

func (s WorldState) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Time", int64(s.Time))
	v.Set("Bounds", s.Bounds.JSValue())
	v.Set("EntitiesRemoved", s.EntitiesRemoved.JSValue())
	v.Set("EntitiesNew", s.EntitiesNew.JSValue())
	v.Set("EntitiesChanged", s.EntitiesChanged.JSValue())
	v.Set("EntitiesUnchanged", s.EntitiesUnchanged.JSValue())
	v.Set("TerrainMap", s.TerrainMap.JSValue())
	return v
}

func (s WorldStateDiff) JSValue() js.Value {
	v := js.Global().Get("Object").New()
	v.Set("Time", int64(s.Time))
	v.Set("Bounds", s.Bounds.JSValue())

	v.Set("Entities", s.Entities.JSValue())
	v.Set("Removed", s.Removed.JSValue())

	if len(s.TerrainMapSlices) > 0 {
		a := js.Global().Get("Array").New(len(s.TerrainMapSlices))
		for i, slice := range s.TerrainMapSlices {
			a.SetIndex(i, slice)
		}
		v.Set("TerrainMapSlices", a)
	} else {
		v.Set("TerrainMapSlices", js.Null())
	}

	return v
}
