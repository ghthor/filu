package quadstate

import (
	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
)

type Sector int

/*
12223
80004
80004
80004
76665
*/

//go:generate stringer -type=Sector -output=sector_type_string.go
const (
	C Sector = iota
	N
	NE
	E
	SE
	S
	SW
	W
	NW
	SectorSize
)

type entityIndex struct {
	Sector
	Type
}

type Viewport struct {
	Bounds coord.Bounds

	Sector [SectorSize]*Entities

	index map[entity.Id]entityIndex
	// Used to store entities during shifts in the cardinal directions
	OutOfBounds map[entity.Id]struct{}

	arrDeletes []int

	HasChanged bool
	IsNew      bool
}

func NewViewport(bounds coord.Bounds) *Viewport {
	p := &Viewport{
		Bounds:      bounds,
		index:       make(map[entity.Id]entityIndex),
		OutOfBounds: make(map[entity.Id]struct{}),
		HasChanged:  true,
		IsNew:       true,
	}

	p.Sector[NW] = NewEntities(1)
	p.Sector[NE] = NewEntities(1)
	p.Sector[SE] = NewEntities(1)
	p.Sector[SW] = NewEntities(1)

	p.Sector[N] = NewEntities(bounds.Width())
	p.Sector[E] = NewEntities(bounds.Height())
	p.Sector[S] = NewEntities(bounds.Width())
	p.Sector[W] = NewEntities(bounds.Height())
	p.Sector[C] = NewEntities((bounds.Width() - 2) * (bounds.Height() - 2))

	return p
}

// Needs to be called during Update Phase IFF Shift() has not been called
func (p *Viewport) ClearTypeInstant() {
	for sector := range p.Sector {
		p.Sector[sector].ByType[TypeInstant] =
			p.Sector[sector].ByType[TypeInstant][:0]
	}
}

// Needs to be called after worldstate.Update has been shipped
func (p *Viewport) ClearOOB() {
	for k := range p.OutOfBounds {
		delete(p.OutOfBounds, k)
	}
}

func (p *Viewport) AccumulateAll(acc Accumulator) {
	for sector := range p.Sector {
		for t := range p.Sector[sector].ByType {
			// if len(p.Sector[sector].ByType[t]) > 0 {
			// 	log.Printf("Sector(%v) and Type(%v)\n%#v", sector, t, p.Sector[sector].ByType[t])
			// }

			acc.AddSlice(p.Sector[sector].ByType[t], Type(t))
		}
	}
}

func (p *Viewport) AccumulateForDiff(acc Accumulator) {
	for sector := range p.Sector {
		for t := range p.Sector[sector].ByType {
			if sector == int(C) && t == int(TypeUnchanged) {
				continue
			}

			if t == int(TypeUnchanged) && !p.HasChanged {
				continue
			}

			acc.AddSlice(p.Sector[sector].ByType[t], Type(t))
		}
	}
}

func (p *Viewport) Insert(e *Entity) {
	// Make sure we replace an existing reference
	// TODO can optimize here if the entities cell hasn't changed
	if index, exists := p.index[e.Id]; exists {
		p.deleteEntityId(e.Id, index)
	}

	// Remove from the OutOfBounds index
	delete(p.OutOfBounds, e.Id)

	p.insertEntity(e)
}

func (p *Viewport) insertEntity(e *Entity) {
	switch e.Cell.X {
	case p.Bounds.TopL.X:
		p.insertWest(e)
		return
	case p.Bounds.BotR.X:
		p.insertEast(e)
		return
	}

	switch e.Cell.Y {
	case p.Bounds.TopL.Y:
		p.insertNorth(e)
		return
	case p.Bounds.BotR.Y:
		p.insertSouth(e)
		return
	}

	p.insert(C, e)
}

func (p *Viewport) insertWest(e *Entity) {
	switch e.Cell.Y {
	case p.Bounds.TopL.Y:
		p.insert(NW, e)
	case p.Bounds.BotR.Y:
		p.insert(SW, e)
	default:
		p.insert(W, e)
	}
}

func (p *Viewport) insertEast(e *Entity) {
	switch e.Cell.Y {
	case p.Bounds.TopL.Y:
		p.insert(NE, e)
	case p.Bounds.BotR.Y:
		p.insert(SE, e)
	default:
		p.insert(E, e)
	}
}

func (p *Viewport) insertNorth(e *Entity) {
	switch e.Cell.X {
	case p.Bounds.TopL.X:
		p.insert(NW, e)
	case p.Bounds.BotR.X:
		p.insert(NE, e)
	default:
		p.insert(N, e)
	}
}

func (p *Viewport) insertSouth(e *Entity) {
	switch e.Cell.X {
	case p.Bounds.TopL.X:
		p.insert(SW, e)
	case p.Bounds.BotR.X:
		p.insert(SE, e)
	default:
		p.insert(S, e)
	}
}

func (p *Viewport) insert(s Sector, e *Entity) {
	switch e.Type {
	case TypeInstant:
		p.Sector[s].Add(e)
	default:
		p.Sector[s].Add(e)
		p.index[e.Id] = entityIndex{s, e.Type}
		//log.Printf("Inserting %d to Sector(%v) and Type(%v)", e.Id, s, e.Type)
	}
}

func (p *Viewport) deleteEntityId(id entity.Id, index entityIndex) {
	a := p.Sector[index.Sector].ByType[index.Type]
	for i, e := range a {
		if e.Id == id {
			//log.Printf("Deleting %d from Sector(%v) and Type(%v)", id, index.Sector, index.Type)
			p.Sector[index.Sector].ByType[index.Type] =
				append(a[:i], a[i+1:]...)
			return
		}
	}
}

// Needs to be called during Update phase
func (p *Viewport) Shift(d coord.Direction) {
	b := p.Bounds
	p.Bounds = coord.Bounds{
		TopL: b.TopL.Neighbor(d),
		BotR: b.BotR.Neighbor(d),
	}
	p.HasChanged = true

	switch d {
	case coord.N:
		p.shiftNorth()
	case coord.E:
		p.shiftEast()
	case coord.S:
		p.shiftSouth()
	case coord.W:
		p.shiftWest()
	}

}

type dropSectorCmd [3]Sector

var dropNorth = dropSectorCmd{NW, N, NE}
var dropEast = dropSectorCmd{NE, E, SE}
var dropSouth = dropSectorCmd{SE, S, SW}
var dropWest = dropSectorCmd{SW, W, NW}

type sectorShift struct {
	from, to Sector
}

type complexShiftCmd struct {
	sectors simpleShiftCmd
	coord.Direction
}
type simpleShiftCmd [3]sectorShift

var complexShiftNorth = complexShiftCmd{
	simpleShiftCmd{{W, SW}, {C, S}, {E, SE}},
	coord.North,
}

var simpleShiftNorth = simpleShiftCmd{
	{NW, W}, {N, C}, {NE, E},
}

var complexShiftEast = complexShiftCmd{
	simpleShiftCmd{
		{N, NW},
		{C, W},
		{S, SW},
	},
	coord.East,
}

var simpleShiftEast = simpleShiftCmd{
	{NE, N},
	{E, C},
	{SE, S},
}

var complexShiftSouth = complexShiftCmd{
	simpleShiftCmd{
		{W, NW}, {C, N}, {E, NE},
	},
	coord.South,
}

var simpleShiftSouth = simpleShiftCmd{
	{SW, W}, {S, C}, {SE, E},
}

var complexShiftWest = complexShiftCmd{
	simpleShiftCmd{
		{N, NE},
		{C, E},
		{S, SE},
	},
	coord.West,
}

var simpleShiftWest = simpleShiftCmd{
	{NW, N},
	{W, C},
	{SW, S},
}

func (p *Viewport) shiftNorth() {
	p.shift(dropSouth, complexShiftNorth, simpleShiftNorth)
}

func (p *Viewport) shiftEast() {
	p.shift(dropWest, complexShiftEast, simpleShiftEast)
}

func (p *Viewport) shiftSouth() {
	p.shift(dropNorth, complexShiftSouth, simpleShiftSouth)
}

func (p *Viewport) shiftWest() {
	p.shift(dropEast, complexShiftWest, simpleShiftWest)
}

func (p *Viewport) shift(drop dropSectorCmd, complexShift complexShiftCmd, simpleShift simpleShiftCmd) {
	var sector *Entities
	for _, s := range drop {
		sector = p.Sector[s]
		for t := range sector.ByType {
			switch t {
			case int(TypeInstant):
			default:
				for _, e := range sector.ByType[t] {
					p.setOOB(e)
				}
			}
			sector.ByType[t] = sector.ByType[t][:0]
		}
	}

	bounds := p.Bounds
	var onEdge func(c coord.Cell) bool

	switch complexShift.Direction {
	case coord.North:
		onEdge = func(c coord.Cell) bool { return bounds.BotR.Y == c.Y }
	case coord.East:
		onEdge = func(c coord.Cell) bool { return bounds.TopL.X == c.X }
	case coord.South:
		onEdge = func(c coord.Cell) bool { return bounds.TopL.Y == c.Y }
	case coord.West:
		onEdge = func(c coord.Cell) bool { return bounds.BotR.X == c.X }
	}

	markDelete := p.arrDeletes

	for _, shift := range complexShift.sectors {
		sector = p.Sector[shift.from]
		for t := range sector.ByType {
			switch t {
			case int(TypeInstant):
				sector.ByType[TypeInstant] = sector.ByType[TypeInstant][:0]
			default:
				markDelete = markDelete[:0]
				for i, e := range sector.ByType[t] {
					if !onEdge(e.Cell) {
						continue
					}

					markDelete = append(markDelete, i)
					p.insert(shift.to, e)
				}
				for j := len(markDelete) - 1; j >= 0; j-- {
					i := markDelete[j]
					sector.ByType[t] = append(sector.ByType[t][:i], sector.ByType[t][i+1:]...)
				}
			}
		}
	}

	p.arrDeletes = markDelete

	for _, shift := range simpleShift {
		sector = p.Sector[shift.from]
		for t := range sector.ByType {
			switch t {
			case int(TypeInstant):
			default:
				for _, e := range sector.ByType[t] {
					p.insert(shift.to, e)
				}
			}
			sector.ByType[t] = sector.ByType[t][:0]
		}
	}
}

func (p *Viewport) setOOB(e *Entity) {
	delete(p.index, e.Id)
	p.OutOfBounds[e.Id] = struct{}{}
}
