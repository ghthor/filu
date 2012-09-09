package engine

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
)

type (
	TerrainType rune

	TerrainMap struct {
		Bounds AABB
		// y, x
		TerrainTypes [][]TerrainType
	}

	TerrainMapJson struct {
		// Used to calculate diff's
		TerrainMap `json:"-"`

		// A Slice of new terrain the client doesn't have
		Bounds  *AABB  `json:"bounds,omitempty"`
		Terrain string `json:"terrain,omitempty"`

		// An array of type changes
		Changes []TerrainTypeChange `json:"changes,omitempty"`
	}

	TerrainTypeChange struct {
		Cell        Cell        `json:"cell"`
		TerrainType TerrainType `json:"type"`
	}
)

const (
	TT_GRASS TerrainType = 'G'
	TT_DIRT  TerrainType = 'D'
	TT_ROCK  TerrainType = 'R'
)

func NewTerrainMap(bounds AABB, s string) (TerrainMap, error) {
	if len(s) == 0 {
		return TerrainMap{}, errors.New("invalid TerrainType")
	}

	w, h := bounds.Width(), bounds.Height()
	tm := make([][]TerrainType, h)

	if len(s) == 1 {
		for row, _ := range tm {
			tm[row] = make([]TerrainType, w)
		}

		for _, row := range tm {
			for x, _ := range row {
				row[x] = TerrainType(s[0])
			}
		}
	} else {
		s = strings.TrimLeft(s, "\n")
		if strings.Count(s, "\n") != h {
			return TerrainMap{}, errors.New("bounds height doesn't match num lines")
		}

		buf := bufio.NewReader(strings.NewReader(s))
		for y, _ := range tm {
			row := make([]TerrainType, 0, w)
			rowStr, err := buf.ReadString("\n"[0])
			if err != nil {
				return TerrainMap{}, err
			}

			rowStr = strings.TrimRight(rowStr, "\n")

			if len(rowStr) != w {
				return TerrainMap{}, errors.New("bounds width doesn't match line width")
			}

			for _, c := range rowStr {
				row = append(row, TerrainType(c))
			}
			tm[y] = row
		}
	}
	return TerrainMap{bounds, tm}, nil
}

func (m TerrainMap) Cell(c Cell) TerrainType {
	x := c.X - m.Bounds.TopL.X
	y := -(c.Y - m.Bounds.TopL.Y)

	return m.TerrainTypes[y][x]
}

func (m TerrainMap) Slice(bounds AABB) (TerrainMap, error) {
	// Check if the slice asked for is contained within my bounds
	if !m.Bounds.Contains(bounds.TopL) || !m.Bounds.Contains(bounds.BotR) {
		return m, errors.New("invalid terrain slice")
	}

	x := bounds.TopL.X - m.Bounds.TopL.X
	y := -(bounds.TopL.Y - m.Bounds.TopL.Y)
	w, h := bounds.Width(), bounds.Height()
	rows := make([][]TerrainType, h)

	for i, row := range m.TerrainTypes[y : y+h] {
		rows[i] = row[x : x+w]
	}

	return TerrainMap{
		bounds,
		rows,
	}, nil
}

func (m TerrainMap) String() string {
	w, h := len(m.TerrainTypes[0]), len(m.TerrainTypes)
	w += 1 // For \n char

	buf := bytes.NewBuffer(make([]byte, 0, w*h+1))
	buf.WriteRune('\n')
	for _, row := range m.TerrainTypes {
		for _, t := range row {
			buf.WriteRune(rune(t))
		}
		buf.WriteRune('\n')
	}

	return buf.String()
}

func (m TerrainMap) Json() *TerrainMapJson {
	return &TerrainMapJson{
		TerrainMap: m,
	}
}

// Prepare to be Marshalled
func (m *TerrainMapJson) Prepare() {
	// Set the bounds
	m.Bounds = &m.TerrainMap.Bounds
	// Write out the Map as a string
	m.Terrain = m.TerrainMap.String()
}

func (m *TerrainMapJson) IsEmpty() bool {
	if m == nil {
		return true
	}
	return m.TerrainMap.TerrainTypes == nil
}

func (m *TerrainMapJson) Diff(other *TerrainMapJson) (diff *TerrainMapJson) {
	if m.IsEmpty() {
		return other
	}

	maabb, oaabb := m.TerrainMap.Bounds, other.TerrainMap.Bounds
	if maabb == oaabb {
		// No Overlaps
	} else {

		// Find the non overlapped section and set that in the diff
		switch {
		// Overlap top or bottom
		case maabb.Width() == oaabb.Width() &&
			maabb.TopL.X == oaabb.TopL.X &&
			maabb.BotR.X == oaabb.BotR.X:

			if maabb.Height() != oaabb.Height() {
				panic("invalid diff attempt")
			}

			// Overlaps the top
			if oaabb.TopL.Y > maabb.TopL.Y {
				slice, err := other.Slice(AABB{
					oaabb.TopL,
					Cell{oaabb.BotR.X, maabb.TopL.Y + 1},
				})

				if err != nil {
					panic("invalid diff attempt")
				}

				diff = &TerrainMapJson{TerrainMap: slice}

			} else if oaabb.BotR.Y < maabb.BotR.Y {
				// Overlaps the bottom
				slice, err := other.Slice(AABB{
					Cell{oaabb.TopL.X, maabb.BotR.Y - 1},
					oaabb.BotR,
				})

				if err != nil {
					panic("invalid diff attempt")
				}

				diff = &TerrainMapJson{TerrainMap: slice}
			} else {
				panic("invalid diff attempt")
			}

			// Overlaps left of right
		case maabb.Height() == oaabb.Height() &&
			maabb.TopL.Y == oaabb.TopL.Y &&
			maabb.BotR.Y == oaabb.BotR.Y:

			if maabb.Width() != oaabb.Width() {
				panic("invalid diff attempt")
			}

			// Overlaps the left
			if oaabb.TopL.X < maabb.TopL.X {
				slice, err := other.Slice(AABB{
					oaabb.TopL,
					Cell{maabb.TopL.X - 1, oaabb.BotR.Y},
				})

				if err != nil {
					panic("invalid diff attempt")
				}

				diff = &TerrainMapJson{TerrainMap: slice}
			} else if oaabb.BotR.X > maabb.BotR.X {
				// Overlaps the right
				slice, err := other.Slice(AABB{
					Cell{maabb.BotR.X + 1, oaabb.TopL.Y},
					oaabb.BotR,
				})

				if err != nil {
					panic("invalid diff attempt")
				}

				diff = &TerrainMapJson{TerrainMap: slice}
			} else {
				panic("invalid diff attempt")
			}

		default:
			panic("invalid diff attempt")
		}
	}
	return
}
