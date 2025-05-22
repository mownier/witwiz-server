package game_level

import pb "witwiz/proto"

const (
	defaultResolutionWidth  float32 = 1080
	defaultResolutionHeight float32 = 720
	defaultLevelEdgeAlongX  float32 = 8
	defaultLevelEdgeAlongY  float32 = 8
)

type GameLevel interface {
	LevelId() int32
	LevelSize() *pb.Size
	LevelPosition() *pb.Point
	LevelVelocity() *pb.Vector
	LevelObstacles() []*pb.ObstacleState
	LevelEdges() []*pb.LevelEdgeState
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
	DIRECTION_POSITIVE direction = 1
	DIRECTION_NEGATIVE direction = -1
)

type LevelEdge int8

const (
	LEVEL_EDGE_LEFT   = 1
	LEVEL_EDGE_RIGHT  = 2
	LEVEL_EDGE_BOTTOM = 3
	LEVEL_EDGE_TOP    = 4
)

func defaultLevelEdges(levelSize *pb.Size) []*pb.LevelEdgeState {
	return []*pb.LevelEdgeState{
		{
			Id:       LEVEL_EDGE_LEFT,
			Size:     &pb.Size{Width: defaultLevelEdgeAlongX, Height: levelSize.Height},
			Position: &pb.Point{X: defaultLevelEdgeAlongX / 2, Y: levelSize.Height / 2},
		},
		{
			Id:       LEVEL_EDGE_RIGHT,
			Size:     &pb.Size{Width: defaultLevelEdgeAlongX, Height: levelSize.Height},
			Position: &pb.Point{X: defaultResolutionWidth - defaultLevelEdgeAlongX/2, Y: levelSize.Height / 2},
		},
		{
			Id:       LEVEL_EDGE_BOTTOM,
			Size:     &pb.Size{Width: levelSize.Width, Height: defaultLevelEdgeAlongY},
			Position: &pb.Point{X: levelSize.Width / 2, Y: defaultLevelEdgeAlongY / 2},
		},
		{
			Id:       LEVEL_EDGE_TOP,
			Size:     &pb.Size{Width: levelSize.Width, Height: defaultLevelEdgeAlongY},
			Position: &pb.Point{X: levelSize.Width / 2, Y: defaultResolutionHeight - defaultLevelEdgeAlongY/2},
		},
	}
}
