package server

import (
	"io"
	"log"
	"sync"
	pb "witwiz/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	updateSignal chan struct{}
	gameState    *pb.GameStateUpdate
	gameStateMu  sync.Mutex
	pb.UnimplementedWitWizServer
}

func NewServer() *Server {
	return &Server{
		updateSignal: make(chan struct{}, 1),
		gameState: &pb.GameStateUpdate{
			Players:     []*pb.PlayerState{},
			Projectiles: []*pb.ProjectileState{},
		},
	}
}

func (s *Server) JoinGame(stream pb.WitWiz_JoinGameServer) error {
	select {
	case <-stream.Context().Done():
		return status.Error(codes.Canceled, "cancelled")

	default:
		return s.joinGameInternal(stream)
	}
}

func (s *Server) joinGameInternal(stream pb.WitWiz_JoinGameServer) error {
	s.gameStateMu.Lock()

	if len(s.gameState.Players) >= 2 {
		s.gameStateMu.Unlock()
		return status.Error(codes.OutOfRange, "number of max players reached")
	}

	var playerId int32 = 0
	if len(s.gameState.Players) == 0 {
		playerId = 1
	} else {
		pId := s.gameState.Players[0].PlayerId
		if pId == 1 {
			playerId = 2
		} else {
			playerId = 1
		}
	}

	player := &pb.PlayerState{
		PlayerId:          playerId,
		PositionX:         0,
		PositionY:         0,
		BoundingBoxWidth:  32,
		BoundingBoxHeight: 32,
	}
	s.gameState.Players = append(s.gameState.Players, player)

	log.Printf("player %d joined the game\n", player.PlayerId)

	s.gameStateMu.Unlock()

	defer func() {
		s.gameStateMu.Lock()

		index := -1
		for i, p := range s.gameState.Players {
			if p.PlayerId == player.PlayerId {
				index = i
				break
			}
		}
		if index >= 0 {
			s.gameState.Players = DeleteElementOrdered(s.gameState.Players, index)
		}

		s.gameStateMu.Unlock()

		log.Printf("player %d left the game\n", player.PlayerId)
	}()

	go func() {
		for {
			select {
			case <-stream.Context().Done():
				return

			default:
				input, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						// Normal client disconnection
						log.Printf("player %d disconnected\n", player.PlayerId)
						return
					}
					msg := "failed to receive input"
					log.Printf("%s: %v\n", msg, err)
					return
				}
				if err := s.processInput(input); err != nil {
					msg := "failed to process input"
					log.Printf("%s: %v\n", msg, err)
					return
				}
				log.Printf("input: %v\n", input)
			}
		}
	}()

	initialUpdates := []*pb.GameStateUpdate{
		{
			YourPlayerId: player.PlayerId,
		},
		s.gameState,
	}

	for _, update := range initialUpdates {
		if err := stream.Send(update); err != nil {
			msg := "failed to send initial game state"
			log.Printf("%s: %v\n", msg, err)
			return status.Error(codes.Internal, msg)
		}
	}

	for {
		select {
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "cancelled")

		case <-s.updateSignal:
			s.gameStateMu.Lock()
			if err := stream.Send(s.gameState); err != nil {
				s.gameStateMu.Unlock()
				msg := "failed to send game state"
				log.Printf("%s: %v\n", msg, err)
				return status.Error(codes.Internal, msg)
			}
			s.gameStateMu.Unlock()
		}
	}
}

func (s *Server) processInput(input *pb.PlayerInput) error {
	s.gameStateMu.Lock()
	defer s.gameStateMu.Unlock()

	var player *pb.PlayerState
	for _, p := range s.gameState.Players {
		if p.PlayerId == input.PlayerId {
			player = p
			break
		}
	}
	if player == nil {
		return status.Error(codes.NotFound, "player not found")
	}

	// Handle Action
	switch input.Action {
	case pb.PlayerInput_MOVE_UP:
		player.PositionY += 1
		s.signalUpdate()

	case pb.PlayerInput_MOVE_RIGHT:
		player.PositionX += 1
		s.signalUpdate()

	case pb.PlayerInput_MOVE_DOWN:
		player.PositionY -= 1
		s.signalUpdate()

	case pb.PlayerInput_MOVE_LEFT:
		player.PositionX -= 1
		s.signalUpdate()
	}

	return nil
}

func (s *Server) signalUpdate() {
	select {
	case s.updateSignal <- struct{}{}:
		// send

	default:
		// do not block
	}
}

func DeleteElementOrdered[T any](slice []T, index int) []T {
	if index >= 0 && index < len(slice) {
		return append(slice[:index], slice[index+1:]...)
	}
	return slice
}
