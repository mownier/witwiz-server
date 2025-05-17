package game_level

import (
	pb "witwiz/proto"
)

type GameLevel1 struct {
	levelId        int32
	viewPortBounds *pb.ViewPortBounds
	completed      bool
}

func NewGameLevel1() *GameLevel1 {
	return &GameLevel1{
		levelId:        1,
		viewPortBounds: &pb.ViewPortBounds{MinX: 0, MinY: 0, MaxX: 5120, MaxY: 1024},
		completed:      false,
	}
}

func (gl *GameLevel1) LevelId() int32 {
	return gl.levelId
}

func (gl *GameLevel1) ViewPortBounds() *pb.ViewPortBounds {
	return gl.viewPortBounds
}

func (gl *GameLevel1) Completed() bool {
	return gl.completed
}
