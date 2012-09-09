package engine

import (
	"errors"
)

type (
	TerrainType string

	TerrainMap struct {
		Bounds AABB
		// y, x
		TerrainTypes [][]TerrainType
	}
)

const (
	TT_GRASS TerrainType = "G"
	TT_DIRT  TerrainType = "D"
	TT_ROCK  TerrainType = "R"
)

func NewTerrainMap(bounds AABB) TerrainMap {
	w, h := bounds.Width(), bounds.Height()
	tm := make([][]TerrainType, h)

	for row, _ := range tm {
		tm[row] = make([]TerrainType, w)
	}

	for _, row := range tm {
		for x, _ := range row {
			row[x] = TT_GRASS
		}
	}

	return TerrainMap{bounds, tm}
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

	bytes := make([]byte, 0, w*h+1)

	bytes = append(bytes, []byte("\n")...)
	for _, row := range m.TerrainTypes {
		for _, t := range row {
			bytes = append(bytes, []byte(t)...)
		}
		bytes = append(bytes, []byte("\n")...)
	}

	return string(bytes)
}

func (terrainMap TerrainMap) SaveToFile(filename string) {
}
