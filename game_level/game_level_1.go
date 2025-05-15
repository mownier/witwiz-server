package game_level

import (
	"time"
	pb "witwiz/proto"
)

type GameLevel1 struct {
	viewPort            *pb.ViewPort
	worldScrollSpeed    float32
	bossEncountered     bool
	bossPositionX       float32
	bossHealth          float32
	bossHealthIsTicking bool
}

func NewGameLevel1() *GameLevel1 {
	return &GameLevel1{
		viewPort:            &pb.ViewPort{Width: 5000, Height: 720},
		worldScrollSpeed:    100,
		bossEncountered:     false,
		bossPositionX:       3000,
		bossHealth:          10,
		bossHealthIsTicking: false,
	}
}

func (gl *GameLevel1) LevelId() int32 {
	return 1
}

func (gl *GameLevel1) ComputeWorldOffsetX(currentWorldOffsetX, deltaTime float32) float32 {
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

func (gl *GameLevel1) WorldViewPort() *pb.ViewPort {
	return gl.viewPort
}

func (gl *GameLevel1) Completed() bool {
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
