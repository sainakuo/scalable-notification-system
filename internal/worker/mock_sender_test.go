package worker

import (
	"context"
	"errors"

	"github.com/sainakuo/scalable-notification-system/internal/model"
)

type mockSender struct {
	shouldFail bool
}

func (m *mockSender) Send(
	ctx context.Context,
	task model.Task,
) error {
	if m.shouldFail {
		return errors.New("simulated notification error")
	}

	return nil
}
