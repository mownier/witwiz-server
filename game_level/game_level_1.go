package game_level

import (
	pb "witwiz/proto"
)

type GameLevel1 struct {
	levelId         int32
	levelSize       *pb.Size
	levelBounds     *pb.Bounds
	levelSpeed      float32
	levelVelocity   float32
	viewport        *viewport
	viewportSize    *pb.Size
	completed       bool
	nextLevelPortal *pb.NextLevelPortalState
}

func NewGameLevel1(viewportSize *pb.Size) *GameLevel1 {
	levelSize := &pb.Size{Width: 5120, Height: 1024}
	return &GameLevel1{
		levelId:       1,
		levelBounds:   &pb.Bounds{MinX: 0, MinY: 0, MaxX: levelSize.Width, MaxY: levelSize.Height},
		levelSize:     levelSize,
		levelSpeed:    200,
		levelVelocity: 0,
		viewport: &viewport{
			bounds:   &pb.Bounds{MinX: 0, MinY: 0, MaxX: viewportSize.Width, MaxY: viewportSize.Height},
			velocity: 0,
			paths: []*path{
				{scroll: SCROLL_VERTICALLY, speed: 20, direction: DIRECTION_POSITIVE},
				// {scroll: SCROLL_HORIZONTALLY, speed: 20, direction: DIRECTION_POSITIVE},
				{scroll: SCROLL_VERTICALLY, speed: 20, direction: DIRECTION_NEGATIVE},
				// {scroll: SCROLL_HORIZONTALLY, speed: 20, direction: DIRECTION_NEGATIVE},
			},
			pathIndex: 0,
		},
		viewportSize: viewportSize,
		completed:    false,
	}
}

func (gl *GameLevel1) LevelId() int32 {
	return gl.levelId
}

func (gl *GameLevel1) LevelSize() *pb.Size {
	return gl.levelSize
}

func (gl *GameLevel1) LevelBounds() *pb.Bounds {
	return gl.levelBounds
}

func (gl *GameLevel1) ViewportSize() *pb.Size {
	return gl.viewportSize
}

func (gl *GameLevel1) ViewportBounds() *pb.Bounds {
	return gl.viewport.bounds
}

func (gl *GameLevel1) UpdateViewportBounds(deltaTime float32) {
	if len(gl.viewport.paths) == 0 {
		return
	}

	if gl.viewport.pathIndex >= len(gl.viewport.paths) {
		// TODO: No path to process. What's next?
		gl.viewport.velocity = 0
		gl.levelVelocity = 0
		if gl.nextLevelPortal == nil {
			boundingBox := &pb.Size{Width: 100, Height: gl.viewportSize.Height}
			gl.nextLevelPortal = &pb.NextLevelPortalState{
				BoundingBox: boundingBox,
				Position: &pb.Point{
					X: gl.viewportSize.Width - boundingBox.Width/2,
					Y: boundingBox.Height / 2,
				},
			}
		}
		return
	}

	path := gl.viewport.paths[gl.viewport.pathIndex]
	gl.viewport.velocity = path.speed * float32(path.direction)
	gl.levelVelocity = gl.levelSpeed * float32(path.direction*DIRECTION_NEGATIVE)

	switch path.scroll {
	case SCROLL_HORIZONTALLY:
		if gl.viewport.velocity > 0 { // Scrolling to the right
			if gl.viewport.bounds.MaxX < gl.levelBounds.MaxX {
				gl.viewport.bounds.MinX += gl.viewport.velocity * deltaTime
				gl.viewport.bounds.MaxX += gl.viewport.velocity * deltaTime
				gl.levelBounds.MinX += gl.levelVelocity * deltaTime
				gl.levelBounds.MaxX += gl.levelVelocity * deltaTime
			} else {
				diff := gl.viewport.bounds.MaxX - gl.levelBounds.MaxX
				gl.viewport.bounds.MinX -= diff
				gl.viewport.bounds.MaxX = gl.levelBounds.MaxX
				gl.viewport.velocity = 0
				gl.levelVelocity = 0
				gl.viewport.pathIndex += 1
			}
		} else if gl.viewport.velocity < 0 { // Scrolling to the left
			if gl.viewport.bounds.MinX > gl.levelBounds.MinX {
				gl.viewport.bounds.MinX += gl.viewport.velocity * deltaTime
				gl.viewport.bounds.MaxX += gl.viewport.velocity * deltaTime
				gl.levelBounds.MinX += gl.levelVelocity * deltaTime
				gl.levelBounds.MaxX += gl.levelVelocity * deltaTime
			} else {
				diff := gl.levelBounds.MinX - gl.viewport.bounds.MinX
				gl.viewport.bounds.MaxX += diff
				gl.viewport.bounds.MinX = gl.levelBounds.MinX
				gl.viewport.velocity = 0
				gl.levelVelocity = 0
				gl.viewport.pathIndex += 1
			}
		}

	case SCROLL_VERTICALLY:
		if gl.viewport.velocity > 0 { // Scrolling down
			if gl.viewport.bounds.MaxY < gl.levelBounds.MaxY {
				gl.viewport.bounds.MinY += gl.viewport.velocity * deltaTime
				gl.viewport.bounds.MaxY += gl.viewport.velocity * deltaTime
				gl.levelBounds.MinY += gl.levelVelocity * deltaTime
				gl.levelBounds.MaxY += gl.levelVelocity * deltaTime
			} else {
				diff := gl.viewport.bounds.MaxY - gl.levelBounds.MaxY
				gl.viewport.bounds.MinY -= diff
				gl.viewport.bounds.MaxY = gl.levelBounds.MaxY
				gl.viewport.velocity = 0
				gl.levelVelocity = 0
				gl.viewport.pathIndex += 1
			}
		} else if gl.viewport.velocity < 0 { // Scrolling up
			if gl.viewport.bounds.MinY > gl.levelBounds.MinY {
				gl.viewport.bounds.MinY += gl.viewport.velocity * deltaTime
				gl.viewport.bounds.MaxY += gl.viewport.velocity * deltaTime
				gl.levelBounds.MinY += gl.levelVelocity * deltaTime
				gl.levelBounds.MaxY += gl.levelVelocity * deltaTime
			} else {
				diff := gl.levelBounds.MinY - gl.viewport.bounds.MinY
				gl.viewport.bounds.MaxY += diff
				gl.viewport.bounds.MinY = gl.levelBounds.MinY
				gl.viewport.velocity = 0
				gl.levelVelocity = 0
				gl.viewport.pathIndex += 1
			}
		}
	}
}

func (gl *GameLevel1) NextLevelPortal() *pb.NextLevelPortalState {
	return gl.nextLevelPortal
}
