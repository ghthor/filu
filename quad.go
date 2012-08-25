package engine

type (
	AABB struct {
		TopL, BotR WorldCoord
	}
)

func (aabb AABB) Contains(c WorldCoord) bool {
	return (aabb.TopL.X <= c.X && aabb.BotR.X >= c.X &&
		aabb.TopL.Y >= c.Y && aabb.BotR.Y <= c.Y)
}

func (aabb AABB) Width() int {
	return aabb.BotR.X - aabb.TopL.X + 1
}

func (aabb AABB) Height() int {
	return aabb.TopL.Y - aabb.BotR.Y + 1
}

func (aabb AABB) Area() int {
	return (aabb.BotR.X - aabb.TopL.X + 1) * (aabb.TopL.Y - aabb.BotR.Y + 1)
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
