package engine

type (
	EntityId int64

	Entity interface {
		Id() EntityId
	}

	MoveRequest struct {
		t WorldTime
		Direction
	}

	MotionInfo struct {
		coord  WorldCoord
		facing Direction

		pathActions []*PathAction
	}

	MovableEntity interface {
		Entity
		MotionInfo() *MotionInfo
	}
)

func NewMotionInfo(c WorldCoord, f Direction) *MotionInfo {
	return &MotionInfo{
		c,
		f,
		make([]*PathAction, 0, 2),
	}
}

func (mi MotionInfo) IsMoving() bool {
	return len(mi.pathActions) == 0
}

type PlayerDef struct {
	Name      string
	Facing    Direction
	Coord     WorldCoord
	NewPlayer chan *Player
}

type Player struct {
	entityId   EntityId
	Name       string
	motionInfo *MotionInfo
}

func (p *Player) Id() EntityId {
	return p.entityId
}

func (p *Player) MotionInfo() *MotionInfo {
	return p.motionInfo
}
