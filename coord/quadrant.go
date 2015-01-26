//go:generate stringer -type=Quad -output=quadrant_string.go

package coord

type Quad int

const (
	NW Quad = iota
	QUAD_NE
	QUAD_SE
	QUAD_SW
)
