package server

import "witwiz/proto"

type Server struct {
	proto.UnimplementedWitWizServer
}

func NewServer() *Server {
	return &Server{}
}
