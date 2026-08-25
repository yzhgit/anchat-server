package broker

import (
	"github.com/google/uuid"
)

// NewNotification creates a new notification.
func NewNotification(notificationType string, fromUserID string, priority Priority) *Notification {
	return &Notification{
		ID:         uuid.New().String(),
		Type:       notificationType,
		FromUserID: fromUserID,
		Priority:   priority,
		Payload:    make(map[string]any),
		Metadata:   make(map[string]any),
	}
}

// WithPayload sets payload.
func (n *Notification) WithPayload(payload map[string]any) *Notification {
	n.Payload = payload
	return n
}

// WithMetadata sets metadata.
func (n *Notification) WithMetadata(metadata map[string]any) *Notification {
	n.Metadata = metadata
	return n
}

// AddPayloadField adds a payload field.
func (n *Notification) AddPayloadField(key string, value any) *Notification {
	if n.Payload == nil {
		n.Payload = make(map[string]any)
	}
	n.Payload[key] = value
	return n
}

// AddMetadataField adds a metadata field.
func (n *Notification) AddMetadataField(key string, value any) *Notification {
	if n.Metadata == nil {
		n.Metadata = make(map[string]any)
	}
	n.Metadata[key] = value
	return n
}
