package handler

import (
	conversationv1 "flamingo/api/conversation/v1"
	filev1 "flamingo/api/file/v1"
	friendv1 "flamingo/api/friend/v1"
	groupv1 "flamingo/api/group/v1"
	messagev1 "flamingo/api/message/v1"
	userv1 "flamingo/api/user/v1"

	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for gateway proxy handlers.
var ProviderSet = wire.NewSet(
	NewUserProxy,
	NewFriendProxy,
	NewGroupProxy,
	NewConversationProxy,
	NewMessageProxy,
	NewFileProxy,
)

func NewUserProxy(client userv1.UserServiceClient) userv1.UserServiceHTTPServer {
	return &userProxy{client: client}
}

func NewFriendProxy(client friendv1.FriendServiceClient) friendv1.FriendServiceHTTPServer {
	return &friendProxy{client: client}
}

func NewGroupProxy(client groupv1.GroupServiceClient) groupv1.GroupServiceHTTPServer {
	return &groupProxy{client: client}
}

func NewConversationProxy(client conversationv1.ConversationServiceClient) conversationv1.ConversationServiceHTTPServer {
	return &conversationProxy{client: client}
}

func NewMessageProxy(client messagev1.MessageServiceClient) messagev1.MessageServiceHTTPServer {
	return &messageProxy{client: client}
}

func NewFileProxy(client filev1.FileServiceClient) filev1.FileServiceHTTPServer {
	return &fileProxy{client: client}
}
