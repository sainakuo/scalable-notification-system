package sender

import (
	"context"
	"fmt"

	"github.com/sainakuo/scalable-notification-system/internal/model"
	notificationpb "github.com/sainakuo/scalable-notification-system/proto/notificationpb"
)

type GRPCSender struct {
	client notificationpb.NotificationServiceClient
}

func NewGRPCSender(
	client notificationpb.NotificationServiceClient,
) *GRPCSender {
	if client == nil {
		panic("notification client is nil")
	}

	return &GRPCSender{
		client: client,
	}
}

func (s *GRPCSender) Send(
	ctx context.Context,
	task model.Task,
) error {
	response, err := s.client.SendNotification(
		ctx,
		&notificationpb.SendNotificationRequest{
			TaskId:  int32(task.ID),
			UserId:  int32(task.UserID),
			Type:    task.Type,
			Payload: task.Payload,
		},
	)
	if err != nil {
		return fmt.Errorf("send notification via grpc: %w", err)
	}

	if !response.Success {
		return fmt.Errorf(
			"notification sender failed: %s",
			response.Message,
		)
	}

	return nil
}
