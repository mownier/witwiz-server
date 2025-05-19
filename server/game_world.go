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
	playerAcceleration    float32 = 800
	playerDeceleration    float32 = 1600
	playerMaxSpeed        float32 = 1200
	tickRate                      = time.Millisecond * 50
	defaultViewportWidth  float32 = 1080
	defaultViewportHeight float32 = 720
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
		Players:        []*pb.PlayerState{},
		CharacterIds:   []int32{1, 2, 3, 4, 5},
		LevelId:        0,
		GameStarted:    false,
		GameOver:       false,
		GamePaused:     false,
		IsInitial:      false,
		ViewportBounds: &pb.Bounds{MinX: 0, MinY: 0, MaxX: 0, MaxY: 0},
		LevelBounds:    &pb.Bounds{MinX: 0, MinY: 0, MaxX: 0, MaxY: 0},
		ViewportSize:   &pb.Size{Width: 0, Height: 0},
		LevelSize:      &pb.Size{Width: 0, Height: 0},
	}
}

func newPlayerState(playerId int32) *pb.PlayerState {
	boundingBox := &pb.Size{Width: 32, Height: 32}
	return &pb.PlayerState{
		PlayerId:         playerId,
		CharacterId:      0,
		MaxSpeed:         playerMaxSpeed,
		ViewportPosition: &pb.Point{X: boundingBox.Width / 2, Y: boundingBox.Height / 2},
		LevelPosition:    &pb.Point{X: boundingBox.Width / 2, Y: boundingBox.Height / 2},
		Velocity:         &pb.Vector{X: 0, Y: 0},
		Acceleration:     &pb.Vector{X: 0, Y: 0},
		TargetVelocity:   &pb.Vector{X: 0, Y: 0},
		BoundingBox:      boundingBox,
	}
}

func gameLevelArrangement() []int32 {
	return []int32{1, 2}
}

func (gw *gameWorld) changeLevel(levelId int32) {
	var level gl.GameLevel

	gw.gameLevelMu.Lock()

	viewportSize := &pb.Size{Width: defaultViewportWidth, Height: defaultViewportHeight}

	switch levelId {
	case 1:
		level = gl.NewGameLevel1(viewportSize)
		gw.gameLevel = level

	case 2:
		level = gl.NewGameLevel2(viewportSize)
		gw.gameLevel = level
	}

	gw.gameLevelMu.Unlock()

	if level == nil {
		return
	}

	gw.gameStateMu.Lock()

	for _, player := range gw.gameState.Players {
		player.ViewportPosition.X = 0
		player.ViewportPosition.Y = 0
	}
	gw.gameState.LevelId = levelId
	gw.gameState.LevelSize = level.LevelSize()
	gw.gameState.ViewportSize = level.ViewportSize()
	gw.gameState.LevelBounds = level.LevelBounds()
	gw.gameState.ViewportBounds = level.ViewportBounds()

	gw.gameStateMu.Unlock()
}

func (gw *gameWorld) addPlayer(player *pb.PlayerState, stream pb.WitWiz_JoinGameServer) {
	gw.gameStateMu.Lock()
	gw.gameState.Players = append(gw.gameState.Players, player)
	gw.gameStateMu.Unlock()

	gw.playerConnectionsMu.Lock()
	gw.playerConnections[player.PlayerId] = &playerConnection{playerId: player.PlayerId, stream: stream}
	gw.playerConnectionsMu.Unlock()

	log.Printf("player %d joined the game world\n", player.PlayerId)
}

func (gw *gameWorld) removePlayer(playerId int32) {
	gw.gameStateMu.Lock()
	index := -1
	for i, p := range gw.gameState.Players {
		if p.PlayerId == playerId {
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
		if p.PlayerId == input.PlayerId {
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
					newGameState.Players = append(newGameState.Players, newPlayerState(p.PlayerId))
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
		viewportBounds := gw.gameLevel.ViewportBounds()
		levelBounds := gw.gameLevel.LevelBounds()
		nextLevelPortal := gw.gameLevel.NextLevelPortal()
		gw.gameLevelMu.Unlock()

		gameOver := false
		nextLevelId := int32(-1)
		shouldUpdateViewPortBounds := false

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

			shouldUpdateViewPortBounds = true

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
			player.ViewportPosition.X += player.Velocity.X * deltaTime
			player.ViewportPosition.Y += player.Velocity.Y * deltaTime

			// Bounds check
			if player.ViewportPosition.X <= viewportBounds.MinX+player.BoundingBox.Width/2 {
				player.ViewportPosition.X = viewportBounds.MinX + player.BoundingBox.Width/2
				player.Velocity.X = 0
			} else if player.ViewportPosition.X >= viewportBounds.MaxX-player.BoundingBox.Width/2 {
				player.ViewportPosition.X = viewportBounds.MaxX - player.BoundingBox.Width/2
				player.Velocity.X = 0
			}

			if player.ViewportPosition.Y <= viewportBounds.MinY+player.BoundingBox.Height/2 {
				player.ViewportPosition.Y = viewportBounds.MinY + player.BoundingBox.Height/2
				player.Velocity.Y = 0
			} else if player.ViewportPosition.Y >= viewportBounds.MaxY-player.BoundingBox.Height/2 {
				player.ViewportPosition.Y = viewportBounds.MaxY - player.BoundingBox.Height/2
				player.Velocity.Y = 0
			}

			player.LevelPosition.X = player.ViewportPosition.X - levelBounds.MinX
			player.LevelPosition.Y = player.ViewportPosition.Y - levelBounds.MinY

			// Check Next Level Portal
			if nextLevelPortal != nil {
				gw.gameState.NextLevelPortal = nextLevelPortal
				bounds1 := &pb.Bounds{
					MinX: player.LevelPosition.X - player.BoundingBox.Width/2,
					MaxX: player.LevelPosition.X + player.BoundingBox.Width/2,
					MinY: player.LevelPosition.Y - player.BoundingBox.Height/2,
					MaxY: player.LevelPosition.Y + player.BoundingBox.Height/2,
				}
				bounds2 := &pb.Bounds{
					MinX: nextLevelPortal.Position.X - nextLevelPortal.BoundingBox.Width/2,
					MaxX: nextLevelPortal.Position.X + nextLevelPortal.BoundingBox.Width/2,
					MinY: nextLevelPortal.Position.Y - nextLevelPortal.BoundingBox.Height/2,
					MaxY: nextLevelPortal.Position.Y + nextLevelPortal.BoundingBox.Height/2,
				}
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

		// Compute player view port bounds
		if shouldUpdateViewPortBounds {
			gw.gameLevelMu.Lock()
			gw.gameLevel.UpdateViewportBounds(deltaTime)
			updatedViewportBounds := gw.gameLevel.ViewportBounds()
			updatedLevelBounds := gw.gameLevel.LevelBounds()
			gw.gameLevelMu.Unlock()

			gw.gameState.ViewportBounds = updatedViewportBounds
			gw.gameState.LevelBounds = updatedLevelBounds
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
