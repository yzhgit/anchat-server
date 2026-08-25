package server

import (
	"context"

	"flamingo/app/gateway/internal/middleware"
	"flamingo/pkg/auth"
	cfgpkg "flamingo/pkg/config"
	"flamingo/pkg/metrics"

	conversationv1 "flamingo/api/conversation/v1"
	filev1 "flamingo/api/file/v1"
	friendv1 "flamingo/api/friend/v1"
	groupv1 "flamingo/api/group/v1"
	messagev1 "flamingo/api/message/v1"
	userv1 "flamingo/api/user/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/user.v1.UserService/Register"] = struct{}{}
	whiteList["/user.v1.UserService/Login"] = struct{}{}
	whiteList["/user.v1.UserService/RefreshToken"] = struct{}{}
	whiteList["/user.v1.UserService/SendVerificationCode"] = struct{}{}
	whiteList["/user.v1.UserService/ResetPassword"] = struct{}{}
	whiteList["/group.v1.GroupService/RefreshGroupQRCode"] = struct{}{}
	whiteList["/group.v1.GroupService/GetGroupPreviewByQRCode"] = struct{}{}
	whiteList["/metrics"] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

// NewHTTPServer creates a Kratos HTTP server with all service routes and middleware.
func NewHTTPServer(
	c cfgpkg.Server,
	ac cfgpkg.Auth,
	logger log.Logger,
	userProxy userv1.UserServiceHTTPServer,
	friendProxy friendv1.FriendServiceHTTPServer,
	groupProxy groupv1.GroupServiceHTTPServer,
	conversationProxy conversationv1.ConversationServiceHTTPServer,
	messageProxy messagev1.MessageServiceHTTPServer,
	fileProxy filev1.FileServiceHTTPServer,
) *http.Server {
	jwtMiddleware := jwt.Server(
		func(token *jwtv5.Token) (interface{}, error) {
			return []byte(ac.Secret), nil
		},
		jwt.WithSigningMethod(jwtv5.SigningMethodHS256),
		jwt.WithClaims(func() jwtv5.Claims {
			return &auth.Claims{}
		}),
	)

	var opts = []http.ServerOption{
		http.ResponseEncoder(middleware.ResponseEncoder),
		http.ErrorEncoder(middleware.ErrorEncoder),
		http.Middleware(
			recovery.Recovery(),
			metrics.Middleware(),
			middleware.ErrorCodeMiddleware(),
			tracing.Server(),
			logging.Server(logger),
			metadata.Server(),
			selector.Server(jwtMiddleware).Match(NewWhiteListMatcher()).Build(),
			middleware.UserIDMiddleware(),
			ratelimit.Server(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)

	// Register all service HTTP servers — each proxyServer forwards to its gRPC client.
	userv1.RegisterUserServiceHTTPServer(srv, userProxy)
	friendv1.RegisterFriendServiceHTTPServer(srv, friendProxy)
	groupv1.RegisterGroupServiceHTTPServer(srv, groupProxy)
	conversationv1.RegisterConversationServiceHTTPServer(srv, conversationProxy)
	messagev1.RegisterMessageServiceHTTPServer(srv, messageProxy)
	filev1.RegisterFileServiceHTTPServer(srv, fileProxy)

	// Register Prometheus metrics endpoint — bypasses all auth middleware via whitelist.
	srv.Handle("/metrics", metrics.MetricsHandler())

	return srv
}
