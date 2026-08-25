package grpc

import (
	"context"

	constspkg "flamingo/pkg/consts"

	conversationv1 "flamingo/api/conversation/v1"
	filev1 "flamingo/api/file/v1"
	friendv1 "flamingo/api/friend/v1"
	groupv1 "flamingo/api/group/v1"
	messagev1 "flamingo/api/message/v1"
	pushv1 "flamingo/api/push/v1"
	rtcv1 "flamingo/api/rtc/v1"
	userv1 "flamingo/api/user/v1"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewRtcServiceClient,
	NewConversationServiceClient,
	NewFileServiceClient,
	NewFriendServiceClient,
	NewGroupServiceClient,
	NewMessageServiceClient,
	NewPushServiceClient,
	NewUserServiceClient,
)

func NewRtcServiceClient(r *consul.Registry) rtcv1.RtcServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+constspkg.RtcServiceName),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
			metadata.Client(),
		),
		grpc.WithUnaryInterceptor(ErrorCodeClient()),
	)
	if err != nil {
		panic(err)
	}
	c := rtcv1.NewRtcServiceClient(conn)
	return c
}

func NewConversationServiceClient(r *consul.Registry) conversationv1.ConversationServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+constspkg.ConversationServiceName),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
			metadata.Client(),
		),
		grpc.WithUnaryInterceptor(ErrorCodeClient()),
	)
	if err != nil {
		panic(err)
	}
	c := conversationv1.NewConversationServiceClient(conn)
	return c
}

func NewFileServiceClient(r *consul.Registry) filev1.FileServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+constspkg.FileServiceName),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
			metadata.Client(),
		),
		grpc.WithUnaryInterceptor(ErrorCodeClient()),
	)
	if err != nil {
		panic(err)
	}
	c := filev1.NewFileServiceClient(conn)
	return c
}

func NewFriendServiceClient(r *consul.Registry) friendv1.FriendServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+constspkg.FriendServiceName),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
			metadata.Client(),
		),
		grpc.WithUnaryInterceptor(ErrorCodeClient()),
	)
	if err != nil {
		panic(err)
	}
	c := friendv1.NewFriendServiceClient(conn)
	return c
}

func NewGroupServiceClient(r *consul.Registry) groupv1.GroupServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+constspkg.GroupServiceName),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
			metadata.Client(),
		),
		grpc.WithUnaryInterceptor(ErrorCodeClient()),
	)
	if err != nil {
		panic(err)
	}
	c := groupv1.NewGroupServiceClient(conn)
	return c
}

func NewMessageServiceClient(r *consul.Registry) messagev1.MessageServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+constspkg.MessageServiceName),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
			metadata.Client(),
		),
		grpc.WithUnaryInterceptor(ErrorCodeClient()),
	)
	if err != nil {
		panic(err)
	}
	c := messagev1.NewMessageServiceClient(conn)
	return c
}

func NewPushServiceClient(r *consul.Registry) pushv1.PushServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+constspkg.PushServiceName),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
			metadata.Client(),
		),
		grpc.WithUnaryInterceptor(ErrorCodeClient()),
	)
	if err != nil {
		panic(err)
	}
	c := pushv1.NewPushServiceClient(conn)
	return c
}

func NewUserServiceClient(r *consul.Registry) userv1.UserServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+constspkg.UserServiceName),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
			metadata.Client(),
		),
		grpc.WithUnaryInterceptor(ErrorCodeClient()),
	)
	if err != nil {
		panic(err)
	}
	c := userv1.NewUserServiceClient(conn)
	return c
}
