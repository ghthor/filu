package perlin2d

import (
	"fmt"
	"math"
	"math/rand"
)

type Static [][]float64
type Noise [][]float64

func PrettyPrint(arr [][]float64) (str string) {
	for _, row := range arr {
		for _, val := range row {
			str += fmt.Sprintf("%6.5f, ", val)
		}
		str += "\n"
	}
	return str
}

func NewStatic(w, h int, seed int64) Static {
	// y, x
	s := make([][]float64, h)
	for y, _ := range s {
		s[y] = make([]float64, w)
	}

	r := rand.New(rand.NewSource(seed))

	for _, row := range s {
		for x, _ := range row {
			row[x] = r.Float64()
		}
	}
	return Static(s)
}

func (s Static) SmoothNoise(x, y float64) (noise float64) {

	w, h := len(s[0]), len(s)
	floorX, floorY := int(math.Floor(x)), int(math.Floor(y))

	fractX := x - float64(floorX)
	fractY := y - float64(floorY)

	x1 := (floorX + w) % w
	y1 := (floorY + h) % h

	x2 := (x1 + w - 1) % w
	y2 := (y1 + h - 1) % h

	noise += fractX * fractY * s[y1][x1]
	noise += fractX * (1 - fractY) * s[y2][x1]
	noise += (1 - fractX) * fractY * s[y1][x2]
	noise += (1 - fractX) * (1 - fractY) * s[y2][x2]

	return
}

func (s Static) Turbulence(x, y int, freq, octaves float64) (turb float64) {
	xf, yf := float64(x)*freq, float64(y)*freq
	for o := octaves; o >= 1.0; o /= 2.0 {
		turb += s.SmoothNoise(xf/o, yf/o) * o
	}

	turb /= octaves
	return
}

func NewNoise(static Static, freq, octaves float64) Noise {
	w, h := len(static[0]), len(static)
	n := make([][]float64, h)

	for y, _ := range n {
		n[y] = make([]float64, w)
	}

	for y, row := range n {
		for x, _ := range row {
			row[x] = static.Turbulence(x, y, freq, octaves)
		}
	}

	return Noise(n)
}

func (n Noise) String() string { return PrettyPrint([][]float64(n)) }
