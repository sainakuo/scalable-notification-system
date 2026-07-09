package config

import (
	notificationpb "github.com/sainakuo/scalable-notification-system/proto/notificationpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectNotificationService(addr string) (*grpc.ClientConn, notificationpb.NotificationServiceClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}

	client := notificationpb.NewNotificationServiceClient(conn)

	return conn, client, nil
}
