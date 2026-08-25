package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"flamingo/app/push/internal/handler"
	"flamingo/app/push/internal/model"
	"flamingo/pkg/broker"

	"github.com/go-kratos/kratos/v2/log"
)

// pushTypeTitle default titles for each push type
var pushTypeTitle = map[model.PushType]string{
	model.PushTypeMessageNew:     "New Message",
	model.PushTypeMessageMention: "You were mentioned",
	model.PushTypeFriendRequest:  "Friend Request",
	model.PushTypeGroupInvited:   "Group Invitation",
	model.PushTypeCallInvite:     "Incoming Call",
}

var notificationPushTypeMap = map[string]model.PushType{
	broker.TypeMessageNew:        model.PushTypeMessageNew,
	broker.TypeMessageMentioned:  model.PushTypeMessageMention,
	broker.TypeFriendRequest:     model.PushTypeFriendRequest,
	broker.TypeGroupInvited:      model.PushTypeGroupInvited,
	broker.TypeLiveKitCallInvite: model.PushTypeCallInvite,
}

// handleNotification handles notification events, deciding whether to push.
//
// NOTE: currently pushes for every matching notification regardless of online
// status. Realtime delivers the same events to online users via WebSocket and
// JPush delivers to offline users, so an online user can receive a duplicate.
// To deduplicate, add a user OnlineStatus RPC (e.g. GetUserOnlineStatus) and
// skip SendPush when the user is online.
func handleNotification(svc handler.PushService, logger log.Logger, msg *broker.Message) {
	helper := log.NewHelper(logger)

	var notif broker.Notification
	if err := json.Unmarshal(msg.Data, &notif); err != nil {
		helper.Errorw("msg", "failed to unmarshal notification", "error", err)
		return
	}

	helper.Infow("msg", "handling push notification", "type", notif.Type, "to_user_id", notif.ToUserID)

	pushType, title, ok := buildPushContent(notif)
	if !ok {
		return // unsupported push type
	}

	if notif.ToUserID == "" {
		return
	}

	content := extractContent(notif)
	extras := extractExtras(notif)

	svc.SendPush(context.Background(), //nolint:errcheck
		[]string{notif.ToUserID},
		title, content, pushType,
		extras,
	)
}

// buildPushContent maps notification type to push type and title.
func buildPushContent(notif broker.Notification) (model.PushType, string, bool) {
	pushType, ok := notificationPushTypeMap[notif.Type]
	if !ok {
		return model.PushTypeUnspecified, "", false
	}

	title, exists := pushTypeTitle[pushType]
	if !exists {
		title = "New Notification"
	}
	return pushType, title, true
}

// extractContent extracts push body from notification Payload.
func extractContent(notif broker.Notification) string {
	if notif.Payload == nil {
		return ""
	}
	for _, key := range []string{"content", "body", "text"} {
		if v, ok := notif.Payload[key]; ok {
			if str, ok := v.(string); ok && str != "" {
				if len([]rune(str)) > 80 {
					str = string([]rune(str)[:80]) + "..."
				}
				return str
			}
		}
	}
	return ""
}

// extractExtras extracts string extra data from notification Payload.
func extractExtras(notif broker.Notification) map[string]string {
	extras := map[string]string{
		"notification_type": notif.Type,
	}
	parts := strings.SplitN(notif.Type, ".", 2)
	if len(parts) > 0 {
		extras["service"] = parts[0]
	}
	for k, v := range notif.Payload {
		if str, ok := v.(string); ok {
			extras[k] = str
			continue
		}
		switch val := v.(type) {
		case int:
			extras[k] = fmt.Sprintf("%d", val)
		case int32:
			extras[k] = fmt.Sprintf("%d", val)
		case int64:
			extras[k] = fmt.Sprintf("%d", val)
		case float32:
			extras[k] = fmt.Sprintf("%v", val)
		case float64:
			extras[k] = fmt.Sprintf("%v", val)
		case bool:
			extras[k] = fmt.Sprintf("%t", val)
		}
	}
	return extras
}
