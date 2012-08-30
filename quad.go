package engine

import (
	"errors"
)

type AABB struct {
	TopL, BotR WorldCoord
}

func (aabb AABB) Contains(c WorldCoord) bool {
	return (aabb.TopL.X <= c.X && aabb.BotR.X >= c.X &&
		aabb.TopL.Y >= c.Y && aabb.BotR.Y <= c.Y)
}

func (aabb AABB) HasOnEdge(c WorldCoord) (onEdge bool) {
	x, y := c.X, c.Y
	switch {
	case (x == aabb.TopL.X || x == aabb.BotR.X) && (y <= aabb.TopL.Y && y >= aabb.BotR.Y):
		fallthrough
	case (y == aabb.TopL.Y || y == aabb.BotR.Y) && (x >= aabb.TopL.X && x <= aabb.BotR.X):
		onEdge = true
	default:
	}
	return
}

func (aabb AABB) Width() int {
	return aabb.BotR.X - aabb.TopL.X + 1
}

func (aabb AABB) Height() int {
	return aabb.TopL.Y - aabb.BotR.Y + 1
}

func (aabb AABB) TopR() WorldCoord { return WorldCoord{aabb.BotR.X, aabb.TopL.Y} }
func (aabb AABB) BotL() WorldCoord { return WorldCoord{aabb.TopL.X, aabb.BotR.Y} }

func (aabb AABB) Area() int {
	return (aabb.BotR.X - aabb.TopL.X + 1) * (aabb.TopL.Y - aabb.BotR.Y + 1)
}

func (aabb AABB) Overlaps(other AABB) bool {
	if aabb.Contains(other.TopL) || aabb.Contains(other.BotR) ||
		other.Contains(aabb.TopL) || other.Contains(aabb.BotR) {
		return true
	}

	return aabb.Contains(other.TopR()) || aabb.Contains(other.BotL()) ||
		other.Contains(aabb.TopR()) || other.Contains(aabb.BotL())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (aabb AABB) Intersection(other AABB) (AABB, error) {
	if !aabb.Overlaps(other) {
		return AABB{}, errors.New("no overlap")
	}

	return AABB{
		WorldCoord{max(aabb.TopL.X, other.TopL.X), min(aabb.TopL.Y, other.TopL.Y)},
		WorldCoord{min(aabb.BotR.X, other.BotR.X), max(aabb.BotR.Y, other.BotR.Y)},
	}, nil
}

// Is BotR actually TopL?
func (aabb AABB) IsInverted() bool {
	return aabb.BotR.Y > aabb.TopL.Y && aabb.BotR.X < aabb.TopL.X
}

// Flip TopL and BotR
func (aabb AABB) Invert() AABB {
	return AABB{
		aabb.BotR, aabb.TopL,
	}
}

type (
	quad interface {
		Parent() quad
		AABB() AABB
		Insert(entity) quad
		InsertAll([]entity) quad
		Contains(entity) bool
		AdjustPositions(WorldTime) []movableEntity
	}

	quadLeaf struct {
		parent          quad
		aabb            AABB
		entities        []entity
		movableEntities []movableEntity
		maxEntities     int
	}

	quadTree struct {
		parent quad
		aabb   AABB
		quads  [4]quad
	}
)

const (
	QUAD_NW = iota
	QUAD_NE
	QUAD_SE
	QUAD_SW
)

func splitAABBToQuads(aabb AABB) ([4]AABB, error) {
	var aabbs [4]AABB

	if aabb.IsInverted() {
		return aabbs, errors.New("aabb is inverted")
	}

	w, h := aabb.Width(), aabb.Height()

	if w < 2 || h < 2 {
		return aabbs, errors.New("aabb is too small to split")
	}

	// NorthWest
	nw := AABB{
		aabb.TopL,
		WorldCoord{aabb.TopL.X + (w/2 - 1), aabb.TopL.Y - (h/2 - 1)},
	}

	// NorthEast
	ne := AABB{
		WorldCoord{nw.BotR.X + 1, aabb.TopL.Y},
		WorldCoord{aabb.BotR.X, nw.BotR.Y},
	}

	se := AABB{
		WorldCoord{ne.TopL.X, ne.BotR.Y - 1},
		aabb.BotR,
	}

	sw := AABB{
		WorldCoord{aabb.TopL.X, se.TopL.Y},
		WorldCoord{nw.BotR.X, aabb.BotR.Y},
	}

	aabbs[QUAD_NW] = nw
	aabbs[QUAD_NE] = ne
	aabbs[QUAD_SE] = se
	aabbs[QUAD_SW] = sw

	return aabbs, nil
}

func newQuadTree(aabb AABB, entities []entity, maxPerQuad int) (quad, error) {
	if aabb.IsInverted() {
		return nil, errors.New("aabb is Inverted")
	}

	if aabb.Area() <= 1 {
		return nil, errors.New("aabb Area is invalid")
	}

	if entities != nil {
		panic("unimplemented")
	}

	return &quadLeaf{
		nil,
		aabb,
		make([]entity, 0, maxPerQuad),
		make([]movableEntity, 0, maxPerQuad),
		maxPerQuad,
	}, nil
}

func (q *quadLeaf) Parent() quad { return q.parent }
func (q *quadLeaf) AABB() AABB   { return q.aabb }

func (q *quadLeaf) Insert(e entity) quad {
	// Check if this Quad is full
	if len(q.entities) == q.maxEntities {
		return q.divide().Insert(e)
	}

	q.entities = append(q.entities, e)

	// Collect Movable Entities
	if me, ok := e.(movableEntity); ok {
		q.movableEntities = append(q.movableEntities, me)
	}
	return q
}

func (q *quadLeaf) InsertAll(entities []entity) quad {
	// Check if Quad will overflow in size
	if len(q.entities)+len(entities) > q.maxEntities {
		return q.divide().InsertAll(entities)
	}

	q.entities = append(q.entities, entities...)

	// Collect Movable Entities
	for _, e := range entities {
		if me, ok := e.(movableEntity); ok {
			q.movableEntities = append(q.movableEntities, me)
		}
	}
	return q
}

func (q *quadLeaf) Contains(e entity) bool {
	for _, entity := range q.entities {
		if entity.Id() == e.Id() {
			return true
		}
	}
	return false
}

func (q *quadLeaf) AdjustPositions(t WorldTime) []movableEntity {
	// Worst Case sizing
	movedOutside := make([]movableEntity, 0, len(q.movableEntities))
	for _, e := range q.movableEntities {
		mi := e.motionInfo()

		// Removed finished pathActions
		for _, pa := range mi.pathActions {
			if pa.end <= t {
				mi.lastMoveAction = pa
				mi.pathActions = mi.pathActions[:0]
				mi.coord = pa.Dest

				if !q.aabb.Contains(mi.coord) {
					movedOutside = append(movedOutside, e)
				}
			}
		}
	}

	if q.parent == nil && len(movedOutside) > 0 {
		panic("entity was moved outside of the world's bounds")
	}
	return movedOutside
}

func (q *quadLeaf) divide() (qt *quadTree) {
	if q.aabb.Width() == 1 {
		panic("unable to divide quad with width of 1")
	}

	if q.aabb.Height() == 1 {
		panic("unable to divide quad with height of 1")
	}

	qt = &quadTree{
		parent: q.parent,
		aabb:   q.aabb,
	}

	aabbs, err := splitAABBToQuads(q.aabb)
	if err != nil {
		panic("error spliting aabb into quads")
	}

	//TODO Reuse this leaf forming 3 new leaves + this 1
	for i, _ := range qt.quads {
		qt.quads[i] = &quadLeaf{
			qt,
			aabbs[i],
			make([]entity, 0, cap(q.entities)),
			make([]movableEntity, 0, cap(q.entities)),
			q.maxEntities,
		}
	}

	qt.InsertAll(q.entities)

	return qt
}

func (q *quadTree) Parent() quad { return q.parent }
func (q *quadTree) AABB() AABB   { return q.aabb }

func (q *quadTree) Insert(e entity) quad {
	for i, quad := range q.quads {
		if quad.AABB().Contains(e.Coord()) {
			quad = quad.Insert(e)
			q.quads[i] = quad
			return q
		}
	}
	panic("no quads could contain entity")
}

func (q *quadTree) InsertAll(entities []entity) quad {
	for _, entity := range entities {
		q.Insert(entity)
	}
	return q
}

func (q *quadTree) Contains(e entity) bool {
	for _, quad := range q.quads {
		if quad.Contains(e) {
			return true
		}
	}
	return false
}

func (q *quadTree) AdjustPositions(t WorldTime) []movableEntity {
	changedQuad := make([]movableEntity, 0, 4)
	for _, quad := range q.quads {
		changedQuad = append(changedQuad, quad.AdjustPositions(t)...)
	}

	movedOutside := make([]movableEntity, 0, len(changedQuad))
	for _, e := range changedQuad {
		if q.aabb.Contains(e.Coord()) {
			// Safe to call Insert w/o assignment because quadTree never divides
			q.Insert(e)
		} else {
			movedOutside = append(movedOutside, e)
		}
	}

	if q.parent == nil && len(movedOutside) > 0 {
		panic("entity was moved outside of the world's bounds")
	}
	return movedOutside
}
