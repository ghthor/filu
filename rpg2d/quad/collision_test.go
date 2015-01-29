package quad

func (c Collision) Equals(other interface{}) bool {
	switch oc := other.(type) {
	case Collision:
		return c.IsSameAs(oc)
	default:
	}

	return false
}
