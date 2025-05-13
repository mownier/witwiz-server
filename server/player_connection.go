package server

import pb "witwiz/proto"

type playerConnection struct {
	playerId int32
	stream   pb.WitWiz_JoinGameServer
}
