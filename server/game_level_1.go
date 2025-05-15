package server

import (
	"time"
	pb "witwiz/proto"
)

type gameLevel1 struct {
	viewPort            *pb.ViewPort
	worldScrollSpeed    float32
	bossEncountered     bool
	bossPositionX       float32
	bossHealth          float32
	bossHealthIsTicking bool
}

func newGameLevel1() *gameLevel1 {
	return &gameLevel1{
		viewPort:            &pb.ViewPort{Width: 5000, Height: 720},
		worldScrollSpeed:    100,
		bossEncountered:     false,
		bossPositionX:       3000,
		bossHealth:          10,
		bossHealthIsTicking: false,
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
		result = gl.bossPositionX
		gl.bossEncountered = true
		gl.worldScrollSpeed = 0
	}
	return result
}

func (gl *gameLevel1) worldViewPort() *pb.ViewPort {
	return gl.viewPort
}

func (gl *gameLevel1) completed() bool {
	if !gl.bossEncountered || gl.bossHealthIsTicking {
		return false
	}
	if gl.bossHealth <= 0 {
		return true
	}
	gl.bossHealthIsTicking = true
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			gl.bossHealth -= 2
			if gl.bossHealth <= 0 {
				break
			}
		}
		gl.bossHealthIsTicking = false
	}()
	return false
}
