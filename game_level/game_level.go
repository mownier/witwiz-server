package game_level

import pb "witwiz/proto"

type GameLevel interface {
	LevelId() int32
	LevelBounds() *pb.Bounds
	ViewPortBounds() *pb.Bounds
	UpdateViewPortBounds(deltaTime float32)
	Completed() bool
}
