package rpg2d

import (
	"bufio"
	"bytes"
	"errors"
	"strings"

	. "github.com/ghthor/engine/rpg2d/coord"
)

// Represents a type of terrain in the world.
type TerrainType rune

const (
	TT_GRASS TerrainType = 'G'
	TT_DIRT  TerrainType = 'D'
	TT_ROCK  TerrainType = 'R'
)

// A terrain map is a dense store of terrain state.
// Ever cell in the world has a terrain type.
type TerrainMap struct {
	Bounds Bounds
	// y, x
	TerrainTypes [][]TerrainType
}

// A change to the terrain type of a cell.
type TerrainTypeChange struct {
	Cell        Cell        `json:"cell"`
	TerrainType TerrainType `json:"type"`
}

// TODO extract the errors this constructor returns
// into static error values.
func NewTerrainMap(bounds Bounds, s string) (TerrainMap, error) {
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

// Return the terrain type in a given cell.
func (m TerrainMap) Cell(c Cell) TerrainType {
	x := c.X - m.Bounds.TopL.X
	y := -(c.Y - m.Bounds.TopL.Y)

	return m.TerrainTypes[y][x]
}

// Return a slice of terrain within a given bounds.
func (m TerrainMap) Slice(bounds Bounds) TerrainMap {
	bounds, err := m.Bounds.Intersection(bounds)
	if err != nil {
		panic("invalid terrain map slicing operation: " + err.Error())
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
	}
}

// Create a copy of the terrain map.
// The copy will not share memory with the source.
func (m TerrainMap) Clone() (TerrainMap, error) {
	if m.TerrainTypes == nil {
		return m, nil
	}

	// LoL this is lazy, but it's ok cause this method isn't important right now
	return NewTerrainMap(m.Bounds, m.String())
}

// Produce a string representation of the terrain map.
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

// Produce a terrain map state with the given terrain map.
func (m TerrainMap) ToState() *TerrainMapState {
	return &TerrainMapState{m}
}
