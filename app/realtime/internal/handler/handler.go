package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	conversationv1 "flamingo/api/conversation/v1"
	filev1 "flamingo/api/file/v1"
	friendv1 "flamingo/api/friend/v1"
	groupv1 "flamingo/api/group/v1"
	messagev1 "flamingo/api/message/v1"
	pushv1 "flamingo/api/push/v1"
	rtcv1 "flamingo/api/rtc/v1"
	userv1 "flamingo/api/user/v1"

	"flamingo/pkg/auth"

	"github.com/go-kratos/kratos/v2/log"
	gorillaws "github.com/gorilla/websocket"
)

var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return validateOrigin(r)
	},
}

// allowedOrigins returns the set of origins permitted to open a WebSocket.
// Configure via env var ALLOWED_ORIGINS (comma separated) for deployment.
// Keep empty in production to block all non-localhost origins.
func allowedOrigins() []string {
	if v := strings.TrimSpace(getEnv("ALLOWED_ORIGINS")); v != "" {
		return strings.Split(v, ",")
	}
	return nil
}

func validateOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true // server-to-server / curl / no browser
	}
	origins := allowedOrigins()
	if len(origins) == 0 {
		// Empty whitelist -> reject any explicit browser-origin WebSocket.
		return false
	}
	for _, o := range origins {
		if strings.EqualFold(strings.TrimSpace(o), origin) {
			return true
		}
	}
	return false
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func extractToken(r *http.Request) string {
	// Prefer Authorization header (never put JWT in URL -- it leaks to
	// browser history, server access logs, proxies and Referer headers).
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if token, ok := strings.CutPrefix(auth, "Bearer "); ok {
		return strings.TrimSpace(token)
	}
	// Reject tokens in URL query to prevent log/history leakage.
	_ = r.URL.Query().Get("token")
	return ""
}

// Subscriber is the NATS notification subscription interface. The concrete
// implementation lives in the service layer, which keeps handler free of
// NATS-specific imports.
type Subscriber interface {
	SubscribeUser(userID string) error
	UnsubscribeUser(userID string)
}

// WebSocketHandler handles WebSocket connections: auth, upgrade, and message routing.
type WebSocketHandler struct {
	userClient         userv1.UserServiceClient
	friendClient       friendv1.FriendServiceClient
	groupClient        groupv1.GroupServiceClient
	fileClient         filev1.FileServiceClient
	messageClient      messagev1.MessageServiceClient
	conversationClient conversationv1.ConversationServiceClient
	rtcClient          rtcv1.RtcServiceClient
	pushClient         pushv1.PushServiceClient

	jwtManager  *auth.JWTManager
	connManager *ConnectionManager
	subscriber  Subscriber
	log         *log.Helper
}

// NewWebSocketHandler creates a WebSocketHandler.
func NewWebSocketHandler(
	logger log.Logger,
	userClient userv1.UserServiceClient,
	friendClient friendv1.FriendServiceClient,
	groupClient groupv1.GroupServiceClient,
	fileClient filev1.FileServiceClient,
	messageClient messagev1.MessageServiceClient,
	conversationClient conversationv1.ConversationServiceClient,
	rtcClient rtcv1.RtcServiceClient,
	pushClient pushv1.PushServiceClient,
	jwtManager *auth.JWTManager,
	connManager *ConnectionManager,
	subscriber Subscriber,
) *WebSocketHandler {
	return &WebSocketHandler{
		userClient:         userClient,
		friendClient:       friendClient,
		groupClient:        groupClient,
		fileClient:         fileClient,
		messageClient:      messageClient,
		conversationClient: conversationClient,
		rtcClient:          rtcClient,
		pushClient:         pushClient,
		jwtManager:         jwtManager,
		connManager:        connManager,
		subscriber:         subscriber,
		log:                log.NewHelper(logger),
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		http.Error(w, "unauthorized: token is required", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwtManager.ValidateAccessToken(token)
	if err != nil {
		h.log.Warnw("msg", "WebSocket upgrade rejected: invalid token", "error", err)
		http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	deviceID := claims.DeviceID

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Errorw("msg", "Failed to upgrade WebSocket", "error", err)
		return
	}

	wsClient := NewWebSocketClient(userID, deviceID, conn, h.connManager, h.log.Logger())
	h.connManager.Register(wsClient)

	if err := h.subscriber.SubscribeUser(userID); err != nil {
		h.log.Errorw("msg", "Failed to subscribe user", "userID", userID, "error", err)
	}

	h.log.Infow("msg", "WebSocket client connected", "userID", userID, "deviceID", deviceID, "onlineCount", h.connManager.OnlineCount())

	go wsClient.WritePump()

	// ReadPump blocks until connection disconnects.
	wsClient.ReadPump(h.handleClientMessage)

	// After connection disconnects, only unsubscribe from NATS when the user is truly offline.
	if !h.connManager.IsOnline(userID) {
		h.subscriber.UnsubscribeUser(userID)
	}

	h.log.Infow("msg", "WebSocket client disconnected", "userID", userID, "onlineCount", h.connManager.OnlineCount())
}

// handleClientMessage dispatches incoming WebSocket messages by type.
func (h *WebSocketHandler) handleClientMessage(c *WebSocketClient, msg *Message) {
	switch msg.Type {
	case "ping":
		h.connManager.SendMessageToUser(c.UserID, &Message{Type: "pong"})

	case "message.send":
		h.handleSendMessage(c, msg.Payload)

	case "message.typing":
		h.handleSendTyping(c, msg.Payload)

	default:
		h.log.Debugw("msg", "Unknown WebSocket message type", "type", msg.Type, "userID", c.UserID)
	}
}

// sendMessagePayload is the payload for the client's message.send request.
type sendMessagePayload struct {
	ConversationID string   `json:"conversation_id"`
	ContentType    int32    `json:"content_type"`
	Content        string   `json:"content"`
	ReplyTo        string   `json:"reply_to,omitempty"`
	AtUsers        []string `json:"at_users,omitempty"`
	LocalID        string   `json:"local_id,omitempty"`
}

type sendTypingPayload struct {
	ConversationID string `json:"conversation_id"`
	Typing         *bool  `json:"typing"`
	TTLSeconds     *int32 `json:"ttl_seconds,omitempty"`
	ClientTs       *int64 `json:"client_ts,omitempty"`
}

// sendMessageResult is the server's confirmation after a successful message send.
type sendMessageResult struct {
	MessageID string `json:"message_id"`
	Sequence  int64  `json:"sequence"`
	Timestamp int64  `json:"timestamp"`
	LocalID   string `json:"local_id,omitempty"`
}

type sendMessageError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	LocalID string `json:"local_id,omitempty"`
}

type sendTypingError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// handleSendMessage parses a message.send payload and forwards it via gRPC.
func (h *WebSocketHandler) handleSendMessage(c *WebSocketClient, payload json.RawMessage) {
	var req sendMessagePayload
	if err := json.Unmarshal(payload, &req); err != nil {
		h.log.Warnw("msg", "Invalid message.send payload", "userID", c.UserID, "error", err)
		return
	}

	grpcReq := &messagev1.SendMessageRequest{
		ConversationId: req.ConversationID,
		ContentType:    messagev1.ContentType(req.ContentType),
		Content:        req.Content,
		LocalId:        req.LocalID,
	}
	if req.ReplyTo != "" {
		grpcReq.ReplyTo = &messagev1.Message{MessageId: req.ReplyTo}
	}
	if len(req.AtUsers) > 0 {
		grpcReq.AtUsers = req.AtUsers
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.messageClient.SendMessage(ctx, grpcReq)
	if err != nil {
		h.log.Errorw("msg", "Failed to send message via gRPC", "userID", c.UserID, "error", err)

		wsErr := &sendMessageError{
			Code:    "send_failed",
			Message: err.Error(),
			LocalID: req.LocalID,
		}
		errData, _ := json.Marshal(wsErr)
		h.connManager.SendMessageToUser(c.UserID, &Message{
			Type:    "message.error",
			Payload: json.RawMessage(errData),
		})
		return
	}

	var ts int64
	if resp.Message != nil && resp.Message.Timestamp != nil {
		ts = resp.Message.Timestamp.Seconds
	}

	result := &sendMessageResult{
		MessageID: resp.Message.MessageId,
		Sequence:  resp.Message.Seq,
		Timestamp: ts,
		LocalID:   req.LocalID,
	}

	resultData, _ := json.Marshal(result)
	h.connManager.SendMessageToUser(c.UserID, &Message{
		Type:    "message.sent",
		Payload: json.RawMessage(resultData),
	})
}

func (h *WebSocketHandler) handleSendTyping(c *WebSocketClient, payload json.RawMessage) {
	var req sendTypingPayload
	if err := json.Unmarshal(payload, &req); err != nil {
		h.log.Warnw("msg", "Invalid message.typing payload", "userID", c.UserID, "error", err)
		return
	}
	if req.ConversationID == "" || req.Typing == nil {
		h.log.Warnw("msg", "Invalid message.typing payload fields", "userID", c.UserID, "conversationID", req.ConversationID)
		return
	}

	grpcReq := &messagev1.SendTypingRequest{
		ConversationId: req.ConversationID,
		Typing:         *req.Typing,
	}
	if req.TTLSeconds != nil {
		grpcReq.TtlSeconds = *req.TTLSeconds
	}
	if c.DeviceID != "" {
		grpcReq.DeviceId = c.DeviceID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := h.messageClient.SendTyping(ctx, grpcReq); err != nil {
		h.log.Errorw("msg", "Failed to send typing via gRPC", "userID", c.UserID, "error", err)

		wsErr := &sendTypingError{
			Code:    "typing_failed",
			Message: err.Error(),
		}
		errData, _ := json.Marshal(wsErr)
		h.connManager.SendMessageToUser(c.UserID, &Message{
			Type:    "message.error",
			Payload: json.RawMessage(errData),
		})
	}
}
