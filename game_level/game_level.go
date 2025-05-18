package game_level

import pb "witwiz/proto"

type GameLevel interface {
	LevelId() int32
	LevelBounds() *pb.Bounds
	ViewPortBounds() *pb.Bounds
	UpdateViewPortBounds(deltaTime float32)
	Completed() bool
}

type viewPort struct {
	bounds    *pb.Bounds
	speed     *pb.Vector2
	paths     []*path
	pathIndex int
}

type path struct {
	scroll      scroll
	speed       float32
	direction   direction
	destination float32
}

type scroll int8

const (
	SCROLL_HORIZONTALLY scroll = 0
	SCROLL_VERTICALLY   scroll = 1
)

type direction int8

const (
	DIRECTION_POSITIVE = 1
	DIRECTION_NEGATIVE = -1
)
