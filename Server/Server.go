package main

import (
	"context"
	
	"log"
	"net"
	t "time"
	"myFolder/myPackage"

	"google.golang.org/grpc"
)

type Server struct {
	myPackage.UnimplementedGetXXXServer
}

func (s *Server) GetTime(ctx context.Context, in *myPackage.GetXXXRequest) (*myPackage.GetXXXReply, error) {
	log.Print("Received XXX request")
	return &myPackage.GetXXXReply{Reply: "Your reply here"}, nil
}

func main() {
	// Create listener tcp on port 9080
	list, err := net.Listen("tcp", ":9080")
	if err != nil {
		log.Fatalf("Failed to listen on port 9080: %v", err)
	}
	grpcServer := grpc.NewServer()
	myPackage.RegisterXXXServer(grpcServer, &Server{})

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to server %v", err)
	}
}