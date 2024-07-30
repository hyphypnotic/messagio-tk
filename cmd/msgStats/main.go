package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type server struct {
	msgstats.UnimplementedMsgStatsServiceServer
}

func (s *server) GetStats(ctx context.Context, req *msgstats.MessageStatsRequest) (*msgstats.MsgStatsResponse, error) {
	messages := []string{"Message 1", "Message 2", "Message 3"} // Заполните вашими данными

	return &msgstats.MsgStatsResponse{Messages: messages}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	msgstats.RegisterMsgStatsServiceServer(grpcServer, &server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
