package game_level

import (
	pb "witwiz/proto"
)

type GameLevel1 struct {
	levelId        int32
	levelBounds    *pb.Bounds
	viewPortBounds *pb.Bounds
	completed      bool
	scrollSpeedX   float32
	scrollSpeedY   float32
}

func NewGameLevel1(viewPortWidth, viewPortHeight float32) *GameLevel1 {
	return &GameLevel1{
		levelId:        1,
		levelBounds:    &pb.Bounds{MinX: 0, MinY: 0, MaxX: 5120, MaxY: 1024},
		viewPortBounds: &pb.Bounds{MinX: 0, MinY: 0, MaxX: viewPortWidth, MaxY: viewPortHeight},
		completed:      false,
		scrollSpeedX:   20,
		scrollSpeedY:   20,
	}
}

func (gl *GameLevel1) LevelId() int32 {
	return gl.levelId
}

func (gl *GameLevel1) LevelBounds() *pb.Bounds {
	return gl.levelBounds
}

func (gl *GameLevel1) ViewPortBounds() *pb.Bounds {
	return gl.viewPortBounds
}

func (gl *GameLevel1) UpdateViewPortBounds2(deltaTime float32) {
}

func (gl *GameLevel1) UpdateViewPortBounds(deltaTime float32) {
	if gl.scrollSpeedX > 0 { // Scrolling to the right
		if gl.viewPortBounds.MaxX < gl.levelBounds.MaxX {
			gl.viewPortBounds.MinX += gl.scrollSpeedX * deltaTime
			gl.viewPortBounds.MaxX += gl.scrollSpeedX * deltaTime
			// Optional: Clamp to the exact boundary if needed
			if gl.viewPortBounds.MaxX > gl.levelBounds.MaxX {
				diff := gl.viewPortBounds.MaxX - gl.levelBounds.MaxX
				gl.viewPortBounds.MinX -= diff
				gl.viewPortBounds.MaxX = gl.levelBounds.MaxX
				gl.scrollSpeedX = 0 // Stop horizontal scrolling
			}
		} else {
			gl.scrollSpeedX = 0 // Stop horizontal scrolling once reached
		}
	} else if gl.scrollSpeedX < 0 { // Scrolling to the left (if implemented)
		if gl.viewPortBounds.MinX > gl.levelBounds.MinX {
			gl.viewPortBounds.MinX += gl.scrollSpeedX * deltaTime
			gl.viewPortBounds.MaxX += gl.scrollSpeedX * deltaTime
			// Optional: Clamp to the exact boundary
			if gl.viewPortBounds.MinX < gl.levelBounds.MinX {
				diff := gl.levelBounds.MinX - gl.viewPortBounds.MinX
				gl.viewPortBounds.MaxX += diff
				gl.viewPortBounds.MinX = gl.levelBounds.MinX
				gl.scrollSpeedX = 0 // Stop horizontal scrolling
			}
		} else {
			gl.scrollSpeedX = 0 // Stop horizontal scrolling once reached
		}
	}

	if gl.scrollSpeedY > 0 { // Scrolling down
		if gl.viewPortBounds.MaxY < gl.levelBounds.MaxY {
			gl.viewPortBounds.MinY += gl.scrollSpeedY * deltaTime
			gl.viewPortBounds.MaxY += gl.scrollSpeedY * deltaTime
			// Optional clamping
			if gl.viewPortBounds.MaxY > gl.levelBounds.MaxY {
				diff := gl.viewPortBounds.MaxY - gl.levelBounds.MaxY
				gl.viewPortBounds.MinY -= diff
				gl.viewPortBounds.MaxY = gl.levelBounds.MaxY
				gl.scrollSpeedY = 0
			}
		} else {
			gl.scrollSpeedY = 0
		}
	} else if gl.scrollSpeedY < 0 { // Scrolling up
		if gl.viewPortBounds.MinY > gl.levelBounds.MinY {
			gl.viewPortBounds.MinY += gl.scrollSpeedY * deltaTime
			gl.viewPortBounds.MaxY += gl.scrollSpeedY * deltaTime
			// Optional clamping
			if gl.viewPortBounds.MinY < gl.levelBounds.MinY {
				diff := gl.levelBounds.MinY - gl.viewPortBounds.MinY
				gl.viewPortBounds.MaxY += diff
				gl.viewPortBounds.MinY = gl.levelBounds.MinY
				gl.scrollSpeedY = 0
			}
		} else {
			gl.scrollSpeedY = 0
		}
	}
}

func (gl *GameLevel1) Completed() bool {
	return gl.completed
}
