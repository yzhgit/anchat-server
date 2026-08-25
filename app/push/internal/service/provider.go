package service

import (
	"context"

	"flamingo/app/push/internal/handler"
	"flamingo/pkg/app"
	"flamingo/pkg/broker"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewPushService,
	NewPushLifecycle,
)

// NewPushLifecycle wires the NATS notification listener into the kratos app
// lifecycle so push-service startup/shutdown is driven by kratos rather than
// by newApp.
func NewPushLifecycle(b broker.Broker, svc handler.PushService, logger log.Logger) app.Lifecycle {
	start, stop := StartNotificationListener(b, svc, logger)
	return app.Lifecycle{
		Start: func(ctx context.Context) error { return start(ctx) },
		Stop:  func(ctx context.Context) error { return stop(ctx) },
	}
}
