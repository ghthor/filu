package engine

type DiffConn struct {
	JsonOutputConn
	lastState WorldStateJson
}

func (c *DiffConn) SendJson(msg string, nextState interface{}) error {
	diff := c.lastState.Diff(nextState.(WorldStateJson))
	c.lastState = nextState.(WorldStateJson)

	// Will need this when I start comparing for TerrainType changes
	//c.lastState.Prepare()

	if len(diff.Entities) > 0 || len(diff.Removed) > 0 || diff.TerrainMap != nil {
		// Prepare the state for sending
		diff.Prepare()
		c.JsonOutputConn.SendJson(msg, diff)
	}
	return nil
}
