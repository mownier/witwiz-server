package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"witwiz/proto"
	"witwiz/server"

	"google.golang.org/grpc"
)

func main() {
	gameServer := server.NewServer()
	grpcServer := grpc.NewServer()
	proto.RegisterWitWizServer(grpcServer, gameServer)

	shutdownSigChan := make(chan os.Signal, 1)
	signal.Notify(shutdownSigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdownSigChan
		log.Println("will stop serving")
		grpcServer.GracefulStop()
		log.Println("stopped serving")
	}()

	lis, err := net.Listen("tcp", ":40041")
	if err != nil {
		log.Fatalf("unable to listen at 40041: %v\n", err)
	}

	log.Println("server running at 40041")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("unable to serve at 40041: %v\n", err)
	}

	log.Println("server shutdown")
}
