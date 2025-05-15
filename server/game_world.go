package server

import (
	"log"
	"sync"
	"time"
	pb "witwiz/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	playerAcceleration float32 = 800
	playerDeceleration float32 = 1600
	playerMaxSpeed     float32 = 1200
	tickRate                   = time.Millisecond * 50
)

type gameLevel interface {
	levelId() int32
	worldViewPort() *pb.ViewPort
	computeWorldOffsetX(currentWorldOffsetX float32, deltaTime float32) float32
	completed() bool
}

type gameWorld struct {
	gameState           *pb.GameStateUpdate
	gameStateMu         sync.Mutex
	playerConnections   map[int32]*playerConnection
	playerConnectionsMu sync.Mutex
	playerData          map[int32]*playerData
	playerDataMu        sync.Mutex
	gameLevel           gameLevel
	gameLevelMu         sync.Mutex
	restarting          bool
}

func newGameWorld() *gameWorld {
	return &gameWorld{
		gameState: &pb.GameStateUpdate{
			Players:      []*pb.PlayerState{},
			Projectiles:  []*pb.ProjectileState{},
			WorldOffset:  &pb.Vector2{X: 0, Y: 0},
			GameStarted:  false,
			CharacterIds: []int32{1, 2, 3, 4, 5},
			GameOver:     false,
		},
		playerConnections: make(map[int32]*playerConnection),
		playerData:        make(map[int32]*playerData),
		restarting:        false,
	}
}

func (gw *gameWorld) changeLevel(levelId int32) {
	levelChanged := false

	gw.gameLevelMu.Lock()

	switch levelId {
	case 1:
		gw.gameLevel = newGameLevel1()
		levelChanged = true

	case 2:
		gw.gameLevel = newGameLevel2()
		levelChanged = true
	}

	worldViewPort := gw.gameLevel.worldViewPort()

	gw.gameLevelMu.Unlock()

	if !levelChanged {
		return
	}

	gw.gameStateMu.Lock()

	for _, player := range gw.gameState.Players {
		player.Position.X = 0
		player.Position.Y = 0
	}
	gw.gameState.Projectiles = []*pb.ProjectileState{}
	gw.gameState.WorldOffset = &pb.Vector2{X: 0, Y: 0}
	gw.gameState.LevelId = levelId
	gw.gameState.WorldViewPort = worldViewPort

	gw.gameStateMu.Unlock()
}

func (gw *gameWorld) addPlayer(player *pb.PlayerState, stream pb.WitWiz_JoinGameServer) {
	gw.gameStateMu.Lock()
	gw.gameState.Players = append(gw.gameState.Players, player)
	gw.gameStateMu.Unlock()

	gw.playerConnectionsMu.Lock()
	gw.playerConnections[player.PlayerId] = &playerConnection{playerId: player.PlayerId, stream: stream}
	gw.playerConnectionsMu.Unlock()

	gw.playerDataMu.Lock()
	gw.playerData[player.PlayerId] = &playerData{viewPort: &pb.ViewPort{Width: 0, Height: 0}}
	gw.playerDataMu.Unlock()

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

	gw.playerDataMu.Lock()
	delete(gw.playerData, playerId)
	gw.playerDataMu.Unlock()

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

	// Handle action on game not started
	switch input.Action {
	case pb.PlayerInput_SELECT_CHARACTER:
		player.CharacterId = input.CharacterId

	case pb.PlayerInput_START_GAME:
		gw.gameState.GameStarted = true

	case pb.PlayerInput_REPORT_VIEWPORT:
		gw.playerDataMu.Lock()
		if playerData, exists := gw.playerData[player.PlayerId]; exists {
			playerData.viewPort = input.ViewPort
		}
		gw.playerDataMu.Unlock()
	}

	if !gw.gameState.GameStarted {
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

		if !gw.gameState.GameStarted {
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
				newGameState := &pb.GameStateUpdate{
					Players:      []*pb.PlayerState{},
					Projectiles:  []*pb.ProjectileState{},
					WorldOffset:  &pb.Vector2{X: 0, Y: 0},
					GameStarted:  false,
					CharacterIds: []int32{1, 2, 3, 4, 5},
					GameOver:     false,
				}
				for _, p := range gw.gameState.Players {
					newGameState.Players = append(newGameState.Players, &pb.PlayerState{
						PlayerId:       p.PlayerId,
						Position:       &pb.Vector2{X: 0, Y: 0},
						Acceleration:   &pb.Vector2{X: 0, Y: 0},
						Velocity:       &pb.Vector2{X: 0, Y: 0},
						TargetVelocity: &pb.Vector2{X: 0, Y: 0},
						BoundingBox:    &pb.BoundingBox{Width: 32, Height: 32},
						MaxSpeed:       playerMaxSpeed,
						CharacterId:    0,
					})
				}
				gw.gameState = newGameState
				gw.restarting = false
				gw.gameStateMu.Unlock()

				gw.gameLevelMu.Lock()
				gw.gameLevel = nil
				gw.gameLevelMu.Unlock()
			}()
			gw.gameStateMu.Unlock()
			gw.sendGameStateUpdates()
			continue
		}

		gw.gameLevelMu.Lock()
		if gw.gameLevel == nil {
			gw.gameLevelMu.Unlock()
			gw.gameStateMu.Unlock()
			gw.changeLevel(1)
			gw.gameLevelMu.Lock()
			gw.gameStateMu.Lock()
		}
		gw.gameLevelMu.Unlock()

		worldShouldMove := false

		for _, player := range gw.gameState.Players {
			if player.CharacterId < 1 {
				// Do nothing since player has not yet selected a character
				continue
			}

			worldShouldMove = true

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
			player.Position.X += player.Velocity.X * deltaTime
			player.Position.Y += player.Velocity.Y * deltaTime

			// **Boundary Checks (Keep position >= 0)**
			if player.Position.X < 0 {
				player.Position.X = 0
				player.Velocity.X = 0
			}
			if player.Position.Y < 0 {
				player.Position.Y = 0
				player.Velocity.Y = 0
			}

			gw.gameLevelMu.Lock()
			worldViewPort := gw.gameLevel.worldViewPort()
			gw.gameLevelMu.Unlock()

			gw.playerDataMu.Lock()
			if playerData, exist := gw.playerData[player.PlayerId]; exist {
				if player.Position.X >= (min(playerData.viewPort.Width, worldViewPort.Width) - player.BoundingBox.Width) {
					player.Position.X = (min(playerData.viewPort.Width, worldViewPort.Width) - player.BoundingBox.Width)
					player.Velocity.X = 0
				}
				if player.Position.Y >= (min(playerData.viewPort.Height, worldViewPort.Height) - player.BoundingBox.Height) {
					player.Position.Y = (min(playerData.viewPort.Height, worldViewPort.Height) - player.BoundingBox.Height)
					player.Velocity.Y = 0
				}
			}
			gw.playerDataMu.Unlock()
		}

		if len(gw.gameState.Players) > 0 && worldShouldMove {
			gw.gameLevelMu.Lock()

			worldOffsetX := gw.gameLevel.computeWorldOffsetX(gw.gameState.WorldOffset.X, deltaTime)
			levelId := gw.gameLevel.levelId()
			levelCompleted := gw.gameLevel.completed()

			gw.gameLevelMu.Unlock()

			gw.gameState.WorldOffset.X = worldOffsetX

			if levelCompleted {
				if levelId == 2 {
					gw.gameState.GameOver = true
				} else {
					nextLevelId := levelId + 1
					gw.gameStateMu.Unlock()
					gw.changeLevel(nextLevelId)
					gw.gameStateMu.Lock()
				}
			}
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
