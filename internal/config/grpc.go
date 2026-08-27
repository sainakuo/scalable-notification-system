package config

import (
	"fmt"
	notificationpb "github.com/sainakuo/scalable-notification-system/proto/notificationpb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectNotificationService(
	addr string,
	useTLS bool,
	caFile string,
) (*grpc.ClientConn, notificationpb.NotificationServiceClient, error) {
	var transportCredentials credentials.TransportCredentials
	var err error

	if useTLS {
		if caFile == "" {
			return nil, nil, fmt.Errorf(
				"gRPC TLS enabled but CA file is not configured",
			)
		}

		transportCredentials, err = credentials.NewClientTLSFromFile(
			caFile,
			"grpc-sender",
		)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"load gRPC CA certificate: %w",
				err,
			)
		}
	} else {
		transportCredentials = insecure.NewCredentials()
	}

	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithStatsHandler(
			otelgrpc.NewClientHandler(),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("dial notification service: %w", err)
	}

	client := notificationpb.NewNotificationServiceClient(conn)

	return conn, client, nil
}
