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
		updateSignal: make(chan struct{}),
		gameState: &pb.GameStateUpdate{
			Players:     make(map[int32]*pb.PlayerState),
			Projectiles: make(map[int32]*pb.ProjectileState),
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
	for _, player := range s.gameState.Players {
		playerId = max(playerId, player.PlayerId)
	}

	player := &pb.PlayerState{
		PlayerId:          playerId + 1,
		PositionX:         0,
		PositionY:         0,
		BoundingBoxWidth:  32,
		BoundingBoxHeight: 32,
	}
	s.gameState.Players[player.PlayerId] = player

	log.Printf("player %d joined the game\n", player.PlayerId)

	s.gameStateMu.Unlock()

	defer func() {
		s.gameStateMu.Lock()

		delete(s.gameState.Players, player.PlayerId)

		s.gameStateMu.Unlock()

		log.Printf("player %d left the game\n", player.PlayerId)
	}()

	go func() {
		for {
			select {
			case <-stream.Context().Done():
				return

			case <-s.updateSignal:
				s.gameStateMu.Lock()
				defer s.gameStateMu.Unlock()
				if err := stream.Send(s.gameState); err != nil {
					return
				}
			}
		}
	}()

	s.signalUpdate()

	for {
		select {
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "cancelled")

		default:
			input, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// Normal client disconnection
					log.Println("player disconnected")
					return nil
				}
				msg := "failed to receive input"
				log.Printf("%s: %v\n", msg, err)
				return status.Error(codes.Internal, msg)
			}
			if err := s.processInput(input); err != nil {
				return err
			}
		}
	}
}

func (s *Server) processInput(input *pb.PlayerInput) error {
	s.gameStateMu.Lock()
	defer s.gameStateMu.Unlock()

	player, ok := s.gameState.Players[input.PlayerId]
	if !ok {
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
		// Will send

	default:
		// Channel is full, or no receiver, don't block
	}
}
