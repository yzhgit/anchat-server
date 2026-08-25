package handler

import (
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	gorillaws "github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second // write operation timeout
	pongWait       = 60 * time.Second // timeout waiting for pong response
	pingPeriod     = 54 * time.Second // interval to send ping (less than pongWait)
	maxMessageSize = 65536            // max message size (64KB)
)

// Message WebSocket message wire format.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// WebSocketClient represents a single WebSocket connection.
type WebSocketClient struct {
	UserID   string
	DeviceID string
	Conn     *gorillaws.Conn
	Send     chan []byte   // message queue to send
	Done     chan struct{} // close signal (closed when replaced by new connection)
	manager  *ConnectionManager
	log      *log.Helper
}

// NewWebSocketClient creates a new WebSocketClient.
func NewWebSocketClient(userID, deviceID string, conn *gorillaws.Conn, manager *ConnectionManager, logger log.Logger) *WebSocketClient {
	return &WebSocketClient{
		UserID:   userID,
		DeviceID: deviceID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Done:     make(chan struct{}),
		manager:  manager,
		log:      log.NewHelper(logger),
	}
}

// WritePump handles sending messages to client (with heartbeat), runs in separate goroutine.
func (c *WebSocketClient) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Send channel is closed
				_ = c.Conn.WriteMessage(gorillaws.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(gorillaws.TextMessage, message); err != nil {
				c.log.Warnw("msg", "WebSocket write error", "userID", c.UserID, "error", err)
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(gorillaws.PingMessage, nil); err != nil {
				return
			}
		case <-c.Done:
			// replaced by new connection, close old connection
			_ = c.Conn.WriteMessage(gorillaws.CloseMessage, gorillaws.FormatCloseMessage(
				gorillaws.CloseNormalClosure, "replaced by new connection"))
			return
		}
	}
}

// ReadPump handles receiving messages from client, blocks until connection disconnects.
func (c *WebSocketClient) ReadPump(onMessage func(client *WebSocketClient, msg *Message)) {
	defer func() {
		c.manager.Unregister(c)
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if gorillaws.IsUnexpectedCloseError(err, gorillaws.CloseGoingAway, gorillaws.CloseAbnormalClosure) {
				c.log.Warnw("msg", "WebSocket unexpected close", "userID", c.UserID, "error", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.log.Warnw("msg", "Failed to parse WebSocket message", "userID", c.UserID, "error", err)
			continue
		}

		onMessage(c, &msg)
	}
}
