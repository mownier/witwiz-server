package server

import (
	"log"
	"sync"
	"time"
	gl "witwiz/game_level"
	pb "witwiz/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	playerAcceleration float32 = 800
	playerDeceleration float32 = 1600
	playerMaxSpeed     float32 = 1200
	tickRate                   = time.Millisecond * 5
)

type gameWorld struct {
	gameState           *pb.GameStateUpdate
	gameStateMu         sync.Mutex
	playerConnections   map[int32]*playerConnection
	playerConnectionsMu sync.Mutex
	gameLevel           gl.GameLevel
	gameLevels          []int32
	gameLevelMu         sync.Mutex
	restarting          bool
}

func newGameWorld() *gameWorld {
	return &gameWorld{
		gameState:         newGameStateUpdate(),
		playerConnections: make(map[int32]*playerConnection),
		gameLevel:         nil,
		gameLevels:        gameLevelArrangement(),
		restarting:        false,
	}
}

func newGameStateUpdate() *pb.GameStateUpdate {
	return &pb.GameStateUpdate{
		IsInitial: false,

		GameStarted: false,
		GameOver:    false,
		GamePaused:  false,

		LevelId:       0,
		LevelPosition: &pb.Point{X: 0, Y: 0},
		LevelSize:     &pb.Size{Width: 0, Height: 0},
		LevelEdges:    []*pb.LevelEdgeState{},

		Players:      []*pb.PlayerState{},
		CharacterIds: []int32{1, 2, 3, 4, 5},

		NextLevelPortal: nil,
		Obstacles:       []*pb.ObstacleState{},
	}
}

func newPlayerState(playerId int32) *pb.PlayerState {
	size := &pb.Size{Width: 32, Height: 32}
	return &pb.PlayerState{
		Id:          playerId,
		CharacterId: 0,

		MaxSpeed: playerMaxSpeed,

		Velocity:       &pb.Vector{X: 0, Y: 0},
		Acceleration:   &pb.Vector{X: 0, Y: 0},
		TargetVelocity: &pb.Vector{X: 0, Y: 0},

		Size:     size,
		Position: &pb.Point{X: 0, Y: 0},
	}
}

func gameLevelArrangement() []int32 {
	return []int32{1, 2, 3, 4}
}

func (gw *gameWorld) changeLevel(levelId int32) {
	var level gl.GameLevel

	gw.gameLevelMu.Lock()

	switch levelId {
	case 1:
		level = gl.NewGameLevel1()
		gw.gameLevel = level

	case 2:
		level = gl.NewGameLevel2()
		gw.gameLevel = level

	case 3:
		level = gl.NewGameLevel3()
		gw.gameLevel = level

	case 4:
		level = gl.NewGameLevel4()
		gw.gameLevel = level
	}

	gw.gameLevelMu.Unlock()

	if level == nil {
		return
	}

	gw.gameStateMu.Lock()

	gw.gameState.LevelId = levelId
	gw.gameState.LevelSize = level.LevelSize()
	gw.gameState.LevelPosition = level.LevelPosition()
	gw.gameState.Obstacles = level.LevelObstacles()
	gw.gameState.LevelEdges = level.LevelEdges()
	gw.gameState.NextLevelPortal = nil
	gw.gameState.TileChunks = level.TileChunks()

	spawnPosition := level.PlayerSpawnPosition()
	playerHeight := 32
	playerVSpacing := 128
	playerCount := len(gw.gameState.Players)
	totalPlayerHeight := playerHeight*playerCount + playerVSpacing*(playerCount-1)
	bounds := pb.NewBounds(&pb.Size{Width: 0, Height: float32(totalPlayerHeight)}, spawnPosition)
	offsetY := bounds.MinY + float32(playerHeight)/2

	for index, player := range gw.gameState.Players {
		player.Position.X = spawnPosition.X
		player.Position.Y = offsetY + float32(index*playerHeight) + float32(playerVSpacing*index)
	}

	gw.gameStateMu.Unlock()
}

func (gw *gameWorld) addPlayer(player *pb.PlayerState, stream pb.WitWiz_JoinGameServer) {
	gw.gameStateMu.Lock()
	gw.gameState.Players = append(gw.gameState.Players, player)
	if gw.gameState.GameStarted {
		gw.gameLevelMu.Lock()
		if gw.gameLevel != nil {
			spawnPosition := gw.gameLevel.PlayerSpawnPosition()
			player.Position.X = spawnPosition.X
			player.Position.Y = spawnPosition.Y
		}
		gw.gameLevelMu.Unlock()
	}
	gw.gameStateMu.Unlock()

	gw.playerConnectionsMu.Lock()
	gw.playerConnections[player.Id] = &playerConnection{playerId: player.Id, stream: stream}
	gw.playerConnectionsMu.Unlock()

	log.Printf("player %d joined the game world\n", player.Id)
}

func (gw *gameWorld) removePlayer(playerId int32) {
	gw.gameStateMu.Lock()
	index := -1
	for i, p := range gw.gameState.Players {
		if p.Id == playerId {
			index = i
			break
		}
	}
	if index >= 0 {
		gw.gameState.Players = DeleteElementOrdered(gw.gameState.Players, index)
	}
	gw.gameStateMu.Unlock()

	gw.playerConnectionsMu.Lock()
	delete(gw.playerConnections, playerId)
	gw.playerConnectionsMu.Unlock()

	log.Printf("player %d left the game world\n", playerId)
}

func (gw *gameWorld) processInput(input *pb.PlayerInput) error {
	gw.gameStateMu.Lock()
	defer gw.gameStateMu.Unlock()

	var player *pb.PlayerState
	for _, p := range gw.gameState.Players {
		if p.Id == input.PlayerId {
			player = p
			break
		}
	}
	if player == nil {
		return status.Error(codes.NotFound, "player not found")
	}

	switch input.Action {
	case pb.PlayerInput_SELECT_CHARACTER:
		player.CharacterId = input.CharacterId
		gw.gameState.GameStarted = true
	}

	if !gw.gameState.GameStarted {
		return nil
	}

	switch input.Action {
	case pb.PlayerInput_PAUSE_RESUME:
		for _, characId := range gw.gameState.CharacterIds {
			if characId == player.CharacterId {
				gw.gameState.GamePaused = !gw.gameState.GamePaused
				break
			}
		}
	}

	if gw.gameState.GamePaused {
		player.Acceleration.X = 0.0
		player.Acceleration.Y = 0.0
		player.TargetVelocity.X = 0.0
		player.TargetVelocity.Y = 0.0
		return nil
	}

	switch input.Action {
	case pb.PlayerInput_MOVE_RIGHT_START:
		player.Acceleration.X = playerAcceleration
		player.TargetVelocity.X = playerMaxSpeed

	case pb.PlayerInput_MOVE_RIGHT_STOP:
		player.Acceleration.X = 0.0
		player.TargetVelocity.X = 0.0

	case pb.PlayerInput_MOVE_LEFT_START:
		player.Acceleration.X = -playerAcceleration
		player.TargetVelocity.X = -playerMaxSpeed

	case pb.PlayerInput_MOVE_LEFT_STOP:
		player.Acceleration.X = 0.0
		player.TargetVelocity.X = 0.0

	case pb.PlayerInput_MOVE_UP_START:
		player.Acceleration.Y = playerAcceleration
		player.TargetVelocity.Y = playerMaxSpeed

	case pb.PlayerInput_MOVE_UP_STOP:
		player.Acceleration.Y = 0.0
		player.TargetVelocity.Y = 0.0

	case pb.PlayerInput_MOVE_DOWN_START:
		player.Acceleration.Y = -playerAcceleration
		player.TargetVelocity.Y = -playerMaxSpeed

	case pb.PlayerInput_MOVE_DOWN_STOP:
		player.Acceleration.Y = 0.0
		player.TargetVelocity.Y = 0.0
	}

	return nil
}

func (gw *gameWorld) runGameLoop() {
	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	previousTime := time.Now()

	for range ticker.C {
		currentTime := time.Now()
		deltaTime := float32(currentTime.Sub(previousTime).Seconds())
		previousTime = currentTime

		gw.gameStateMu.Lock()

		if !gw.gameState.GameStarted || gw.gameState.GamePaused {
			gw.gameStateMu.Unlock()
			gw.sendGameStateUpdates()
			continue
		}

		if gw.gameState.GameOver {
			if gw.restarting {
				gw.gameStateMu.Unlock()
				gw.sendGameStateUpdates()
				continue
			}
			gw.restarting = true
			go func() {
				restartTicker := time.NewTicker(5 * time.Second)
				defer restartTicker.Stop()
				<-restartTicker.C
				gw.gameStateMu.Lock()
				newGameState := newGameStateUpdate()
				for _, p := range gw.gameState.Players {
					newGameState.Players = append(newGameState.Players, newPlayerState(p.Id))
				}
				gw.gameState = newGameState
				gw.restarting = false
				gw.gameStateMu.Unlock()

				gw.gameLevelMu.Lock()
				gw.gameLevel = nil
				gw.gameLevels = gameLevelArrangement()
				gw.gameLevelMu.Unlock()
			}()
			gw.gameStateMu.Unlock()
			gw.sendGameStateUpdates()
			continue
		}

		gw.gameLevelMu.Lock()
		if gw.gameLevel == nil {
			nextLevelId := gw.gameLevels[0]
			gw.gameLevels = DeleteElementOrdered(gw.gameLevels, 0)
			gw.gameLevelMu.Unlock()
			gw.gameStateMu.Unlock()
			gw.changeLevel(nextLevelId)
			gw.gameLevelMu.Lock()
			gw.gameStateMu.Lock()
		}
		levelVelocity := gw.gameLevel.LevelVelocity()
		levelEdges := gw.gameLevel.LevelEdges()
		levelSize := gw.gameLevel.LevelSize()
		nextLevelPortal := gw.gameLevel.NextLevelPortal()
		gw.gameState.Obstacles = gw.gameLevel.LevelObstacles()
		gw.gameLevelMu.Unlock()

		gameOver := false
		nextLevelId := int32(-1)
		shouldUpdateLevelPosition := false

		for _, player := range gw.gameState.Players {
			// Check if player has selected a character
			var didSelectCharacter bool = false
			for _, characId := range gw.gameState.CharacterIds {
				if player.CharacterId == characId {
					didSelectCharacter = true
					break
				}
			}
			if !didSelectCharacter {
				// Do nothing since player has not yet selected a character
				continue
			}

			shouldUpdateLevelPosition = true

			// Apply acceleration
			player.Velocity.X += player.Acceleration.X * deltaTime
			player.Velocity.Y += player.Acceleration.Y * deltaTime

			// Apply deceleration if no acceleration is being applied
			if player.Acceleration.X == 0 {
				if player.Velocity.X > 0 {
					player.Velocity.X -= playerDeceleration * deltaTime
					if player.Velocity.X < 0 {
						player.Velocity.X = 0
					}
				} else if player.Velocity.X < 0 {
					player.Velocity.X += playerDeceleration * deltaTime
					if player.Velocity.X > 0 {
						player.Velocity.X = 0
					}
				}
			}

			if player.Acceleration.Y == 0 {
				if player.Velocity.Y > 0 {
					player.Velocity.Y -= playerDeceleration * deltaTime
					if player.Velocity.Y < 0 {
						player.Velocity.Y = 0
					}
				} else if player.Velocity.Y < 0 {
					player.Velocity.Y += playerDeceleration * deltaTime
					if player.Velocity.Y > 0 {
						player.Velocity.Y = 0
					}
				}
			}

			// Apply speed limit
			if player.Velocity.X > player.MaxSpeed {
				player.Velocity.X = player.MaxSpeed
			}
			if player.Velocity.X < -player.MaxSpeed {
				player.Velocity.X = -playerMaxSpeed
			}
			if player.Velocity.Y > player.MaxSpeed {
				player.Velocity.Y = playerMaxSpeed
			}
			if player.Velocity.Y < -player.MaxSpeed {
				player.Velocity.Y = -playerMaxSpeed
			}

			// Update position
			potentialLevelPosX := player.Position.X + player.Velocity.X*deltaTime
			potentialLevelPosY := player.Position.Y + player.Velocity.Y*deltaTime

			// 1. Resolve X-axis movement
			playerBoundsAtPotentialX := pb.NewBounds(player.Size, &pb.Point{X: potentialLevelPosX, Y: player.Position.Y})

			for _, obstacle := range gw.gameState.Obstacles {
				obstacleBounds := pb.NewBounds(obstacle.Size, obstacle.Position)
				if checkCollision(playerBoundsAtPotentialX, obstacleBounds) {
					// Collision on X-axis detected.
					// Determine which side the player hit and clamp their position.
					if player.Velocity.X > 0 { // Moving right, hit obstacle's left side
						player.Position.X = obstacleBounds.MinX - player.Size.Width/2
					} else if player.Velocity.X < 0 { // Moving left, hit obstacle's right side
						player.Position.X = obstacleBounds.MaxX + player.Size.Width/2
					}
					player.Velocity.X = 0
					// No need to check other obstacles on this axis if we clamped.
					// If you have multiple obstacles very close, you might need to iterate further
					// or resolve based on the closest collision. For simplicity, we break.
					break
				}
			}
			// If no X-collision, update X position
			if player.Velocity.X != 0 { // Only update if not stopped by collision
				player.Position.X = potentialLevelPosX
			}

			// 2. Resolve Y-axis movement
			// Assume player attempts to move vertically (after potential X correction)
			playerBoundsAtPotentialY := pb.NewBounds(player.Size, &pb.Point{X: player.Position.X, Y: potentialLevelPosY})
			for _, obstacle := range gw.gameState.Obstacles {
				obstacleBounds := pb.NewBounds(obstacle.Size, obstacle.Position)
				if checkCollision(playerBoundsAtPotentialY, obstacleBounds) {
					// Collision on Y-axis detected.
					// Determine which side the player hit and clamp their position.
					if player.Velocity.Y > 0 { // Moving up, hit obstacle's bottom side
						player.Position.Y = obstacleBounds.MinY - player.Size.Height/2
					} else if player.Velocity.Y < 0 { // Moving down, hit obstacle's top side
						player.Position.Y = obstacleBounds.MaxY + player.Size.Height/2
					}
					player.Velocity.Y = 0 // Stop vertical movement
					break
				}
			}
			// If no Y-collision, update Y position
			if player.Velocity.Y != 0 { // Only update if not stopped by collision
				player.Position.Y = potentialLevelPosY
			}

			// Level edges check
			if len(levelEdges) > 0 {
				playerBounds := pb.NewBounds(player.Size, player.Position)
				for _, edge := range levelEdges {
					edgeBounds := pb.NewBounds(edge.Size, edge.Position)
					if edge.Id == gl.LEVEL_EDGE_LEFT || edge.Id == gl.LEVEL_EDGE_RIGHT {
						if checkCollision(playerBounds, edgeBounds) {
							if edge.Id == gl.LEVEL_EDGE_LEFT && player.Velocity.X == 0 {
								levelVelocityX := player.Position.X + levelVelocity.X*deltaTime
								edgeBoundsX := edgeBounds.MaxX + player.Size.Width/2
								diff := edgeBoundsX - levelVelocityX
								player.Position.X = edgeBoundsX + diff/2

							} else if edge.Id == gl.LEVEL_EDGE_RIGHT && player.Velocity.X == 0 {
								levelVelocityX := player.Position.X + levelVelocity.X*deltaTime
								edgeBoundsX := edgeBounds.MinX - player.Size.Width/2
								diff := levelVelocityX - edgeBoundsX
								player.Position.X = edgeBoundsX - diff/2

							} else if edge.Id == gl.LEVEL_EDGE_LEFT && player.Velocity.X > 0 {
								player.Velocity.X = levelVelocity.X * -1
								player.Position.X = edgeBounds.MaxX + player.Size.Width/2
								player.Position.X += player.Velocity.X * deltaTime

							} else if edge.Id == gl.LEVEL_EDGE_RIGHT && player.Velocity.X < 0 {
								player.Velocity.X = levelVelocity.X * -1
								player.Position.X = edgeBounds.MinX - player.Size.Width/2
								player.Position.X += player.Velocity.X * deltaTime

							} else if edge.Id == gl.LEVEL_EDGE_LEFT && player.Velocity.X < 0 {
								player.Position.X = edgeBounds.MaxX + player.Size.Width/2
								player.Velocity.X = 0

							} else if edge.Id == gl.LEVEL_EDGE_RIGHT && player.Velocity.X > 0 {
								player.Position.X = edgeBounds.MinX - player.Size.Width/2
								player.Velocity.X = 0

							} else {
								player.Velocity.X = 0
							}
							break
						}
					}
				}

				playerBounds = pb.NewBounds(player.Size, player.Position)
				for _, edge := range levelEdges {
					edgeBounds := pb.NewBounds(edge.Size, edge.Position)
					if edge.Id == gl.LEVEL_EDGE_BOTTOM || edge.Id == gl.LEVEL_EDGE_TOP {
						if checkCollision(playerBounds, edgeBounds) {
							if edge.Id == gl.LEVEL_EDGE_BOTTOM && player.Velocity.Y == 0 {
								levelVelocityY := player.Position.Y + levelVelocity.Y*deltaTime
								edgeBoundsY := edgeBounds.MaxY + player.Size.Width/2
								diff := edgeBoundsY - levelVelocityY
								player.Position.Y = edgeBoundsY + diff/2

							} else if edge.Id == gl.LEVEL_EDGE_TOP && player.Velocity.Y == 0 {
								levelVelocityY := player.Position.Y + levelVelocity.Y*deltaTime
								edgeBoundsY := edgeBounds.MinY - player.Size.Width/2
								diff := levelVelocityY - edgeBoundsY
								player.Position.Y = edgeBoundsY - diff/2

							} else if edge.Id == gl.LEVEL_EDGE_BOTTOM && player.Velocity.Y > 0 {
								player.Velocity.Y = levelVelocity.Y * -1
								player.Position.Y = edgeBounds.MaxY + player.Size.Width/2
								player.Position.Y += player.Velocity.Y * deltaTime

							} else if edge.Id == gl.LEVEL_EDGE_TOP && player.Velocity.Y < 0 {
								player.Velocity.Y = levelVelocity.Y * -1
								player.Position.Y = edgeBounds.MinY - player.Size.Width/2
								player.Position.Y += player.Velocity.Y * deltaTime

							} else if edge.Id == gl.LEVEL_EDGE_BOTTOM && player.Velocity.Y < 0 {
								player.Position.Y = edgeBounds.MaxY + player.Size.Width/2
								player.Velocity.Y = 0

							} else if edge.Id == gl.LEVEL_EDGE_TOP && player.Velocity.Y > 0 {
								player.Position.Y = edgeBounds.MinY - player.Size.Width/2
								player.Velocity.Y = 0

							} else {
								player.Velocity.Y = 0
							}
							break
						}
					}
				}
			}

			// Bounds check
			if player.Position.X < player.Size.Width/2 {
				player.Position.X = player.Size.Width / 2
				player.Velocity.X = 0
			} else if player.Position.X > levelSize.Width-player.Size.Width/2 {
				player.Position.X = levelSize.Width - player.Size.Width/2
				player.Velocity.X = 0
			}

			if player.Position.Y < player.Size.Height/2 {
				player.Position.Y = player.Size.Height / 2
				player.Velocity.Y = 0
			} else if player.Position.Y > levelSize.Height-player.Size.Height/2 {
				player.Position.Y = levelSize.Height - player.Size.Height/2
				player.Velocity.Y = 0
			}

			// Check Next Level Portal
			if nextLevelPortal != nil {
				gw.gameState.NextLevelPortal = nextLevelPortal
				bounds1 := pb.NewBounds(player.Size, player.Position)
				bounds2 := pb.NewBounds(nextLevelPortal.Size, nextLevelPortal.Position)
				collided := checkCollision(bounds1, bounds2)
				if collided {
					if len(gw.gameLevels) == 0 {
						gameOver = true
					} else {
						nextLevelId = gw.gameLevels[0]
						gw.gameLevelMu.Lock()
						gw.gameLevels = DeleteElementOrdered(gw.gameLevels, 0)
						gw.gameLevelMu.Unlock()
					}
				}
			}
		}

		if shouldUpdateLevelPosition {
			gw.gameLevelMu.Lock()
			gw.gameLevel.UpdateLevelPosition(deltaTime)
			updatedLevelPosition := gw.gameLevel.LevelPosition()
			updatedTileChunks := gw.gameLevel.TileChunks()
			gw.gameLevelMu.Unlock()
			gw.gameState.LevelPosition = updatedLevelPosition
			gw.gameState.TileChunks = updatedTileChunks
		}

		if gameOver {
			gw.gameState.GameOver = true
		} else if nextLevelId != -1 {
			gw.gameState.NextLevelPortal = nil
			gw.gameStateMu.Unlock()
			gw.changeLevel(nextLevelId)
			gw.gameStateMu.Lock()
		}

		gw.gameStateMu.Unlock()

		gw.sendGameStateUpdates()
	}
}

func (gw *gameWorld) sendGameStateUpdates() {
	gw.gameStateMu.Lock()
	gameStateToSend := proto.Clone(gw.gameState).(*pb.GameStateUpdate)
	gw.gameStateMu.Unlock()

	playersToRemove := []int32{}
	gw.playerConnectionsMu.Lock()
	for _, conn := range gw.playerConnections {
		select {
		case <-conn.stream.Context().Done():
			log.Printf("stream cancelled for player %d\n", conn.playerId)
			playersToRemove = append(playersToRemove, conn.playerId)

		default:
			if err := conn.stream.Send(gameStateToSend); err != nil {
				log.Printf("failed to send game state to player %d: %v\n", conn.playerId, err)
				playersToRemove = append(playersToRemove, conn.playerId)
			}
		}
	}
	gw.playerConnectionsMu.Unlock()

	for _, pId := range playersToRemove {
		gw.removePlayer(pId)
	}
}

func checkCollision(bounds1, bounds2 *pb.Bounds) bool {
	// Check for x-axis overlap
	xOverlap := bounds1.MaxX > bounds2.MinX && bounds1.MinX < bounds2.MaxX

	// Check for y-axis overlap
	yOverlap := bounds1.MaxY > bounds2.MinY && bounds1.MinY < bounds2.MaxY

	// Collision occurs if there is overlap on both axes
	collided := xOverlap && yOverlap

	return collided
}
