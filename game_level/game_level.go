package game_level

import (
	"math"
	pb "witwiz/proto"
)

const (
	defaultResolutionWidth  float32 = 1080
	defaultResolutionHeight float32 = 720
	defaultLevelEdgeAlongX  float32 = 8
	defaultLevelEdgeAlongY  float32 = 8
	tileSize                int     = 32
	chunkSize               int     = 16
)

type GameLevel interface {
	LevelId() int32
	LevelSize() *pb.Size
	LevelPosition() *pb.Point
	LevelVelocity() *pb.Vector
	LevelObstacles() []*pb.ObstacleState
	LevelEdges() []*pb.LevelEdgeState
	UpdateLevelPosition(deltaTime float32)
	NextLevelPortal() *pb.NextLevelPortalState
	TileChunks() []*pb.TileChunk
}

type path struct {
	scroll    scroll
	speed     float32
	direction direction
}

type scroll int8

const (
	SCROLL_HORIZONTALLY scroll = 0
	SCROLL_VERTICALLY   scroll = 1
)

type direction int8

const (
	DIRECTION_POSITIVE direction = 1
	DIRECTION_NEGATIVE direction = -1
)

type LevelEdge int8

const (
	LEVEL_EDGE_LEFT   = 1
	LEVEL_EDGE_RIGHT  = 2
	LEVEL_EDGE_BOTTOM = 3
	LEVEL_EDGE_TOP    = 4
)

func defaultLevelEdges(levelSize *pb.Size) []*pb.LevelEdgeState {
	return []*pb.LevelEdgeState{
		{
			Id:       LEVEL_EDGE_LEFT,
			Size:     &pb.Size{Width: defaultLevelEdgeAlongX, Height: levelSize.Height},
			Position: &pb.Point{X: defaultLevelEdgeAlongX / 2, Y: levelSize.Height / 2},
		},
		{
			Id:       LEVEL_EDGE_RIGHT,
			Size:     &pb.Size{Width: defaultLevelEdgeAlongX, Height: levelSize.Height},
			Position: &pb.Point{X: defaultResolutionWidth - defaultLevelEdgeAlongX/2, Y: levelSize.Height / 2},
		},
		{
			Id:       LEVEL_EDGE_BOTTOM,
			Size:     &pb.Size{Width: levelSize.Width, Height: defaultLevelEdgeAlongY},
			Position: &pb.Point{X: levelSize.Width / 2, Y: defaultLevelEdgeAlongY / 2},
		},
		{
			Id:       LEVEL_EDGE_TOP,
			Size:     &pb.Size{Width: levelSize.Width, Height: defaultLevelEdgeAlongY},
			Position: &pb.Point{X: levelSize.Width / 2, Y: defaultResolutionHeight - defaultLevelEdgeAlongY/2},
		},
	}
}

type baseGameLevel struct {
	levelId         int32
	levelSize       *pb.Size
	levelPosition   *pb.Point
	levelVelocity   *pb.Vector
	levelEdges      []*pb.LevelEdgeState
	nextLevelPortal *pb.NextLevelPortalState
	obstacles       []*pb.ObstacleState
	paths           []*path
	pathIndex       int
	tiles           [][]int32
	tileColCount    int
	tileRowCount    int
	chunkRowCount   int
	chunkColCount   int
}

func newBaseGameLevel(levelId int32, levelSize *pb.Size) *baseGameLevel {
	tileRowCount := int(levelSize.Height) / tileSize
	tileColCount := int(levelSize.Width) / tileSize
	tiles := make([][]int32, tileRowCount)
	for row := 0; row < tileRowCount; row++ {
		tiles[row] = make([]int32, tileColCount)
	}
	return &baseGameLevel{
		levelId:         levelId,
		levelSize:       levelSize,
		levelPosition:   &pb.Point{X: 0, Y: 0},
		levelVelocity:   &pb.Vector{X: 0, Y: 0},
		obstacles:       []*pb.ObstacleState{},
		nextLevelPortal: nil,
		paths:           []*path{},
		pathIndex:       -1,
		levelEdges:      defaultLevelEdges(levelSize),
		tiles:           tiles,
		tileRowCount:    tileRowCount,
		tileColCount:    tileColCount,
		chunkRowCount:   tileRowCount / int(chunkSize),
		chunkColCount:   tileColCount / int(chunkSize),
	}
}

func (gl *baseGameLevel) LevelId() int32 {
	return gl.levelId
}

func (gl *baseGameLevel) LevelSize() *pb.Size {
	return gl.levelSize
}

func (gl *baseGameLevel) LevelPosition() *pb.Point {
	return gl.levelPosition
}

func (gl *baseGameLevel) LevelVelocity() *pb.Vector {
	return gl.levelVelocity
}

func (gl *baseGameLevel) LevelObstacles() []*pb.ObstacleState {
	return gl.obstacles
}

func (gl *baseGameLevel) LevelEdges() []*pb.LevelEdgeState {
	return gl.levelEdges
}

func (gl *baseGameLevel) UpdateLevelPosition(deltaTime float32) {
	if len(gl.paths) == 0 {
		return
	}

	if gl.pathIndex >= len(gl.paths) {
		gl.levelVelocity.X = 0
		gl.levelVelocity.Y = 0
		return
	}

	path := gl.paths[gl.pathIndex]

	switch path.scroll {
	case SCROLL_HORIZONTALLY:
		gl.levelVelocity.Y = 0
		gl.levelVelocity.X = path.speed * float32(path.direction)
		gl.levelPosition.X += gl.levelVelocity.X * deltaTime
		// Bounds check
		if gl.levelPosition.X < -gl.levelSize.Width+defaultResolutionWidth {
			gl.levelPosition.X = -gl.levelSize.Width + defaultResolutionWidth
			gl.levelVelocity.X = 0
			gl.pathIndex += 1
		} else if gl.levelPosition.X > 0 {
			gl.levelPosition.X = 0
			gl.levelVelocity.X = 0
			gl.pathIndex += 1
		}

	case SCROLL_VERTICALLY:
		gl.levelVelocity.X = 0
		gl.levelVelocity.Y = path.speed * float32(path.direction)
		gl.levelPosition.Y += gl.levelVelocity.Y * deltaTime
		// Bounds check
		if gl.levelPosition.Y < -gl.levelSize.Height+defaultResolutionHeight {
			gl.levelPosition.Y = -gl.levelSize.Height + defaultResolutionHeight
			gl.levelVelocity.Y = 0
			gl.pathIndex += 1
		} else if gl.levelPosition.Y > 0 {
			gl.levelPosition.Y = 0
			gl.levelVelocity.Y = 0
			gl.pathIndex += 1
		}
	}

	for _, edge := range gl.levelEdges {
		if edge.Id == LEVEL_EDGE_LEFT || edge.Id == LEVEL_EDGE_RIGHT {
			edge.Position.X += gl.levelVelocity.X * -1 * deltaTime
			edgeBounds := pb.NewBounds(edge.Size, edge.Position)
			if edge.Id == LEVEL_EDGE_LEFT && edgeBounds.MinX < 0 {
				edge.Position.X = edge.Size.Width / 2
			} else if edge.Id == LEVEL_EDGE_RIGHT && edgeBounds.MaxX > gl.levelSize.Width {
				edge.Position.X = gl.levelSize.Width - edge.Size.Width/2
			} else if edge.Id == LEVEL_EDGE_LEFT {
				diff := (edge.Position.X + gl.levelPosition.X) - (edge.Size.Width / 2)
				if diff != 0.0000000000000000000000 {
					edge.Position.X -= diff
				}
			} else if edge.Id == LEVEL_EDGE_RIGHT {
				diff := (edge.Position.X + gl.levelPosition.X - defaultResolutionWidth) + (edge.Size.Width / 2)
				if diff != 0.0000000000000000000000 {
					edge.Position.X -= diff
				}
			}

		} else if edge.Id == LEVEL_EDGE_BOTTOM || edge.Id == LEVEL_EDGE_TOP {
			edge.Position.Y += gl.levelVelocity.Y * -1 * deltaTime
			edgeBounds := pb.NewBounds(edge.Size, edge.Position)
			if edge.Id == LEVEL_EDGE_BOTTOM && edgeBounds.MinY < 0 {
				edge.Position.Y = edge.Size.Height / 2
			} else if edge.Id == LEVEL_EDGE_TOP && edgeBounds.MaxY > gl.levelSize.Height {
				edge.Position.Y = gl.levelSize.Height - edge.Size.Height/2
			} else if edge.Id == LEVEL_EDGE_BOTTOM {
				diff := (edge.Position.Y + gl.levelPosition.Y) - (edge.Size.Height / 2)
				if diff != 0.0000000000000000000000 {
					edge.Position.Y -= diff
				}
			} else if edge.Id == LEVEL_EDGE_TOP {
				diff := (edge.Position.Y + gl.levelPosition.Y - defaultResolutionHeight) + (edge.Size.Height / 2)
				if diff != 0.0000000000000000000000 {
					edge.Position.Y -= diff
				}
			}
		}
	}
}

func (gl *baseGameLevel) NextLevelPortal() *pb.NextLevelPortalState {
	return gl.nextLevelPortal
}

func (gl *baseGameLevel) TileChunks() []*pb.TileChunk {
	minX := gl.levelPosition.X * -1
	maxX := minX + defaultResolutionWidth
	minY := gl.levelPosition.Y * -1
	maxY := minY + defaultResolutionHeight

	minTileCol := math.Floor(float64(minX / float32(tileSize)))
	maxTileCol := math.Ceil(float64(maxX / float32(tileSize)))
	minTileRow := math.Floor(float64(minY / float32(tileSize)))
	maxTileRow := math.Ceil(float64(maxY / float32(tileSize)))

	minChunkCol := int(math.Floor(float64(minTileCol / float64(chunkSize))))
	maxChunkCol := int(math.Floor(float64(maxTileCol / float64(chunkSize))))
	minChunkRow := int(math.Floor(float64(minTileRow / float64(chunkSize))))
	maxChunkRow := int(math.Floor(float64(maxTileRow / float64(chunkSize))))

	tileChunks := []*pb.TileChunk{}

	for chunkRow := minChunkRow; chunkRow <= maxChunkRow; chunkRow++ {
		for chunkCol := minChunkCol; chunkCol <= maxChunkCol; chunkCol++ {
			tileChunk := &pb.TileChunk{Row: int32(chunkRow), Col: int32(chunkCol), Tiles: []*pb.Tile{}}

			// Calculate the starting tile coordinates for the current chunk
			startTileCol := chunkCol * chunkSize
			startTileRow := chunkRow * chunkSize

			// Calculate the ending tile coordinates for the current chunk (exclusive)
			endTileCol := startTileCol + chunkSize
			endTileRow := startTileRow + chunkSize

			for tileRow := startTileRow; tileRow < endTileRow; tileRow++ {
				for tileCol := startTileCol; tileCol < endTileCol; tileCol++ {
					if tileRow < 0 || tileRow >= gl.tileRowCount ||
						tileCol < 0 || tileCol >= gl.tileColCount {
						continue
					}
					tileId := gl.tiles[tileRow][tileCol]
					tileChunk.Tiles = append(tileChunk.Tiles, &pb.Tile{Row: int32(tileRow), Col: int32(tileCol), Id: tileId})
				}
			}

			if len(tileChunk.Tiles) == 0 {
				continue
			}

			tileChunks = append(tileChunks, tileChunk)
		}
	}

	return tileChunks
}
