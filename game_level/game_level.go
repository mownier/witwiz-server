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

type baseGameLevel struct {
	levelId         int32
	levelSize       *pb.Size
	levelPosition   *pb.Point
	levelVelocity   *pb.Vector
	levelEdges      []*pb.LevelEdgeState
	nextLevelPortal *pb.NextLevelPortalState
	obstacles       []*pb.ObstacleState
	paths           []*path
	pathIndex       int
}

func newBaseGameLevel(levelId int32, levelSize *pb.Size) *baseGameLevel {
	return &baseGameLevel{
		levelId:         levelId,
		levelSize:       levelSize,
		levelPosition:   &pb.Point{X: 0, Y: 0},
		levelVelocity:   &pb.Vector{X: 0, Y: 0},
		obstacles:       []*pb.ObstacleState{},
		nextLevelPortal: nil,
		paths:           []*path{},
		pathIndex:       -1,
		levelEdges:      defaultLevelEdges(levelSize),
	}
}

func (gl *baseGameLevel) LevelId() int32 {
	return gl.levelId
}

func (gl *baseGameLevel) LevelSize() *pb.Size {
	return gl.levelSize
}

func (gl *baseGameLevel) LevelPosition() *pb.Point {
	return gl.levelPosition
}

func (gl *baseGameLevel) LevelVelocity() *pb.Vector {
	return gl.levelVelocity
}

func (gl *baseGameLevel) LevelObstacles() []*pb.ObstacleState {
	return gl.obstacles
}

func (gl *baseGameLevel) LevelEdges() []*pb.LevelEdgeState {
	return gl.levelEdges
}

func (gl *baseGameLevel) UpdateLevelPosition(deltaTime float32) {
	if len(gl.paths) == 0 {
		return
	}

	if gl.pathIndex >= len(gl.paths) {
		gl.levelVelocity.X = 0
		gl.levelVelocity.Y = 0
		return
	}

	path := gl.paths[gl.pathIndex]

	switch path.scroll {
	case SCROLL_HORIZONTALLY:
		gl.levelVelocity.Y = 0
		gl.levelVelocity.X = path.speed * float32(path.direction)
		gl.levelPosition.X += gl.levelVelocity.X * deltaTime
		// Bounds check
		if gl.levelPosition.X < -gl.levelSize.Width+defaultResolutionWidth {
			gl.levelPosition.X = -gl.levelSize.Width + defaultResolutionWidth
			gl.levelVelocity.X = 0
			gl.pathIndex += 1
		} else if gl.levelPosition.X > 0 {
			gl.levelPosition.X = 0
			gl.levelVelocity.X = 0
			gl.pathIndex += 1
		}

	case SCROLL_VERTICALLY:
		gl.levelVelocity.X = 0
		gl.levelVelocity.Y = path.speed * float32(path.direction)
		gl.levelPosition.Y += gl.levelVelocity.Y * deltaTime
		// Bounds check
		if gl.levelPosition.Y < -gl.levelSize.Height+defaultResolutionHeight {
			gl.levelPosition.Y = -gl.levelSize.Height + defaultResolutionHeight
			gl.levelVelocity.Y = 0
			gl.pathIndex += 1
		} else if gl.levelPosition.Y > 0 {
			gl.levelPosition.Y = 0
			gl.levelVelocity.Y = 0
			gl.pathIndex += 1
		}
	}

	for _, edge := range gl.levelEdges {
		if edge.Id == 1 || edge.Id == 2 {
			edge.Position.X += gl.levelVelocity.X * -1 * deltaTime
		} else if edge.Id == 3 || edge.Id == 4 {
			edge.Position.Y += gl.levelVelocity.Y * -1 * deltaTime
		}
	}
}

func (gl *baseGameLevel) NextLevelPortal() *pb.NextLevelPortalState {
	return gl.nextLevelPortal
}
