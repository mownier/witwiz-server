package game_level

import (
	pb "witwiz/proto"
)

type GameLevel3 struct {
	*baseGameLevel
}

func NewGameLevel3() *GameLevel3 {
	base := newBaseGameLevel(3, &pb.Size{Width: 5120, Height: 1024})
	base.paths = append(base.paths,
		&path{scroll: SCROLL_VERTICALLY, speed: 200, direction: DIRECTION_NEGATIVE},
		&path{scroll: SCROLL_VERTICALLY, speed: 200, direction: DIRECTION_POSITIVE},
		&path{scroll: SCROLL_HORIZONTALLY, speed: 200, direction: DIRECTION_NEGATIVE},
		&path{scroll: SCROLL_HORIZONTALLY, speed: 200, direction: DIRECTION_POSITIVE},
	)
	base.pathIndex = 0
	base.obstacles = append(base.obstacles,
		&pb.ObstacleState{
			Id:       1,
			Size:     &pb.Size{Width: 200, Height: 200},
			Position: &pb.Point{X: 2048, Y: 400},
		},
	)
	for row := 0; row < base.tileRowCount; row++ {
		for col := 0; col < base.tileColCount; col++ {
			tileId := row*base.tileColCount + col + 1
			base.tiles[row][col] = int32(tileId)
		}
	}

	return &GameLevel3{baseGameLevel: base}
}

func (gl *GameLevel3) UpdateLevelPosition(deltaTime float32) {
	gl.baseGameLevel.UpdateLevelPosition(deltaTime)
	if gl.pathIndex < len(gl.paths) ||
		gl.nextLevelPortal != nil {
		return
	}
	size := &pb.Size{Width: 100, Height: defaultResolutionHeight}
	gl.nextLevelPortal = &pb.NextLevelPortalState{
		Size: size,
		Position: &pb.Point{
			X: defaultResolutionWidth - size.Width/2,
			Y: defaultResolutionHeight - size.Height/2,
		},
	}
}
