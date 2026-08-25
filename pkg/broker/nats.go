package broker

import (
	"encoding/json"
	"fmt"
	"time"

	"flamingo/pkg/config"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
)

var ProviderSet = wire.NewSet(NewBroker)

// Broker is the transport-agnostic interface for publishing and subscribing
// to notification events. Each concrete implementation (NATS, Kafka, ...)
// provides its own adapter that translates between this interface and the
// broker-specific primitives (topics, subjects, partitions, ...).
type Broker interface {
	Publish(n *Notification) error
	PublishToUser(userID string, n *Notification) error
	PublishToUsers(userIDs []string, n *Notification) error
	PublishToGroup(groupID string, n *Notification) error
	PublishBroadcast(n *Notification) error
	Subscribe(topic string, handler Handler) (Subscription, error)
}

// natsBroker adapts the NATS client to the Broker interface.
type natsBroker struct {
	nc *nats.Conn
}

// natsSubscription wraps *nats.Subscription behind the transport-agnostic
// Subscription interface.
type natsSubscription struct {
	sub *nats.Subscription
}

func (ns *natsSubscription) Unsubscribe() error {
	return ns.sub.Unsubscribe()
}

// NewBroker creates a broker from the given config. The returned cleanup
// function closes the underlying connection. The factory name is broker-
// agnostic so future implementations (Kafka, ...) can be selected without
// touching the caller.
func NewBroker(cfg config.Broker) (Broker, func(), error) {
	switch cfg.Type {
	case "nats", "":
		return newNatsBroker(cfg)
	default:
		return nil, nil, fmt.Errorf("unsupported broker type: %s", cfg.Type)
	}
}

func newNatsBroker(cfg config.Broker) (Broker, func(), error) {
	nc, err := nats.Connect(cfg.Addr)
	if err != nil {
		return nil, nil, err
	}
	b := &natsBroker{nc: nc}
	return b, func() {
		log.Info("cleanup: closing nats broker")
		if err := nc.Drain(); err != nil {
			log.Errorf("Failed to drain NATS connection: %v", err)
		}
	}, nil
}

// Publish publishes a notification (generic method, requires ToUserID).
func (p *natsBroker) Publish(n *Notification) error {
	if n.ToUserID == "" {
		return fmt.Errorf("toUserID is required for publishing notification")
	}
	return p.PublishToUser(n.ToUserID, n)
}

// PublishToUser publishes a notification to a specific user.
func (p *natsBroker) PublishToUser(userID string, n *Notification) error {
	n.ToUserID = userID
	if n.Timestamp == 0 {
		n.Timestamp = time.Now().Unix()
	}

	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	topic := BuildUserNotificationTopic(n.Type, userID)
	return p.nc.Publish(topic, data)
}

// PublishToUsers batch publishes notifications to multiple users.
func (p *natsBroker) PublishToUsers(userIDs []string, n *Notification) error {
	if n.Timestamp == 0 {
		n.Timestamp = time.Now().Unix()
	}

	for _, userID := range userIDs {
		n.ToUserID = userID
		if err := p.PublishToUser(userID, n); err != nil {
			return err
		}
	}
	return nil
}

// PublishToGroup publishes a group notification.
func (p *natsBroker) PublishToGroup(groupID string, n *Notification) error {
	if n.Timestamp == 0 {
		n.Timestamp = time.Now().Unix()
	}

	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	topic := BuildGroupNotificationTopic(n.Type, groupID)
	return p.nc.Publish(topic, data)
}

// PublishBroadcast broadcasts a notification to all users.
func (p *natsBroker) PublishBroadcast(n *Notification) error {
	if n.Timestamp == 0 {
		n.Timestamp = time.Now().Unix()
	}

	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	topic := BuildBroadcastTopic(n.Type)
	return p.nc.Publish(topic, data)
}

func (p *natsBroker) Subscribe(topic string, handler Handler) (Subscription, error) {
	sub, err := p.nc.Subscribe(topic, func(natsMsg *nats.Msg) {
		msg := &Message{
			Topic: natsMsg.Subject,
			Data:  natsMsg.Data,
		}
		if err := handler(msg); err != nil {
			log.Errorf("failed to handle notification: %v", err)
		}
	})
	if err != nil {
		return nil, err
	}
	return &natsSubscription{sub: sub}, nil
}

// BuildUserNotificationTopic builds a user-notification topic.
// Format: notification.{notificationType}.{user_id}
// Example: notification.friend.request.user-123
//
// NATS uses wildcard subjects (a dotted hierarchy), so "*" can match an
// entire segment. Kafka topics are flat, so a future Kafka adapter maps the
// dotted path onto a single flat topic (e.g. "notification.friend.request")
// and filters by user_id inside the payload.
func BuildUserNotificationTopic(notificationType, userID string) string {
	return fmt.Sprintf("notification.%s.%s", notificationType, userID)
}

// BuildGroupNotificationTopic builds a group-notification topic.
// Format: notification.{notificationType}.{group_id}
// Example: notification.group.member_joined.group-456
func BuildGroupNotificationTopic(notificationType, groupID string) string {
	return fmt.Sprintf("notification.%s.%s", notificationType, groupID)
}

// BuildBroadcastTopic builds a broadcast-notification topic.
// Format: notification.{notificationType}.broadcast
// Example: notification.admin.announcement.broadcast
func BuildBroadcastTopic(notificationType string) string {
	return fmt.Sprintf("notification.%s.broadcast", notificationType)
}
