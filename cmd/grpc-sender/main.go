package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/sender"
	notificationpb "github.com/sainakuo/scalable-notification-system/proto/notificationpb"
	"google.golang.org/grpc"
)

type server struct {
	notificationpb.UnimplementedNotificationServiceServer

	deduplicator sender.Deduplicator
}

func newServer(deduplicator sender.Deduplicator) *server {
	if deduplicator == nil {
		panic("deduplicator is nil")
	}

	return &server{
		deduplicator: deduplicator,
	}
}

func (s *server) SendNotification(
	ctx context.Context,
	req *notificationpb.SendNotificationRequest,
) (*notificationpb.SendNotificationResponse, error) {

	taskID := int(req.TaskId)

	processed, err := s.deduplicator.IsProcessed(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("check notification idempotency: %w", err)
	}

	if processed {
		return &notificationpb.SendNotificationResponse{
			Success: true,
			Message: "notification already sent",
		}, nil
	}
	fmt.Println("Sending notification:", req.TaskId, req.UserId, req.Type, req.Payload)

	if err := s.deduplicator.MarkProcessed(ctx, taskID); err != nil {
		return nil, fmt.Errorf("mark notification processed: %w", err)
	}

	return &notificationpb.SendNotificationResponse{
		Success: true,
		Message: "notification sent",
	}, nil
}

func main() {
	cfg := config.LoadConfig()

	redisClient := config.ConnectRedis(cfg.RedisAddr)
	defer redisClient.Close()

	deduplicator := sender.NewRedisDeduplicator(
		redisClient,
		24*time.Hour,
	)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()

	notificationpb.RegisterNotificationServiceServer(grpcServer, newServer(deduplicator))

	log.Println("gRPC sender service started on :50051")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
