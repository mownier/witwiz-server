package game_level

import (
	pb "witwiz/proto"
)

type GameLevel1 struct {
	levelId     int32
	levelBounds *pb.Bounds
	viewPort    *viewPort
	completed   bool
}

func NewGameLevel1(viewPortWidth, viewPortHeight float32) *GameLevel1 {
	return &GameLevel1{
		levelId:     1,
		levelBounds: &pb.Bounds{MinX: 0, MinY: 0, MaxX: 5120, MaxY: 1024},
		viewPort: &viewPort{
			bounds: &pb.Bounds{MinX: 0, MinY: 0, MaxX: viewPortWidth, MaxY: viewPortHeight},
			speed:  &pb.Vector2{X: 0, Y: 0},
			paths: []*path{
				{scroll: SCROLL_VERTICALLY, destination: 1024, speed: 20, direction: DIRECTION_POSITIVE},
				{scroll: SCROLL_HORIZONTALLY, destination: 5120, speed: 20, direction: DIRECTION_POSITIVE},
				{scroll: SCROLL_VERTICALLY, destination: 720, speed: 20, direction: DIRECTION_NEGATIVE},
				{scroll: SCROLL_HORIZONTALLY, destination: 1080, speed: 20, direction: DIRECTION_NEGATIVE},
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
		gl.viewPort.speed.X = 0
		gl.viewPort.speed.Y = 0
		// TODO: No path to process. What's next?
		return
	}

	if gl.viewPort.pathIndex < len(gl.viewPort.paths) && gl.viewPort.paths[gl.viewPort.pathIndex].scroll == SCROLL_HORIZONTALLY {
		currentPathIndex := gl.viewPort.pathIndex
		path := gl.viewPort.paths[currentPathIndex]
		gl.viewPort.speed.X = path.speed * float32(path.direction)
		levelSpeed := float32(200) * float32(path.direction*-1)

		if gl.viewPort.speed.X > 0 { // Scrolling to the right
			if gl.viewPort.bounds.MaxX < gl.levelBounds.MaxX {
				gl.viewPort.bounds.MinX += gl.viewPort.speed.X * deltaTime
				gl.viewPort.bounds.MaxX += gl.viewPort.speed.X * deltaTime
				gl.levelBounds.MinX += levelSpeed * deltaTime
				gl.levelBounds.MaxX += levelSpeed * deltaTime
			} else {
				diff := gl.viewPort.bounds.MaxX - gl.levelBounds.MaxX
				gl.viewPort.bounds.MinX -= diff
				gl.viewPort.bounds.MaxX = gl.levelBounds.MaxX
				gl.viewPort.speed.X = 0 // Stop horizontal scrolling once reached
				gl.viewPort.pathIndex += 1
			}
		} else if gl.viewPort.speed.X < 0 { // Scrolling to the left (if implemented)
			if gl.viewPort.bounds.MinX > gl.levelBounds.MinX {
				gl.viewPort.bounds.MinX += gl.viewPort.speed.X * deltaTime
				gl.viewPort.bounds.MaxX += gl.viewPort.speed.X * deltaTime
				gl.levelBounds.MinX += levelSpeed * deltaTime
				gl.levelBounds.MaxX += levelSpeed * deltaTime
			} else {
				diff := gl.levelBounds.MinX - gl.viewPort.bounds.MinX
				gl.viewPort.bounds.MaxX += diff
				gl.viewPort.bounds.MinX = gl.levelBounds.MinX
				gl.viewPort.speed.X = 0 // Stop horizontal scrolling once reached
				gl.viewPort.pathIndex += 1
			}
		}
	}

	if gl.viewPort.pathIndex < len(gl.viewPort.paths) && gl.viewPort.paths[gl.viewPort.pathIndex].scroll == SCROLL_VERTICALLY {
		currentPathIndex := gl.viewPort.pathIndex
		path := gl.viewPort.paths[currentPathIndex]
		gl.viewPort.speed.Y = path.speed * float32(path.direction)
		levelSpeed := float32(500) * float32(path.direction*-1)

		if gl.viewPort.speed.Y > 0 { // Scrolling down
			if gl.viewPort.bounds.MaxY < gl.levelBounds.MaxY {
				gl.viewPort.bounds.MinY += gl.viewPort.speed.Y * deltaTime
				gl.viewPort.bounds.MaxY += gl.viewPort.speed.Y * deltaTime
				gl.levelBounds.MinY += levelSpeed * deltaTime
				gl.levelBounds.MaxY += levelSpeed * deltaTime
			} else {
				diff := gl.viewPort.bounds.MaxY - gl.levelBounds.MaxY
				gl.viewPort.bounds.MinY -= diff
				gl.viewPort.bounds.MaxY = gl.levelBounds.MaxY
				gl.viewPort.speed.Y = 0
				gl.viewPort.pathIndex += 1
			}
		} else if gl.viewPort.speed.Y < 0 { // Scrolling up
			if gl.viewPort.bounds.MinY > gl.levelBounds.MinY {
				gl.viewPort.bounds.MinY += gl.viewPort.speed.Y * deltaTime
				gl.viewPort.bounds.MaxY += gl.viewPort.speed.Y * deltaTime
				gl.levelBounds.MinY += levelSpeed * deltaTime
				gl.levelBounds.MaxY += levelSpeed * deltaTime
			} else {
				diff := gl.levelBounds.MinY - gl.viewPort.bounds.MinY
				gl.viewPort.bounds.MaxY += diff
				gl.viewPort.bounds.MinY = gl.levelBounds.MinY
				gl.viewPort.speed.Y = 0
				gl.viewPort.pathIndex += 1
			}
		}
	}
}

func (gl *GameLevel1) Completed() bool {
	return gl.completed
}
