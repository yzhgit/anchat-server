package service

import (
	"context"
	"time"

	"flamingo/pkg/app"
	"flamingo/pkg/broker"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

const (
	autoDeleteBatchSize = 1000
	autoDeleteInterval  = 1 * time.Minute
)

var ProviderSet = wire.NewSet(
	NewMessageService,
	NewMessageLifecycle,
)

// NewMessageLifecycle wires the auto-delete worker into the kratos app
// lifecycle so the worker's start/stop is driven by kratos rather than by
// newApp or by an ad-hoc initWorker call in main.
func NewMessageLifecycle(logger log.Logger, repo MessageRepository, broker broker.Broker) (app.Lifecycle, func(), error) {
	w, cleanup, err := NewAutoDeleteWorker(logger, repo, broker, autoDeleteBatchSize, autoDeleteInterval)
	if err != nil {
		return app.Lifecycle{}, nil, err
	}
	return app.Lifecycle{
		Start: func(context.Context) error {
			w.StartAsync()
			return nil
		},
		Stop: func(context.Context) error {
			cleanup()
			return nil
		},
	}, cleanup, nil
}
