package time

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

type TimeSpan struct {
	Start, End WorldTime
	Duration   int64
}

func NewTimeSpan(start, end WorldTime) TimeSpan {
	return TimeSpan{
		start,
		end,
		int64(end) - int64(start),
	}
}

func (a TimeSpan) Remaining(from WorldTime) int64 {
	return int64(a.End) - int64(from)
}

func (a TimeSpan) Contains(t WorldTime) bool {
	return a.Start <= t && t <= a.End
}

func (a TimeSpan) Overlaps(other TimeSpan) bool {
	return a.Contains(other.Start) ||
		a.Contains(other.End) ||
		other.Contains(a.Start) ||
		other.Contains(a.End)
}
