package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/sender"
	"github.com/sainakuo/scalable-notification-system/internal/tracing"
	notificationpb "github.com/sainakuo/scalable-notification-system/proto/notificationpb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
	ctx := context.Background()

	cfg := config.LoadConfig()

	tracerProvider, err := tracing.NewTracerProvider(
		ctx,
		"sns-grpc-sender",
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			log.Println("tracer shutdown error:", err)
		}
	}()

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

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(),
		),
	)

	notificationpb.RegisterNotificationServiceServer(grpcServer, newServer(deduplicator))

	log.Println("gRPC sender service started on :50051")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
