package handler

import (
	"github.com/anychat/server/internal/gateway/client"
	gwmiddleware "github.com/anychat/server/internal/gateway/middleware"
	gwnotification "github.com/anychat/server/internal/gateway/notification"
	"github.com/anychat/server/internal/gateway/websocket"
	"github.com/anychat/server/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes
func RegisterRoutes(r *gin.Engine, clientManager *client.Manager, jwtManager *jwt.Manager,
	wsManager *websocket.Manager, subscriber *gwnotification.Subscriber) {
	// create handlers
	authHandler := NewAuthHandler(clientManager)
	userHandler := NewUserHandler(clientManager)
	friendHandler := NewFriendHandler(clientManager)
	groupHandler := NewGroupHandler(clientManager)
	fileHandler := NewFileHandler(clientManager)
	logHandler := NewLogHandler(clientManager)
	messageHandler := NewMessageHandler(clientManager)
	wsHandler := NewWSHandler(clientManager, jwtManager, wsManager, subscriber)
	conversationHandler := NewConversationHandler(clientManager)
	syncHandler := NewSyncHandler(clientManager)
	callingHandler := NewCallingHandler(clientManager)
	versionHandler := NewVersionHandler(clientManager)
	// API v1
	v1 := r.Group("/api/v1")
	{
		// WebSocket endpoint (token authenticated via query parameter)
		v1.GET("/ws", wsHandler.HandleWebSocket)
		// public routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/send-code", authHandler.SendCode)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/password/reset", authHandler.ResetPassword)
		}

		// group QR code preview (no auth required)
		v1.GET("/groups/preview", groupHandler.GetGroupPreviewByQRCode)
		v1.GET("/group/preview", groupHandler.GetGroupPreviewByQRCode) // compatible with singular path in design doc

		// routes requiring authentication
		authorized := v1.Group("")
		authorized.Use(gwmiddleware.JWTAuth(jwtManager))
		{
			// Auth routes
			authGroup := authorized.Group("/auth")
			{
				authGroup.POST("/logout", authHandler.Logout)
				authGroup.POST("/password/change", authHandler.ChangePassword)
			}

			// Version routes (client version check - public)
			versions := v1.Group("/versions")
			{
				versions.GET("/check", versionHandler.CheckVersion)
				versions.GET("/latest", versionHandler.GetLatestVersion)
				versions.GET("/list", versionHandler.ListVersions)
			}

			// Version routes requiring auth
			authorizedVersions := authorized.Group("/versions")
			{
				authorizedVersions.POST("/report", versionHandler.ReportVersion)
			}

			// User routes
			users := authorized.Group("/users")
			{
				// personal profile
				users.GET("/me", userHandler.GetProfile)
				users.PUT("/me", userHandler.UpdateProfile)
				users.POST("/me/phone/bind", userHandler.BindPhone)
				users.POST("/me/phone/change", userHandler.ChangePhone)
				users.POST("/me/email/bind", userHandler.BindEmail)
				users.POST("/me/email/change", userHandler.ChangeEmail)

				// user search
				users.GET("/:user_id", userHandler.GetUserInfo)
				users.GET("/search", userHandler.SearchUsers)

				// settings
				users.GET("/me/settings", userHandler.GetSettings)
				users.PUT("/me/settings", userHandler.UpdateSettings)

				// QR code
				users.POST("/me/qrcode/refresh", userHandler.RefreshQRCode)
				users.GET("/qrcode", userHandler.GetUserByQRCode)

				// push token
				users.POST("/me/push-token", userHandler.UpdatePushToken)
			}

			// Friend routes
			friends := authorized.Group("/friends")
			{
				// friend list
				friends.GET("", friendHandler.GetFriends)

				// friend requests
				friends.GET("/requests", friendHandler.GetFriendRequests)
				friends.POST("/requests", friendHandler.SendFriendRequest)
				friends.PUT("/requests/:id", friendHandler.HandleFriendRequest)

				// friend operations
				friends.DELETE("/:id", friendHandler.DeleteFriend)
				friends.PUT("/:id/remark", friendHandler.UpdateRemark)

				// blacklist
				friends.GET("/blacklist", friendHandler.GetBlacklist)
				friends.POST("/blacklist", friendHandler.AddToBlacklist)
				friends.DELETE("/blacklist/:id", friendHandler.RemoveFromBlacklist)
			}

			// Group routes
			groups := authorized.Group("/groups")
			{
				// group management
				groups.POST("", groupHandler.CreateGroup)
				groups.GET("", groupHandler.GetMyGroups)
				groups.POST("/join-by-qrcode", groupHandler.JoinGroupByQRCode)
				groups.GET("/:id", groupHandler.GetGroupInfo)
				groups.PUT("/:id", groupHandler.UpdateGroup)
				groups.DELETE("/:id", groupHandler.DissolveGroup)

				// member management
				groups.GET("/:id/members", groupHandler.GetGroupMembers)
				groups.POST("/:id/members", groupHandler.InviteMembers)
				groups.DELETE("/:id/members/:user_id", groupHandler.RemoveMember)
				groups.PUT("/:id/members/:user_id/mute", groupHandler.MuteMember)
				groups.DELETE("/:id/members/:user_id/mute", groupHandler.UnmuteMember)
				groups.PUT("/:id/members/:user_id/role", groupHandler.UpdateMemberRole)
				groups.PUT("/:id/nickname", groupHandler.UpdateMemberNickname)
				groups.PUT("/:id/remark", groupHandler.UpdateMemberRemark)
				groups.POST("/:id/quit", groupHandler.QuitGroup)
				groups.POST("/:id/transfer", groupHandler.TransferOwnership)
				groups.GET("/:id/settings", groupHandler.GetGroupSettings)
				groups.PUT("/:id/settings", groupHandler.UpdateGroupSettings)
				groups.PUT("/:id/mute", groupHandler.SetGroupMute)
				groups.POST("/:id/pin", groupHandler.PinGroupMessage)
				groups.DELETE("/:id/pin/:message_id", groupHandler.UnpinGroupMessage)
				groups.GET("/:id/pins", groupHandler.GetPinnedMessages)

				// QR code
				groups.GET("/:id/qrcode", groupHandler.GetGroupQRCode)
				groups.POST("/:id/qrcode/refresh", groupHandler.RefreshGroupQRCode)

				// join requests
				groups.POST("/:id/join", groupHandler.JoinGroup)
				groups.GET("/:id/requests", groupHandler.GetJoinRequests)
				groups.PUT("/:id/requests/:requestId", groupHandler.HandleJoinRequest)
			}

			// File routes
			files := authorized.Group("/files")
			{
				files.POST("/upload-token", fileHandler.GenerateUploadToken)
				files.POST("/:fileId/complete", fileHandler.CompleteUpload)
				files.GET("/:fileId/download", fileHandler.GenerateDownloadURL)
				files.GET("/:fileId", fileHandler.GetFileInfo)
				files.DELETE("/:fileId", fileHandler.DeleteFile)
				files.GET("", fileHandler.ListFiles)
			}

			// Log routes
			logs := authorized.Group("/logs")
			{
				logs.POST("/upload", logHandler.UploadLog)
				logs.POST("/complete", logHandler.CompleteUpload)
				logs.GET("", logHandler.ListLogs)
				logs.GET("/:log_id/download", logHandler.DownloadLog)
				logs.DELETE("/:log_id", logHandler.DeleteLog)
			}

			// Message routes
			messages := authorized.Group("/messages")
			{
				messages.POST("", messageHandler.SendMessage)
				messages.GET("/search", messageHandler.SearchMessages)
				messages.GET("/:message_id", messageHandler.GetMessageByID)
				messages.POST("/read-triggers", messageHandler.AckReadTriggers)
				messages.POST("/recall", messageHandler.RecallMessage)
				messages.DELETE("/:message_id", messageHandler.DeleteMessage)
			}

			// Conversation routes
			conversations := authorized.Group("/conversations")
			{
				conversations.GET("", conversationHandler.GetConversations)
				conversations.GET("/unread/total", conversationHandler.GetTotalUnread)
				conversations.GET("/:conversation_id", conversationHandler.GetConversation)
				conversations.GET("/:conversation_id/messages/before", messageHandler.GetMessagesBefore)
				conversations.GET("/:conversation_id/messages/after", messageHandler.GetMessagesAfter)
				conversations.GET("/:conversation_id/messages/around-anchor", messageHandler.GetMessagesAroundAnchor)
				conversations.GET("/:conversation_id/messages/first-unread-anchor", messageHandler.GetFirstUnreadAnchor)
				conversations.GET("/:conversation_id/messages/unread-count", conversationHandler.GetMessageUnreadCount)
				conversations.GET("/:conversation_id/messages/read-receipts", conversationHandler.GetMessageReadReceipts)
				conversations.GET("/:conversation_id/messages/sequence", conversationHandler.GetMessageSequence)
				conversations.POST("/:conversation_id/messages/read", conversationHandler.MarkMessagesRead)
				conversations.DELETE("/:conversation_id", conversationHandler.DeleteConversation)
				conversations.PUT("/:conversation_id/pin", conversationHandler.SetPinned)
				conversations.PUT("/:conversation_id/mute", conversationHandler.SetMuted)
				conversations.PUT("/:conversation_id/burn", conversationHandler.SetBurnAfterReading)
				conversations.PUT("/:conversation_id/auto_delete", conversationHandler.SetAutoDelete)
				conversations.POST("/:conversation_id/read-all", conversationHandler.MarkRead)
			}

			// Sync routes
			sync := authorized.Group("/sync")
			{
				sync.POST("", syncHandler.Sync)
				sync.POST("/messages", syncHandler.SyncMessages)
			}

			// Calling routes (recommended)
			calling := authorized.Group("/calling")
			{
				// one-on-one calls
				calling.POST("/calls", callingHandler.InitiateCall)
				calling.GET("/calls", callingHandler.ListCallLogs)
				calling.GET("/calls/:call_id", callingHandler.GetCallSession)
				calling.POST("/calls/:call_id/join", callingHandler.JoinCall)
				calling.POST("/calls/:call_id/reject", callingHandler.RejectCall)
				calling.POST("/calls/:call_id/end", callingHandler.EndCall)

				// meeting rooms
				calling.POST("/meetings", callingHandler.CreateMeeting)
				calling.GET("/meetings", callingHandler.ListMeetings)
				calling.GET("/meetings/:room_id", callingHandler.GetMeeting)
				calling.POST("/meetings/:room_id/join", callingHandler.JoinMeeting)
				calling.POST("/meetings/:room_id/end", callingHandler.EndMeeting)
			}
		}
	}

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "gateway-service",
		})
	})
}
