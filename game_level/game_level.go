package game_level

import pb "witwiz/proto"

type GameLevel interface {
	LevelId() int32
	LevelSize() *pb.Size
	LevelBounds() *pb.Bounds
	ViewportSize() *pb.Size
	ViewportBounds() *pb.Bounds
	UpdateViewportBounds(deltaTime float32)
	NextLevelPortal() *pb.NextLevelPortalState
	LevelObstacles() []*pb.ObstacleState
}

type viewport struct {
	bounds    *pb.Bounds
	velocity  float32
	paths     []*path
	pathIndex int
}

type path struct {
	scroll    scroll
	speed     float32
	direction direction
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
