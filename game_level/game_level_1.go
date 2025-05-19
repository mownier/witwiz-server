package game_level

import (
	pb "witwiz/proto"
)

type GameLevel1 struct {
	levelId       int32
	levelBounds   *pb.Bounds
	levelSpeed    float32
	levelVelocity float32
	viewPort      *viewPort
	completed     bool
}

func NewGameLevel1(viewPortWidth, viewPortHeight float32) *GameLevel1 {
	return &GameLevel1{
		levelId:       1,
		levelBounds:   &pb.Bounds{MinX: 0, MinY: 0, MaxX: 5120, MaxY: 1024},
		levelSpeed:    200,
		levelVelocity: 0,
		viewPort: &viewPort{
			bounds:   &pb.Bounds{MinX: 0, MinY: 0, MaxX: viewPortWidth, MaxY: viewPortHeight},
			velocity: 0,
			paths: []*path{
				{scroll: SCROLL_VERTICALLY, speed: 20, direction: DIRECTION_POSITIVE},
				{scroll: SCROLL_HORIZONTALLY, speed: 20, direction: DIRECTION_POSITIVE},
				{scroll: SCROLL_VERTICALLY, speed: 20, direction: DIRECTION_NEGATIVE},
				{scroll: SCROLL_HORIZONTALLY, speed: 20, direction: DIRECTION_NEGATIVE},
			},
			pathIndex: 0,
		},
		completed: false,
	}
}

func (gl *GameLevel1) LevelId() int32 {
	return gl.levelId
}

func (gl *GameLevel1) LevelBounds() *pb.Bounds {
	return gl.levelBounds
}

func (gl *GameLevel1) ViewPortBounds() *pb.Bounds {
	return gl.viewPort.bounds
}

func (gl *GameLevel1) ViewPortPathSpeed() float32 {
	if gl.viewPort.pathIndex >= len(gl.viewPort.paths) {
		return 0
	}
	return gl.viewPort.paths[gl.viewPort.pathIndex].speed
}

func (gl *GameLevel1) UpdateViewPortBounds(deltaTime float32) {
	if len(gl.viewPort.paths) == 0 {
		return
	}

	if gl.viewPort.pathIndex >= len(gl.viewPort.paths) {
		// TODO: No path to process. What's next?
		gl.viewPort.velocity = 0
		gl.levelVelocity = 0
		return
	}

	path := gl.viewPort.paths[gl.viewPort.pathIndex]
	gl.viewPort.velocity = path.speed * float32(path.direction)
	gl.levelVelocity = gl.levelSpeed * float32(path.direction*DIRECTION_NEGATIVE)

	switch path.scroll {
	case SCROLL_HORIZONTALLY:
		if gl.viewPort.velocity > 0 { // Scrolling to the right
			if gl.viewPort.bounds.MaxX < gl.levelBounds.MaxX {
				gl.viewPort.bounds.MinX += gl.viewPort.velocity * deltaTime
				gl.viewPort.bounds.MaxX += gl.viewPort.velocity * deltaTime
				gl.levelBounds.MinX += gl.levelVelocity * deltaTime
				gl.levelBounds.MaxX += gl.levelVelocity * deltaTime
			} else {
				diff := gl.viewPort.bounds.MaxX - gl.levelBounds.MaxX
				gl.viewPort.bounds.MinX -= diff
				gl.viewPort.bounds.MaxX = gl.levelBounds.MaxX
				gl.viewPort.velocity = 0
				gl.levelVelocity = 0
				gl.viewPort.pathIndex += 1
			}
		} else if gl.viewPort.velocity < 0 { // Scrolling to the left
			if gl.viewPort.bounds.MinX > gl.levelBounds.MinX {
				gl.viewPort.bounds.MinX += gl.viewPort.velocity * deltaTime
				gl.viewPort.bounds.MaxX += gl.viewPort.velocity * deltaTime
				gl.levelBounds.MinX += gl.levelVelocity * deltaTime
				gl.levelBounds.MaxX += gl.levelVelocity * deltaTime
			} else {
				diff := gl.levelBounds.MinX - gl.viewPort.bounds.MinX
				gl.viewPort.bounds.MaxX += diff
				gl.viewPort.bounds.MinX = gl.levelBounds.MinX
				gl.viewPort.velocity = 0
				gl.levelVelocity = 0
				gl.viewPort.pathIndex += 1
			}
		}

	case SCROLL_VERTICALLY:
		if gl.viewPort.velocity > 0 { // Scrolling down
			if gl.viewPort.bounds.MaxY < gl.levelBounds.MaxY {
				gl.viewPort.bounds.MinY += gl.viewPort.velocity * deltaTime
				gl.viewPort.bounds.MaxY += gl.viewPort.velocity * deltaTime
				gl.levelBounds.MinY += gl.levelVelocity * deltaTime
				gl.levelBounds.MaxY += gl.levelVelocity * deltaTime
			} else {
				diff := gl.viewPort.bounds.MaxY - gl.levelBounds.MaxY
				gl.viewPort.bounds.MinY -= diff
				gl.viewPort.bounds.MaxY = gl.levelBounds.MaxY
				gl.viewPort.velocity = 0
				gl.levelVelocity = 0
				gl.viewPort.pathIndex += 1
			}
		} else if gl.viewPort.velocity < 0 { // Scrolling up
			if gl.viewPort.bounds.MinY > gl.levelBounds.MinY {
				gl.viewPort.bounds.MinY += gl.viewPort.velocity * deltaTime
				gl.viewPort.bounds.MaxY += gl.viewPort.velocity * deltaTime
				gl.levelBounds.MinY += gl.levelVelocity * deltaTime
				gl.levelBounds.MaxY += gl.levelVelocity * deltaTime
			} else {
				diff := gl.levelBounds.MinY - gl.viewPort.bounds.MinY
				gl.viewPort.bounds.MaxY += diff
				gl.viewPort.bounds.MinY = gl.levelBounds.MinY
				gl.viewPort.velocity = 0
				gl.levelVelocity = 0
				gl.viewPort.pathIndex += 1
			}
		}
	}
}

func (gl *GameLevel1) Completed() bool {
	return gl.completed
}
