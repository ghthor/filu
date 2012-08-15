package world

type (
	// Represents an Instant in Time
	// Spans of Time are represented by int64
	WorldTime int64
	Clock     WorldTime
)

func (c Clock) Now() WorldTime {
	return WorldTime(c)
}

func (c Clock) Tick() Clock {
	return Clock(int64(c) + 1)
}

func (c Clock) Future(mag int64) WorldTime {
	f := WorldTime(int64(c) + mag)
	return f
}
