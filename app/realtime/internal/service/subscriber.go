package service

import (
	"encoding/json"
	"fmt"
	"sync"

	"flamingo/app/realtime/internal/handler"
	"flamingo/pkg/broker"

	"github.com/go-kratos/kratos/v2/log"
)

// NotificationSubscriber subscribes to per-user notification topics and pushes
// them to the corresponding WebSocket client(s). It implements the
// handler.Subscriber interface so the handler layer stays free of broker
// imports.
type NotificationSubscriber struct {
	msgBroker broker.Broker // message broker for subscribing/publishing
	manager   *handler.ConnectionManager
	subs      map[string]broker.Subscription // userID -> subscription
	mu        sync.RWMutex
	log       *log.Helper
}

// NewNotificationSubscriber creates a NotificationSubscriber.
func NewNotificationSubscriber(logger log.Logger, msgBroker broker.Broker, manager *handler.ConnectionManager) handler.Subscriber {
	return &NotificationSubscriber{
		msgBroker: msgBroker,
		manager:   manager,
		subs:      make(map[string]broker.Subscription),
		log:       log.NewHelper(logger),
	}
}

// SubscribeUser subscribes to notifications for the user (idempotent).
// Topic: notification.*.*.{userID} -- matches all service notifications for this user.
func (s *NotificationSubscriber) SubscribeUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.subs[userID]; exists {
		return nil
	}

	topic := fmt.Sprintf("notification.*.*.%s", userID)
	sub, err := s.msgBroker.Subscribe(topic, func(msg *broker.Message) error {
		s.handleNotification(userID, msg.Data)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe for user %s: %w", userID, err)
	}

	s.subs[userID] = sub
	s.log.Infow("msg", "Subscribed to user notifications", "userID", userID, "topic", topic)
	return nil
}

// UnsubscribeUser removes the user's notification subscription.
func (s *NotificationSubscriber) UnsubscribeUser(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sub, exists := s.subs[userID]; exists {
		if err := sub.Unsubscribe(); err != nil {
			s.log.Warnw("msg", "Failed to unsubscribe", "userID", userID, "error", err)
		}
		delete(s.subs, userID)
		s.log.Infow("msg", "Unsubscribed from user notifications", "userID", userID)
	}
}

// handleNotification deserializes a NATS notification and pushes it to the user's WebSocket.
func (s *NotificationSubscriber) handleNotification(userID string, data []byte) {
	var notif broker.Notification
	if err := json.Unmarshal(data, &notif); err != nil {
		return
	}

	payload, err := json.Marshal(&notif)
	if err != nil {
		return
	}

	if !s.manager.SendNotificationToUser(userID, payload) {
		s.log.Debugw("msg", "User not connected, notification dropped", "userID", userID, "type", notif.Type)
	}
}
