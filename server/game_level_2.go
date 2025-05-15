package server

import (
	pb "witwiz/proto"
)

type gameLevel2 struct {
	viewPort         *pb.ViewPort
	worldScrollSpeed float32
	bossEncountered  bool
	bossPositionX    float32
}

func newGameLevel2() *gameLevel2 {
	return &gameLevel2{
		viewPort:         &pb.ViewPort{Width: 6000, Height: 720},
		worldScrollSpeed: 200,
		bossEncountered:  false,
		bossPositionX:    4000,
	}
}

func (gl *gameLevel2) levelId() int32 {
	return 2
}

func (gl *gameLevel2) computeWorldOffsetX(currentWorldOffsetX float32, deltaTime float32) float32 {
	if gl.bossEncountered {
		return currentWorldOffsetX
	}
	result := currentWorldOffsetX
	result += gl.worldScrollSpeed * deltaTime
	if result > gl.bossPositionX {
		result = gl.bossPositionX
		gl.bossEncountered = true
		gl.worldScrollSpeed = 0
	}
	return result
}

func (gl *gameLevel2) worldViewPort() *pb.ViewPort {
	return gl.viewPort
}

func (gl *gameLevel2) completed() bool {
	return false
}
