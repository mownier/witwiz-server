package game_level

import (
	pb "witwiz/proto"
)

type GameLevel1 struct {
	*baseGameLevel
}

func NewGameLevel1() *GameLevel1 {
	base := newBaseGameLevel(1, &pb.Size{Width: 5120, Height: 1024})
	base.paths = append(base.paths,
		&path{scroll: SCROLL_VERTICALLY, speed: 200, direction: DIRECTION_NEGATIVE},
		&path{scroll: SCROLL_VERTICALLY, speed: 200, direction: DIRECTION_POSITIVE},
		// &path{scroll: SCROLL_HORIZONTALLY, speed: 200, direction: DIRECTION_NEGATIVE},
		// &path{scroll: SCROLL_HORIZONTALLY, speed: 200, direction: DIRECTION_POSITIVE},
	)
	base.pathIndex = 0
	for row := 0; row < base.tileRowCount; row++ {
		for col := 0; col < base.tileColCount; col++ {
			if col%2 == 0 {
				if row%2 == 0 {
					base.tiles[row][col] = 1
				} else {
					base.tiles[row][col] = 2
				}
			} else {
				if row%2 == 0 {
					base.tiles[row][col] = 2
				} else {
					base.tiles[row][col] = 1
				}
			}
		}
	}
	return &GameLevel1{baseGameLevel: base}
}

func (gl *GameLevel1) UpdateLevelPosition(deltaTime float32) {
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
	obstacle := &pb.ObstacleState{
		Id:       1,
		Size:     &pb.Size{Width: 200, Height: 200},
		Position: &pb.Point{X: 400, Y: 400},
	}
	gl.obstacles = append(gl.obstacles, obstacle)
}
