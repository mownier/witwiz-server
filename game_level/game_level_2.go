package game_level

import (
	pb "witwiz/proto"
)

type GameLevel2 struct {
	levelId        int32
	viewPortBounds *pb.ViewPortBounds
	completed      bool
}

func NewGameLevel2() *GameLevel2 {
	return &GameLevel2{
		levelId:        1,
		viewPortBounds: &pb.ViewPortBounds{MinX: 0, MinY: 0, MaxX: 5120, MaxY: 1024},
		completed:      false,
	}
}

func (gl *GameLevel2) LevelId() int32 {
	return gl.levelId
}

func (gl *GameLevel2) ViewPortBounds() *pb.ViewPortBounds {
	return gl.viewPortBounds
}

func (gl *GameLevel2) Completed() bool {
	return gl.completed
}
