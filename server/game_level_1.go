package server

import pb "witwiz/proto"

type gameLevel1 struct {
	viewPort         *pb.ViewPort
	worldOffset      *pb.Vector2
	worldScrollSpeed float32
	bossEncountered  bool
	bossPositionX    float32
}

func newGameLevel1() *gameLevel1 {
	return &gameLevel1{
		viewPort:         &pb.ViewPort{Width: 5000, Height: 640},
		worldOffset:      &pb.Vector2{X: 0, Y: 0},
		worldScrollSpeed: 100,
		bossEncountered:  false,
		bossPositionX:    3000,
	}
}

func (gl *gameLevel1) levelId() int32 {
	return 1
}

func (gl *gameLevel1) computeWorldOffsetX(currentWorldOffsetX float32, deltaTime float32) float32 {
	if gl.bossEncountered {
		return currentWorldOffsetX
	}
	result := currentWorldOffsetX
	result += gl.worldScrollSpeed * deltaTime
	if result > gl.bossPositionX {
		gl.bossEncountered = true
		gl.worldScrollSpeed = 0
	}
	return result
}

func (gl *gameLevel1) worldViewPort() *pb.ViewPort {
	return gl.viewPort
}
