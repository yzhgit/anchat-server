package handler

import (
	"encoding/json"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
)

// ConnectionManager maintains the registry of active WebSocket connections,
// keyed by userID and deviceID.
type ConnectionManager struct {
	clients map[string]map[string]*WebSocketClient // userID -> deviceID -> client
	mu      sync.RWMutex
	log     *log.Helper
}

// NewConnectionManager creates a ConnectionManager.
func NewConnectionManager(logger log.Logger) *ConnectionManager {
	return &ConnectionManager{
		clients: make(map[string]map[string]*WebSocketClient),
		log:     log.NewHelper(logger),
	}
}

// Register registers a new client; when the same user/device reconnects, the old connection is replaced.
func (m *ConnectionManager) Register(client *WebSocketClient) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[client.UserID]; !exists {
		m.clients[client.UserID] = make(map[string]*WebSocketClient)
	}

	if old, exists := m.clients[client.UserID][client.DeviceID]; exists {
		close(old.Done)
		m.log.Infow("msg", "Replaced existing WebSocket connection", "userID", client.UserID, "deviceID", client.DeviceID)
	}

	m.clients[client.UserID][client.DeviceID] = client
	m.log.Infow("msg", "WebSocket client registered", "userID", client.UserID, "deviceID", client.DeviceID, "deviceCount", len(m.clients[client.UserID]))
}

// Unregister unregisters a client (only if the passed client is still the active one).
func (m *ConnectionManager) Unregister(client *WebSocketClient) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if userClients, exists := m.clients[client.UserID]; exists {
		if current, ok := userClients[client.DeviceID]; ok && current == client {
			delete(userClients, client.DeviceID)
			if len(userClients) == 0 {
				delete(m.clients, client.UserID)
			}
			m.log.Infow("msg", "WebSocket client unregistered", "userID", client.UserID, "deviceID", client.DeviceID)
		}
	}
}

// SendToUser sends a raw message to all online devices of the given user (non-blocking, drops if full).
// Returns true if at least one device received it.
func (m *ConnectionManager) SendToUser(userID string, data []byte) bool {
	m.mu.RLock()
	userClients, exists := m.clients[userID]
	if !exists || len(userClients) == 0 {
		m.mu.RUnlock()
		return false
	}
	type targetClient struct {
		deviceID string
		client   *WebSocketClient
	}
	targets := make([]targetClient, 0, len(userClients))
	for deviceID, client := range userClients {
		targets = append(targets, targetClient{
			deviceID: deviceID,
			client:   client,
		})
	}
	m.mu.RUnlock()

	sent := false
	for _, target := range targets {
		select {
		case target.client.Send <- data:
			sent = true
		default:
			m.log.Warnw("msg", "WebSocket send buffer full, dropping message", "userID", userID, "deviceID", target.deviceID)
		}
	}
	return sent
}

// SendMessageToUser marshals a Message and sends it to the given user.
func (m *ConnectionManager) SendMessageToUser(userID string, msg *Message) bool {
	data, err := json.Marshal(msg)
	if err != nil {
		m.log.Errorw("msg", "Failed to marshal WebSocket message", "error", err)
		return false
	}
	return m.SendToUser(userID, data)
}

// SendNotificationToUser serializes a broker notification as a "notification"
// WebSocket message and pushes it to the user. It is exported so the service
// layer can push notifications without importing this package's Message type.
func (m *ConnectionManager) SendNotificationToUser(userID string, payload json.RawMessage) bool {
	msg := Message{Type: "notification", Payload: payload}
	return m.SendMessageToUser(userID, &msg)
}

// IsOnline reports whether the user has at least one active WebSocket connection.
func (m *ConnectionManager) IsOnline(userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	userClients, exists := m.clients[userID]
	return exists && len(userClients) > 0
}

// OnlineCount returns the number of users with at least one active connection.
func (m *ConnectionManager) OnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}
