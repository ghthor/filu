package engine

type (
	EntityId int64

	entity interface {
		Id() EntityId
		Cell() Cell
		AABB() AABB
		Json() EntityJson
	}

	EntityJson interface {
		Id() EntityId
		IsDifferentFrom(EntityJson) bool
	}

	moveRequest struct {
		t WorldTime
		Direction
	}

	motionInfo struct {
		cell   Cell
		facing Direction
		speed  uint

		moveRequest *moveRequest

		// fifo
		pathActions []*PathAction

		lastMoveAction MoveAction
		UndoLastApply  func()
	}

	movableEntity interface {
		entity
		motionInfo() *motionInfo
	}

	collidableEntity interface {
		entity
		collides(collidableEntity) bool
		collideWith(collidableEntity, WorldTime)
	}

	entityCollision struct {
		time WorldTime
		A, B collidableEntity
	}
)

func newMotionInfo(cell Cell, facing Direction, speed uint) *motionInfo {
	return &motionInfo{
		cell,
		facing,
		speed,
		nil,
		make([]*PathAction, 0, 2),
		nil,
		nil,
	}
}

func (mi motionInfo) isMoving() bool {
	return len(mi.pathActions) != 0
}

func (mi motionInfo) AABB() (aabb AABB) {
	if mi.isMoving() {
		pa := mi.pathActions[0]
		switch pa.Direction() {
		case West:
			fallthrough
		case North:
			aabb.TopL = pa.Dest
			aabb.BotR = pa.Orig
		case East:
			fallthrough
		case South:
			aabb.TopL = pa.Orig
			aabb.BotR = pa.Dest
		}
	} else {
		aabb = AABB{mi.cell, mi.cell}
	}
	return
}

func (mi *motionInfo) Apply(moveAction MoveAction) {
	switch action := moveAction.(type) {
	case TurnAction:
		mi.UndoLastApply = nil
		mi.facing = action.to
		mi.lastMoveAction = action

	case *PathAction:
		prevFacing := mi.facing
		prevMoveRequest := mi.moveRequest
		mi.UndoLastApply = func() {
			mi.UndoLastApply = nil
			mi.facing = prevFacing
			mi.pathActions = mi.pathActions[:0]
			mi.moveRequest = prevMoveRequest
		}

		mi.facing = action.Direction()
		mi.pathActions = append(mi.pathActions, action)

	default:
		panic("unknown MoveRequest type")
	}

	mi.moveRequest = nil
}

func (c entityCollision) SameAs(other entityCollision) (same bool) {
	if c.time != other.time {
		return false
	}

	switch {
	case c.A == other.A && c.B == other.B:
		fallthrough
	case c.A == other.B && c.B == other.A:
		same = true
	}
	return
}

func (c entityCollision) collide() {
	c.A.collideWith(c.B, c.time)
	c.B.collideWith(c.A, c.time)
}

type InputCmd struct {
	timeIssued WorldTime
	cmd        string
	params     string
}

func newMoveRequest(input InputCmd) *moveRequest {
	switch input.params {
	case "north":
		return &moveRequest{
			t:         input.timeIssued,
			Direction: North,
		}
	case "east":
		return &moveRequest{
			t:         input.timeIssued,
			Direction: East,
		}
	case "south":
		return &moveRequest{
			t:         input.timeIssued,
			Direction: South,
		}
	case "west":
		return &moveRequest{
			t:         input.timeIssued,
			Direction: West,
		}

	}
	panic("never reached")
}
