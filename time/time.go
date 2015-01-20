package time

type (
	// Represents an Instant in Time
	// Spans of Time are represented by int64
	Time  int64
	Clock Time
)

func (c Clock) Now() Time {
	return Time(c)
}

func (c Clock) Tick() Clock {
	return Clock(int64(c) + 1)
}

func (c Clock) Future(mag int64) Time {
	f := Time(int64(c) + mag)
	return f
}

type Span struct {
	Start, End Time
	Duration   int64
}

func NewSpan(start, end Time) Span {
	return Span{
		start,
		end,
		int64(end) - int64(start),
	}
}

func (a Span) Remaining(from Time) int64 {
	return int64(a.End) - int64(from)
}

func (a Span) Contains(t Time) bool {
	return a.Start <= t && t <= a.End
}

func (a Span) Overlaps(other Span) bool {
	return a.Contains(other.Start) ||
		a.Contains(other.End) ||
		other.Contains(a.Start) ||
		other.Contains(a.End)
}
