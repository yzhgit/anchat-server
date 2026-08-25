package service

import (
	"context"
	"time"

	"flamingo/app/message/internal/model"
	"flamingo/pkg/broker"

	"github.com/go-kratos/kratos/v2/log"
)

type AutoDeleteWorker struct {
	messageRepo MessageRepository
	broker      broker.Broker
	batchSize   int
	interval    time.Duration
	stopCh      chan struct{}
	log         *log.Helper
}

func NewAutoDeleteWorker(
	logger log.Logger,
	messageRepo MessageRepository,
	broker broker.Broker,
	batchSize int,
	interval time.Duration,
) (*AutoDeleteWorker, func(), error) {
	w := &AutoDeleteWorker{
		messageRepo: messageRepo,
		broker:      broker,
		batchSize:   batchSize,
		interval:    interval,
		stopCh:      make(chan struct{}),
		log:         log.NewHelper(logger),
	}
	return w, func() {
		w.Stop()
	}, nil
}

func (w *AutoDeleteWorker) Start() {
	w.log.Infow("msg", "Auto-delete worker started", "batchSize", w.batchSize, "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			w.log.Infow("msg", "Auto-delete worker stopped")
			return
		case <-ticker.C:
			w.cleanup()
		}
	}
}

func (w *AutoDeleteWorker) Stop() {
	close(w.stopCh)
}

func (w *AutoDeleteWorker) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	l := w.log.WithContext(ctx)
	now := time.Now()

	for {
		select {
		case <-w.stopCh:
			return
		default:
		}

		expiredMessages, err := w.messageRepo.GetExpiredMessages(ctx, now, w.batchSize)
		if err != nil {
			l.Errorw("msg", "Failed to get expired messages", "error", err)
			break
		}

		if len(expiredMessages) == 0 {
			break
		}

		messageIDs := make([]string, 0, len(expiredMessages))
		reasons := map[string][]string{
			"auto_delete":        {},
			"burn_after_reading": {},
			"both":               {},
		}
		for _, msg := range expiredMessages {
			messageIDs = append(messageIDs, msg.MessageID)
			reason := inferDeleteReason(msg)
			reasons[reason] = append(reasons[reason], msg.MessageID)
		}

		if err := w.messageRepo.BatchUpdateStatus(ctx, messageIDs, 2); err != nil {
			l.Errorw("msg", "Failed to batch update message status", "error", err)
			break
		}

		l.Infow("msg", "Deleted expired messages", "count", len(messageIDs))

		for reason, ids := range reasons {
			if len(ids) == 0 {
				continue
			}
			w.publishNotification(ctx, ids, reason)
		}
	}
}

func inferDeleteReason(msg *model.Message) string {
	hasAuto := msg.AutoDeleteExpireTime != nil
	hasBurn := msg.BurnAfterReadingExpireTime != nil

	if hasAuto && hasBurn {
		if msg.AutoDeleteExpireTime.Before(*msg.BurnAfterReadingExpireTime) {
			return "auto_delete"
		}
		if msg.BurnAfterReadingExpireTime.Before(*msg.AutoDeleteExpireTime) {
			return "burn_after_reading"
		}
		return "both"
	}
	if hasBurn {
		return "burn_after_reading"
	}
	return "auto_delete"
}

func (w *AutoDeleteWorker) publishNotification(_ context.Context, messageIDs []string, reason string) {
	notif := broker.NewNotification(broker.TypeMessageAutoDeleted, "", broker.PriorityNormal).
		AddPayloadField("message_ids", messageIDs).
		AddPayloadField("reason", reason)

	if err := w.broker.Publish(notif); err != nil {
		w.log.Warnw("msg", "Failed to publish auto-delete notification", "error", err)
	}
}

func (w *AutoDeleteWorker) StartAsync() {
	go w.Start()
}
