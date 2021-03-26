package worldterrain

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/ghthor/filu/rpg2d/coord"
)

// Represents a type of terrain in the world.
type Type rune

const (
	TT_UNKNOWN Type = 'U'
	TT_GRASS   Type = 'G'
	TT_DIRT    Type = 'D'
	TT_ROCK    Type = 'R'
)

type Type2dArray [][]Type

// A terrain map is a dense store of terrain state.
// Ever cell in the world has a terrain type.
type Map struct {
	Bounds coord.Bounds
	// y, x
	Types Type2dArray
}

// A change to the terrain type of a cell.
type TypeChange struct {
	Cell        coord.Cell `json:"cell"`
	TerrainType Type       `json:"type"`
}

func NewType2dArray(bounds coord.Bounds, s string) (Type2dArray, error) {
	if len(s) == 0 {
		return nil, errors.New("invalid [][]TerrainType defination")
	}

	w, h := bounds.Width(), bounds.Height()
	tm := make([][]Type, h)

	if len(s) == 1 {
		for row, _ := range tm {
			tm[row] = make([]Type, w)
		}

		for _, row := range tm {
			for x, _ := range row {
				row[x] = Type(s[0])
			}
		}
	} else {
		s = strings.TrimLeft(s, "\n")
		if strings.Count(s, "\n") != h {
			return nil, errors.New("bounds height doesn't match num lines")
		}

		buf := bufio.NewReader(strings.NewReader(s))
		for y, _ := range tm {
			row := make([]Type, 0, w)
			rowStr, err := buf.ReadString("\n"[0])
			if err != nil {
				return nil, err
			}

			rowStr = strings.TrimRight(rowStr, "\n")

			if len(rowStr) != w {
				return nil, errors.New("bounds width doesn't match line width")
			}

			for _, c := range rowStr {
				row = append(row, Type(c))
			}
			tm[y] = row
		}
	}

	return tm, nil
}

// TODO extract the errors this constructor returns
// into static error values.
func NewMap(bounds coord.Bounds, s string) (Map, error) {
	tm, err := NewType2dArray(bounds, s)
	if err != nil {
		return Map{}, err
	}

	return Map{bounds, tm}, nil
}

// Return the terrain type in a given cell.
func (m Map) Cell(c coord.Cell) Type {
	x := c.X - m.Bounds.TopL.X
	y := -(c.Y - m.Bounds.TopL.Y)

	return m.Types[y][x]
}

func (m *Map) SetType(t Type, c coord.Cell) {
	x := c.X - m.Bounds.TopL.X
	y := -(c.Y - m.Bounds.TopL.Y)

	m.Types[y][x] = t
}

// Return a slice of terrain within a given bounds.
// This method doesn't copy any memory.
// The slice is viewport into the same memeory as
// the map it is sliced from.
func (m Map) Slice(bounds coord.Bounds) Map {
	bounds, err := m.Bounds.Intersection(bounds)
	if err != nil {
		panic("invalid terrain map slicing operation: " + err.Error())
	}

	x := bounds.TopL.X - m.Bounds.TopL.X
	y := -(bounds.TopL.Y - m.Bounds.TopL.Y)
	w, h := bounds.Width(), bounds.Height()
	rows := make([][]Type, h)

	for i, row := range m.Types[y : y+h] {
		rows[i] = row[x : x+w]
	}

	return Map{
		bounds,
		rows,
	}
}

// Create a copy of the terrain map.
// The copy will not share memory with the source.
func (m Map) Clone() (Map, error) {
	if m.Types == nil {
		return m, nil
	}

	// TODO LoL this is lazy, but it's ok cause this method isn't important right now
	return NewMap(m.Bounds, m.String())
}

func (a Type2dArray) String() string {
	w, h := len(a[0]), len(a)
	w += 1 // For \n char

	buf := bytes.NewBuffer(make([]byte, 0, w*h+1))
	buf.WriteRune('\n')
	for _, row := range a {
		for _, t := range row {
			buf.WriteRune(rune(t))
		}
		buf.WriteRune('\n')
	}

	return buf.String()
}

// Produce a string representation of the terrain map.
func (m Map) String() string {
	return m.Types.String()
}

// Produce a terrain map state with the given terrain map.
func (m Map) ToState() *MapState {
	return &MapState{m}
}

// MergeDiff will merge the slices of terrain into
// the TerrainMap. TerrainMap will have the bounds of
// newBounds once the operation is complete. If slices
// are unmergable MergeDiff will return an error.
func (m *Map) MergeDiff(newBounds coord.Bounds, slices ...MapStateSlice) error {
	maps := make([]Map, 0, len(slices)+1)
	for _, slice := range slices {
		m, err := NewMap(slice.Bounds, slice.Terrain)
		if err != nil {
			return err
		}
		maps = append(maps, m)
	}
	maps = append(maps, m.Slice(newBounds))

	joined, err := JoinTerrain(newBounds, maps...)
	if err != nil {
		return err
	}

	*m = joined

	return nil
}

func JoinTerrain(newBounds coord.Bounds, maps ...Map) (Map, error) {
	switch len(maps) {
	case 2:
		b0 := maps[0].Bounds
		b1 := maps[1].Bounds

		switch {
		case b0.TopL.X == b1.TopL.X:
			var top, bot Map

			switch {
			case b0.TopL == newBounds.TopL || b0.TopR() == newBounds.TopR():
				top, bot = maps[0], maps[1]
			case b0.BotL() == newBounds.BotL() || b0.BotR == newBounds.BotR:
				top, bot = maps[1], maps[0]
			}
			return join2vert(newBounds, top, bot), nil

		case b0.TopL.Y == b1.TopL.Y:
			var left, right Map

			switch {
			case b0.TopL == newBounds.TopL || b0.BotL() == newBounds.BotL():
				left, right = maps[0], maps[1]
			case b0.TopR() == newBounds.TopR() || b0.BotR == newBounds.BotR:
				left, right = maps[1], maps[0]
			}
			return join2horz(newBounds, left, right), nil

		default:
			return Map{}, fmt.Errorf("invalid 2 terrain join: %v", maps)
		}
	case 4:
		var tl, tr, br, bl Map
		for _, m := range maps {
			switch {
			case m.Bounds.TopL == newBounds.TopL:
				tl = m
			case m.Bounds.TopR() == newBounds.TopR():
				tr = m
			case m.Bounds.BotR == newBounds.BotR:
				br = m
			case m.Bounds.BotL() == newBounds.BotL():
				bl = m

			default:
				return Map{}, fmt.Errorf("invalid 4 terrain join: %v", maps)
			}
		}

		return join4(newBounds, tl, tr, br, bl), nil

	default:
		return Map{}, fmt.Errorf("unsupported terrain map join: %v", maps)
	}
}

func join2horz(newBounds coord.Bounds, left, right Map) Map {
	for y, row := range right.Types {
		left.Types[y] = append(left.Types[y], row...)
	}

	left.Bounds.TopL = left.Bounds.TopL
	left.Bounds.BotR = right.Bounds.BotR
	return left
}

func join2vert(newBounds coord.Bounds, top, bot Map) Map {
	top.Types = append(top.Types, bot.Types...)
	top.Bounds.BotR = bot.Bounds.BotR
	return top
}

func join4(newBounds coord.Bounds, tl, tr, br, bl Map) Map {
	return join2vert(newBounds,
		join2horz(newBounds, tl, tr),
		join2horz(newBounds, bl, br),
	)
}
