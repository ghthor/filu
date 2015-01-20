package engine

import (
	"errors"

	. "github.com/ghthor/engine/coord"
	. "github.com/ghthor/engine/time"
)

type AABB struct {
	TopL Cell `json:"tl"`
	BotR Cell `json:"br"`
}

func (aabb AABB) Contains(c Cell) bool {
	return (aabb.TopL.X <= c.X && aabb.BotR.X >= c.X &&
		aabb.TopL.Y >= c.Y && aabb.BotR.Y <= c.Y)
}

func (aabb AABB) HasOnEdge(c Cell) (onEdge bool) {
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

func (aabb AABB) TopR() Cell { return Cell{aabb.BotR.X, aabb.TopL.Y} }
func (aabb AABB) BotL() Cell { return Cell{aabb.TopL.X, aabb.BotR.Y} }

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
		Cell{max(aabb.TopL.X, other.TopL.X), min(aabb.TopL.Y, other.TopL.Y)},
		Cell{min(aabb.BotR.X, other.BotR.X), max(aabb.BotR.Y, other.BotR.Y)},
	}, nil
}

func (aabb AABB) Expand(mag int) AABB {
	aabb.TopL = aabb.TopL.Add(-mag, mag)
	aabb.BotR = aabb.BotR.Add(mag, -mag)
	return aabb
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
		Remove(entity)
		Contains(entity) bool
		QueryAll(AABB) []entity
		QueryCollidables(Cell) []collidableEntity

		StepTo(WorldTime)

		// Broad phase
		// 1. updatePosistions - serial
		updatePositions(WorldTime) []movableEntity

		// 2. stepTo           - concurrent
		stepTo(WorldTime, chan []movableEntity)
	}

	quadLeaf struct {
		parent      quad
		aabb        AABB
		entities    []entity
		movable     []movableEntity
		collidable  []collidableEntity
		maxEntities int
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
		Cell{aabb.TopL.X + (w/2 - 1), aabb.TopL.Y - (h/2 - 1)},
	}

	// NorthEast
	ne := AABB{
		Cell{nw.BotR.X + 1, aabb.TopL.Y},
		Cell{aabb.BotR.X, nw.BotR.Y},
	}

	se := AABB{
		Cell{ne.TopL.X, ne.BotR.Y - 1},
		aabb.BotR,
	}

	sw := AABB{
		Cell{aabb.TopL.X, se.TopL.Y},
		Cell{nw.BotR.X, aabb.BotR.Y},
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
		make([]collidableEntity, 0, maxPerQuad),
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

	// Index Movable Entities
	if me, ok := e.(movableEntity); ok {
		q.movable = append(q.movable, me)
	}

	// Index collidable Entities
	if ce, ok := e.(collidableEntity); ok {
		q.collidable = append(q.collidable, ce)
	}
	return q
}

func (q *quadLeaf) Remove(e entity) {
	// Remove entity
	for i, entity := range q.entities {
		if entity.Id() == e.Id() {
			q.entities = append(q.entities[:i], q.entities[i+1:]...)
		}
	}

	// Remove Movable Entity
	if _, ok := e.(movableEntity); ok {
		for i, me := range q.movable {
			if me.Id() == e.Id() {
				q.movable = append(q.movable[:i], q.movable[i+1:]...)
			}
		}
	}

	// Remove Collidable Entity
	if _, ok := e.(collidableEntity); ok {
		for i, ce := range q.collidable {
			if ce.Id() == e.Id() {
				q.collidable = append(q.collidable[:i], q.collidable[i+1:]...)
			}
		}
	}
}

func (q *quadLeaf) InsertAll(entities []entity) quad {
	// Check if Quad will overflow in size
	if len(q.entities)+len(entities) > q.maxEntities {
		return q.divide().InsertAll(entities)
	}

	q.entities = append(q.entities, entities...)

	for _, e := range entities {
		// Index Movable Entities
		if me, ok := e.(movableEntity); ok {
			q.movable = append(q.movable, me)
		}

		// Index Collidable Entities
		if ce, ok := e.(collidableEntity); ok {
			q.collidable = append(q.collidable, ce)
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

func (q *quadLeaf) QueryAll(aabb AABB) []entity {
	matches := make([]entity, 0, len(q.entities))
	for _, e := range q.entities {
		if aabb.Overlaps(e.AABB()) {
			matches = append(matches, e)
		}
	}
	return matches
}

func (q *quadLeaf) QueryCollidables(c Cell) []collidableEntity {
	matches := make([]collidableEntity, 0, len(q.collidable))
	for _, ce := range q.collidable {
		if ce.AABB().Contains(c) {
			matches = append(matches, ce)
		}
	}
	return matches
}

func (q *quadLeaf) updatePositions(t WorldTime) []movableEntity {
	// Worst Case sizing
	movedOutside := make([]movableEntity, 0, len(q.movable))
	for _, e := range q.movable {
		mi := e.motionInfo()

		// Removed finished pathActions
		for _, pa := range mi.pathActions {
			if pa.Span.End <= t {
				mi.lastMoveAction = pa
				mi.pathActions = mi.pathActions[:0]
				mi.cell = pa.Dest

				if !q.aabb.Contains(mi.cell) {
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

func stepBounded(q quad, t WorldTime) {
	unsolvable := make(chan []movableEntity)
	go q.stepTo(t, unsolvable)

	// These collisions were unsolvable in ANY of the leafs
	// because the collision's AABB wanders outside of
	// the AABB of the world.
	entities := <-unsolvable

	// Bounds check the entities on the edge of the world
	for _, e := range entities {
		mi := e.motionInfo()
		if !q.AABB().Contains(mi.pathActions[0].Dest) {
			mi.UndoLastApply()
		}
	}
}

func (q *quadLeaf) StepTo(t WorldTime) {
	if q.parent != nil {
		panic("StepTo called on child quadLeaf")
	}

	stepBounded(q, t)
}

func (q *quadLeaf) stepTo(t WorldTime, unsolvable chan []movableEntity) {

	// This loop filters out Actions that can't happen yet because of TurnAction Delays
	beganMoving := make([]movableEntity, 0, len(q.movable))
	for _, e := range q.movable {
		mi := e.motionInfo()

		// No movement Request
		if mi.moveRequest == nil {
			continue
		}

		// Can't accept movement Request becuase entity is already moving
		if mi.isMoving() {
			continue
		}

		dest := mi.cell.Neighbor(mi.moveRequest.Direction)
		direction := mi.cell.DirectionTo(dest)

		// If the last MoveAction was a PathAction that ended on this Step
		if pathAction, ok := mi.lastMoveAction.(*PathAction); (ok && pathAction.End() == t) || (mi.facing == direction) {
			pathAction = &PathAction{
				NewSpan(t, t+WorldTime(mi.speed)),
				mi.cell,
				dest,
			}

			if pathAction.CanHappenAfter(mi.lastMoveAction) {
				mi.Apply(pathAction)
				beganMoving = append(beganMoving, e)
			}
		} else if mi.facing != direction {
			turnAction := TurnAction{
				mi.facing, direction,
				t,
			}

			// Attempt to Turn Facing
			if turnAction.CanHappenAfter(mi.lastMoveAction) {
				mi.Apply(turnAction)
			}
		}
	}

	unsolvables := make([]movableEntity, 0, len(beganMoving))

	collisions := make([]entityCollision, 0, len(q.collidable))
	// Filters out collisions that are the same
	addCollision := func(c entityCollision) {
		for _, other := range collisions {
			if c.SameAs(other) {
				return
			}
		}

		collisions = append(collisions, c)
	}

	// Find and collect collisions and unsolvables
	for _, me := range beganMoving {
		pa := me.motionInfo().pathActions[0]

		// Find unsolvables
		if !q.aabb.Contains(pa.Dest) {
			unsolvables = append(unsolvables, me)

		} else {
			// Find and collect collisions
			if ce, canCollide := me.(collidableEntity); canCollide {
				for _, other := range q.collidable {
					if ce != other && other.AABB().Contains(pa.Dest) {
						// This isn't really an entityCollison, more like a potentional collison...
						addCollision(entityCollision{t, ce, other})
					}
				}
			}
		}
	}

	// Run collisions
	for _, c := range collisions {
		c.collide()
	}

	unsolvable <- unsolvables
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
			make([]collidableEntity, 0, cap(q.entities)),
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
		if quad.AABB().Contains(e.Cell()) {
			quad = quad.Insert(e)
			q.quads[i] = quad
			return q
		}
	}
	panic("no quads could contain entity")
}

func (q *quadTree) Remove(e entity) {
	for _, quad := range q.quads {
		if quad.AABB().Contains(e.Cell()) {
			quad.Remove(e)
			return
		}
	}
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

func (q *quadTree) QueryAll(aabb AABB) []entity {
	matches := make([]entity, 0, 10)
	for _, quad := range q.quads {
		if quad.AABB().Overlaps(aabb) {
			matches = append(matches, quad.QueryAll(aabb)...)
		}
	}
	return matches
}

func (q *quadTree) QueryCollidables(c Cell) []collidableEntity {
	matches := make([]collidableEntity, 0, 10)
	for _, quad := range q.quads {
		if quad.AABB().Expand(1).Contains(c) {
			matches = append(matches, quad.QueryCollidables(c)...)
		}
	}
	return matches
}

func (q *quadTree) updatePositions(t WorldTime) []movableEntity {
	changedQuad := make([]movableEntity, 0, 4)

	// Adjust positions of all MovableEntities stored in every quadLeaf
	for _, quad := range q.quads {
		changedQuad = append(changedQuad, quad.updatePositions(t)...)
	}

	// Insert any Entities into the new leaf there position is within
	movedOutside := make([]movableEntity, 0, len(changedQuad))
	for _, e := range changedQuad {
		if q.aabb.Contains(e.Cell()) {
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

// Broad phase
func (q *quadTree) StepTo(t WorldTime) {
	q.updatePositions(t)

	if q.parent != nil {
		panic("StepTo called on child quadTree")
	}

	stepBounded(q, t)
}

func (q *quadTree) stepTo(t WorldTime, unsolvable chan []movableEntity) {
	leftToSolve := make(chan []movableEntity, 4)

	for _, quad := range q.quads {
		go quad.stepTo(t, leftToSolve)
	}

	entities := make([]movableEntity, 0, 10)
	for i := 0; i < 4; i++ {
		unsolved := <-leftToSolve
		entities = append(entities, unsolved...)
	}

	unsolvables := make([]movableEntity, 0, len(entities))

	collisions := make([]entityCollision, 0, len(entities))
	// Filters out collisions that are the same
	addCollision := func(c entityCollision) {
		for _, other := range collisions {
			if c.SameAs(other) {
				return
			}
		}

		collisions = append(collisions, c)
	}

	// Find and collect collisions and unsolvables
	for _, me := range entities {
		pa := me.motionInfo().pathActions[0]

		// Find and collect collisions and unsolvables
		if !q.aabb.Contains(pa.Dest) {
			unsolvables = append(unsolvables, me)

		} else {
			// Find and collect collisions
			if ce, isCollidable := me.(collidableEntity); isCollidable {
				for _, other := range q.QueryCollidables(pa.Dest) {
					if ce != other {
						addCollision(entityCollision{t, ce, other})
					}
				}
			}
		}
	}

	// Run collisions
	for _, c := range collisions {
		c.collide()
	}

	unsolvable <- unsolvables
}
