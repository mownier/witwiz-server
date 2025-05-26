package game_level

import (
	pb "witwiz/proto"
)

type GameLevel2 struct {
	*baseGameLevel
}

func NewGameLevel2() *GameLevel2 {
	base := newBaseGameLevel(2, &pb.Size{Width: 5120, Height: 1024})
	base.paths = append(base.paths,
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
			if col%2 == 0 {
				if row%2 == 0 {
					base.tiles[row][col] = 3
				} else {
					base.tiles[row][col] = 4
				}
			} else {
				if row%2 == 0 {
					base.tiles[row][col] = 4
				} else {
					base.tiles[row][col] = 3
				}
			}
		}
	}
	return &GameLevel2{baseGameLevel: base}
}

func (gl *GameLevel2) UpdateLevelPosition(deltaTime float32) {
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
