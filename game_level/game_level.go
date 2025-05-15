package game_level

import pb "witwiz/proto"

type GameLevel interface {
	LevelId() int32
	WorldViewPort() *pb.ViewPort
	ComputeWorldOffsetX(currentWorldOffsetX, delta float32) float32
	Completed() bool
}
