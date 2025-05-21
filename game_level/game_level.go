package game_level

import pb "witwiz/proto"

const (
	defaultResolutionWidth  float32 = 1080
	defaultResolutionHeight float32 = 720
)

type GameLevel interface {
	LevelId() int32
	LevelSize() *pb.Size
	LevelPosition() *pb.Point
	LevelVelocity() *pb.Vector
	LevelObstacles() []*pb.ObstacleState
	UpdateLevelPosition(deltaTime float32)
	NextLevelPortal() *pb.NextLevelPortalState
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
