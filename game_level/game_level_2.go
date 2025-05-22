package game_level

import (
	pb "witwiz/proto"
)

type GameLevel2 struct {
	base *baseGameLevel
}

func NewGameLevel2() *GameLevel2 {
	base := newBaseGameLevel(2, &pb.Size{Width: 5120, Height: 1024})
	base.paths = append(base.paths,
		&path{scroll: SCROLL_HORIZONTALLY, speed: 200, direction: DIRECTION_NEGATIVE},
		&path{scroll: SCROLL_HORIZONTALLY, speed: 200, direction: DIRECTION_POSITIVE},
	)
	base.pathIndex = 0
	base.obstacles = append(base.obstacles,
		&pb.ObstacleState{
			Id:       1,
			Size:     &pb.Size{Width: 200, Height: 200},
			Position: &pb.Point{X: 2048, Y: 400},
		},
	)
	return &GameLevel2{base: base}
}

func (gl *GameLevel2) LevelId() int32 {
	return gl.base.LevelId()
}

func (gl *GameLevel2) LevelSize() *pb.Size {
	return gl.base.LevelSize()
}

func (gl *GameLevel2) LevelPosition() *pb.Point {
	return gl.base.LevelPosition()
}

func (gl *GameLevel2) UpdateLevelPosition(deltaTime float32) {
	gl.base.UpdateLevelPosition(deltaTime)
	if gl.base.pathIndex < len(gl.base.paths) ||
		gl.base.nextLevelPortal != nil {
		return
	}
	size := &pb.Size{Width: 100, Height: defaultResolutionHeight}
	gl.base.nextLevelPortal = &pb.NextLevelPortalState{
		Size: size,
		Position: &pb.Point{
			X: defaultResolutionWidth - size.Width/2,
			Y: defaultResolutionHeight - size.Height/2,
		},
	}
}

func (gl *GameLevel2) NextLevelPortal() *pb.NextLevelPortalState {
	return gl.base.NextLevelPortal()
}

func (gl *GameLevel2) LevelObstacles() []*pb.ObstacleState {
	return gl.base.LevelObstacles()
}

func (gl *GameLevel2) LevelVelocity() *pb.Vector {
	return gl.base.LevelVelocity()
}

func (gl *GameLevel2) LevelEdges() []*pb.LevelEdgeState {
	return gl.base.LevelEdges()
}
