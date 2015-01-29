package quad

// Implement the gospec.Equals interface
func (c Chunk) Equals(other interface{}) bool {
	switch oc := other.(type) {
	case Chunk:
		if c.Bounds == oc.Bounds {
			return chunksContainSameEntities(c, oc)
		}
	default:
	}

	return false
}

func chunksContainSameEntities(c1, c2 Chunk) bool {
	if len(c1.Entities) != len(c2.Entities) {
		return false
	}

forEach:
	for _, e1 := range c1.Entities {
		for _, e2 := range c2.Entities {
			if e1 == e2 {
				continue forEach
			}
		}
		// c2.Entities did not contain e1
		return false
	}
	return true
}
