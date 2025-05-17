package game_level

import pb "witwiz/proto"

type GameLevel interface {
	LevelId() int32
	ViewPortBounds() *pb.ViewPortBounds
	Completed() bool
}
