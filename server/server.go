package server

import (
	"io"
	"log"
	pb "witwiz/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	gameWorld *gameWorld
	pb.UnimplementedWitWizServer
}

func NewServer() *Server {
	return &Server{
		gameWorld: newGameWorld(),
	}
}

func (s *Server) Serve() error {
	go s.gameWorld.runGameLoop()
	return nil
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
	playerId, err := s.generateUniquePlayerId()
	if err != nil {
		return err
	}

	player := newPlayerState(playerId)

	s.gameWorld.addPlayer(player, stream)

	defer func() {
		s.gameWorld.removePlayer(player.Id)
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
						log.Printf("player %d disconnected\n", player.Id)
						return
					}
					log.Printf("failed to receive input: %v\n", err)
					return
				}
				if err := s.gameWorld.processInput(input); err != nil {
					log.Printf("failed to process input: %v\n", err)
					return
				}
				log.Printf("input from player %d: %v\n", player.Id, input)
			}
		}
	}()

	initialData := newGameStateUpdate()
	initialData.IsInitial = true
	initialData.Players = append(initialData.Players, newPlayerState(playerId))
	if err := stream.Send(initialData); err != nil {
		msg := "failed to send initial data"
		log.Printf("%s: %v\n", msg, err)
		return status.Error(codes.Internal, msg)
	}

	s.gameWorld.gameStateMu.Lock()
	initialGameState := proto.Clone(s.gameWorld.gameState).(*pb.GameStateUpdate)
	s.gameWorld.gameStateMu.Unlock()

	if err := stream.Send(initialGameState); err != nil {
		msg := "failed to send initial game state"
		log.Printf("%s: %v\n", msg, err)
		return status.Error(codes.Internal, msg)
	}

	<-stream.Context().Done()
	return nil
}

func (s *Server) generateUniquePlayerId() (int32, error) {
	s.gameWorld.gameStateMu.Lock()
	defer s.gameWorld.gameStateMu.Unlock()

	if len(s.gameWorld.gameState.Players) >= 2 {
		return -1, status.Error(codes.OutOfRange, "number of max players reached")
	}

	var playerId int32 = 0
	if len(s.gameWorld.gameState.Players) == 0 {
		playerId = 1
	} else {
		pId := s.gameWorld.gameState.Players[0].Id
		if pId == 1 {
			playerId = 2
		} else {
			playerId = 1
		}
	}

	return playerId, nil
}

func DeleteElementOrdered[T any](slice []T, index int) []T {
	if index >= 0 && index < len(slice) {
		return append(slice[:index], slice[index+1:]...)
	}
	return slice
}
