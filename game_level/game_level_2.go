package game_level

import (
	pb "witwiz/proto"
)

type GameLevel2 struct {
	levelId         int32
	levelSize       *pb.Size
	levelPosition   *pb.Point
	levelVelocity   *pb.Vector
	nextLevelPortal *pb.NextLevelPortalState
	obstacles       []*pb.ObstacleState
	paths           []*path
	pathIndex       int
}

func NewGameLevel2() *GameLevel2 {
	return &GameLevel2{
		levelId:         1,
		levelSize:       &pb.Size{Width: 5120, Height: 1024},
		levelPosition:   &pb.Point{X: 0, Y: 0},
		levelVelocity:   &pb.Vector{X: 0, Y: 0},
		nextLevelPortal: nil,
		paths: []*path{
			// {scroll: SCROLL_VERTICALLY, speed: 200, direction: DIRECTION_NEGATIVE},
			{scroll: SCROLL_HORIZONTALLY, speed: 200, direction: DIRECTION_NEGATIVE},
			// {scroll: SCROLL_VERTICALLY, speed: 200, direction: DIRECTION_POSITIVE},
			{scroll: SCROLL_HORIZONTALLY, speed: 200, direction: DIRECTION_POSITIVE},
		},
		pathIndex: 0,
		obstacles: []*pb.ObstacleState{
			{
				Id:       1,
				Size:     &pb.Size{Width: 200, Height: 200},
				Position: &pb.Point{X: 2048, Y: 400},
			},
		},
	}
}

func (gl *GameLevel2) LevelId() int32 {
	return gl.levelId
}

func (gl *GameLevel2) LevelSize() *pb.Size {
	return gl.levelSize
}

func (gl *GameLevel2) LevelPosition() *pb.Point {
	return gl.levelPosition
}

func (gl *GameLevel2) UpdateLevelPosition(deltaTime float32) {
	if len(gl.paths) == 0 {
		return
	}

	if gl.pathIndex >= len(gl.paths) {
		gl.levelVelocity.X = 0
		gl.levelVelocity.Y = 0
		if gl.nextLevelPortal == nil {
			size := &pb.Size{Width: 100, Height: 720}
			gl.nextLevelPortal = &pb.NextLevelPortalState{
				Size: size,
				Position: &pb.Point{
					X: 1080 - size.Width/2,
					Y: 720 - size.Height/2,
				},
			}
		}
		return
	}

	path := gl.paths[gl.pathIndex]

	switch path.scroll {
	case SCROLL_HORIZONTALLY:
		gl.levelVelocity.Y = 0
		gl.levelVelocity.X = path.speed * float32(path.direction)
		gl.levelPosition.X += gl.levelVelocity.X * deltaTime
		// Bounds check
		if gl.levelPosition.X < -gl.levelSize.Width {
			gl.levelPosition.X = -gl.levelSize.Width
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
		if gl.levelPosition.Y < -gl.levelSize.Height {
			gl.levelPosition.Y = -gl.levelSize.Height
			gl.levelVelocity.Y = 0
			gl.pathIndex += 1
		} else if gl.levelPosition.Y > 0 {
			gl.levelPosition.Y = 0
			gl.levelVelocity.Y = 0
			gl.pathIndex += 1
		}
	}
}

func (gl *GameLevel2) NextLevelPortal() *pb.NextLevelPortalState {
	return gl.nextLevelPortal
}

func (gl *GameLevel2) LevelObstacles() []*pb.ObstacleState {
	return gl.obstacles
}

func (gl *GameLevel2) LevelVelocity() *pb.Vector {
	return gl.levelVelocity
}
