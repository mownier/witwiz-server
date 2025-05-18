package game_level

import (
	"math"
	pb "witwiz/proto"
)

type GameLevel2 struct {
	levelId        int32
	levelBounds    *pb.Bounds
	viewPortBounds *pb.Bounds
	completed      bool
	scrollSpeedX   float32
	scrollSpeedY   float32
}

func NewGameLevel2(viewPortWidth, viewPortHeight float32) *GameLevel2 {
	return &GameLevel2{
		levelId:        1,
		levelBounds:    &pb.Bounds{MinX: 0, MinY: 0, MaxX: 5120, MaxY: 1024},
		viewPortBounds: &pb.Bounds{MinX: 0, MinY: 0, MaxX: viewPortWidth, MaxY: viewPortHeight},
		completed:      false,
		scrollSpeedX:   20,
		scrollSpeedY:   0,
	}
}

func (gl *GameLevel2) LevelId() int32 {
	return gl.levelId
}

func (gl *GameLevel2) LevelBounds() *pb.Bounds {
	return gl.levelBounds
}

func (gl *GameLevel2) ViewPortBounds() *pb.Bounds {
	return gl.viewPortBounds
}

func (gl *GameLevel2) UpdateViewPortBounds(deltaTime float32) {
	gl.viewPortBounds.MinX += gl.scrollSpeedX * deltaTime
	gl.viewPortBounds.MaxX += gl.scrollSpeedX * deltaTime
	gl.viewPortBounds.MinY += gl.scrollSpeedY * deltaTime
	gl.viewPortBounds.MaxY += gl.scrollSpeedY * deltaTime

	gl.viewPortBounds.MinX = float32(math.Max(float64(gl.viewPortBounds.MinX), float64(gl.levelBounds.MinX)))
	gl.viewPortBounds.MaxX = float32(math.Min(float64(gl.viewPortBounds.MaxX), float64(gl.levelBounds.MaxX)))
	gl.viewPortBounds.MinY = float32(math.Max(float64(gl.viewPortBounds.MinY), float64(gl.levelBounds.MinY)))
	gl.viewPortBounds.MaxY = float32(math.Min(float64(gl.viewPortBounds.MaxY), float64(gl.levelBounds.MaxY)))
}

func (gl *GameLevel2) Completed() bool {
	return gl.completed
}
