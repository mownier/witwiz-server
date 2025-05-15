package game_level

import (
	"time"
	pb "witwiz/proto"
)

type GameLevel2 struct {
	viewPort            *pb.ViewPort
	worldScrollSpeed    float32
	bossEncountered     bool
	bossPositionX       float32
	bossHealth          float32
	bossHealthIsTicking bool
}

func NewGameLevel2() *GameLevel2 {
	return &GameLevel2{
		viewPort:            &pb.ViewPort{Width: 6000, Height: 720},
		worldScrollSpeed:    200,
		bossEncountered:     false,
		bossPositionX:       4000,
		bossHealth:          10,
		bossHealthIsTicking: false,
	}
}

func (gl *GameLevel2) LevelId() int32 {
	return 2
}

func (gl *GameLevel2) ComputeWorldOffsetX(currentWorldOffsetX, deltaTime float32) float32 {
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

func (gl *GameLevel2) WorldViewPort() *pb.ViewPort {
	return gl.viewPort
}

func (gl *GameLevel2) Completed() bool {
	if !gl.bossEncountered || gl.bossHealthIsTicking {
		return false
	}
	if gl.bossHealth <= 0 {
		return true
	}
	gl.bossHealthIsTicking = true
	go func() {
		ticker := time.NewTicker(1 * time.Second)
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
