package service

import (
	"context"

	"flamingo/app/push/internal/handler"
	"flamingo/pkg/broker"

	"github.com/go-kratos/kratos/v2/log"
)

const notificationTopic = "notification.*.*.*"

// StartNotificationListener prepares push-service broker subscription lifecycle
// hooks. The returned start/stop functions match the kratos Option signature
// (func(context.Context) error) and are meant to be passed to
// kratos.BeforeStart / kratos.AfterStop.
//
// NOTE: realtime (online users, via WebSocket) and push (offline users, via
// JPush) both receive these events. Without an online-status check, an online
// user may get a duplicate. To deduplicate, add a user OnlineStatus RPC
// (e.g. GetUserOnlineStatus) and skip forwarding when the target user is
// online.
func StartNotificationListener(b broker.Broker, svc handler.PushService, logger log.Logger) (
	start func(context.Context) error,
	stop func(context.Context) error,
) {
	helper := log.NewHelper(logger)
	var sub broker.Subscription

	start = func(context.Context) error {
		var err error
		sub, err = b.Subscribe(notificationTopic, func(msg *broker.Message) error {
			handleNotification(svc, logger, msg)
			return nil
		})
		if err != nil {
			helper.Errorw("msg", "failed to subscribe push notifications",
				"topic", notificationTopic, "error", err)
			return err
		}
		helper.Infow("msg", "push notification subscriber started",
			"topic", notificationTopic)
		return nil
	}

	stop = func(context.Context) error {
		if sub == nil {
			return nil
		}
		helper.Infow("msg", "push notification subscriber stopping")
		return sub.Unsubscribe()
	}

	return start, stop
}
