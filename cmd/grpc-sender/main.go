package main

import (
	"context"
	"fmt"
	"log"
	"net"

	notificationpb "github.com/sainakuo/scalable-notification-system/proto/notificationpb"
	"google.golang.org/grpc"
)

type server struct {
	notificationpb.UnimplementedNotificationServiceServer
}

func (s *server) SendNotification(
	ctx context.Context,
	req *notificationpb.SendNotificationRequest,
) (*notificationpb.SendNotificationResponse, error) {
	fmt.Println("Sending notification:", req.TaskId, req.UserId, req.Type, req.Payload)

	return &notificationpb.SendNotificationResponse{
		Success: true,
		Message: "notification sent",
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()

	notificationpb.RegisterNotificationServiceServer(grpcServer, &server{})

	fmt.Println("gRPC sender service started on :50051")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
