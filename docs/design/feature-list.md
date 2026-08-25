# AnyChat Server 功能列表

> 用于跟踪开发进度的功能清单

---

## 1. Auth Service (认证服务)

> 已合并：认证功能已整合至 User Service 中，不再作为独立服务维护。相关HTTP接口和WebSocket通知已归入下方的 "User Service (用户与认证服务)" 章节。
>
> WebSocket连接建立由 Realtime Service 统一处理，参见 Realtime Service 章节。

### HTTP接口

> 已合并至 User Service (用户与认证服务)

### WebSocket接口

> WebSocket连接由 Realtime Service 统一管理

### WebSocket通知

> 认证相关通知已合并至 User Service (用户与认证服务)

---

## 2. User Service (用户与认证服务)

> 已合并原 Auth Service 的认证、注册、设备管理等接口。

### HTTP接口

| 接口路由 | 功能 | 完成状态 |
|---------|------|----------|
| POST /api/v1/users/send-code | 发送验证码 | ✅ 完成 |
| POST /api/v1/users/register | 用户注册 | ✅ 完成 |
| POST /api/v1/users/login | 用户登录 | ✅ 完成 |
| POST /api/v1/users/refresh | 刷新Token | ✅ 完成 |
| POST /api/v1/users/logout | 用户登出 | ✅ 完成 |
| POST /api/v1/users/password/change | 修改密码 | ✅ 完成 |
| POST /api/v1/users/password/reset | 重置密码 | ✅ 完成 |
| POST /api/v1/users/device/list | 设备列表查询 | ✅ 完成 |
| POST /api/v1/users/device/logout | 设备下线 | ✅ 完成 |
| GET /api/v1/users/me | 获取个人资料 | ✅ 完成 |
| PUT /api/v1/users/me | 更新个人资料 | ✅ 完成 |
| GET /api/v1/users/:userId | 获取指定用户信息 | ✅ 完成 |
| GET /api/v1/users/search | 搜索用户 | ✅ 完成 |
| POST /api/v1/users/me/phone/bind | 绑定手机号 | ✅ 完成 |
| POST /api/v1/users/me/phone/change | 更换手机号 | ✅ 完成 |
| POST /api/v1/users/me/email/bind | 绑定邮箱 | ✅ 完成 |
| POST /api/v1/users/me/email/change | 更换邮箱 | ✅ 完成 |
| GET /api/v1/users/me/settings | 获取用户设置 | ✅ 完成 |
| PUT /api/v1/users/me/settings | 更新用户设置 | ✅ 完成 |
| POST /api/v1/users/me/qrcode/refresh | 刷新二维码 | ✅ 完成 |
| GET /api/v1/users/qrcode | 扫码获取用户 | ✅ 完成 |
| POST /api/v1/users/me/push-token | 更新推送Token | ✅ 完成 |

### WebSocket接口

| 接口 | 功能 | 完成状态 |
|------|------|----------|
| 通过Gateway订阅 | 用户状态变更推送 | ✅ 完成 |

### WebSocket通知

| 通知类型 | 主题 | 完成状态 |
|---------|------|----------|
| 多端互踢通知 | notification.auth.force_logout.{user_id} | ✅ 完成 |
| 异常登录提醒 | notification.auth.unusual_login.{user_id} | ✅ 完成 |
| 密码修改通知 | notification.auth.password_changed.{user_id} | ✅ 完成 |
| 用户资料更新通知 | notification.user.profile_updated.{user_id} | ✅ 完成 |
| 好友资料变更通知 | notification.user.friend_profile_changed.{user_id} | ✅ 完成 |
| 在线状态变更通知 | notification.user.status_changed.{user_id} | ✅ 完成 |

---

## 3. Friend Service (好友服务)

### HTTP接口

| 接口路由 | 功能 | 完成状态 |
|---------|------|----------|
| GET /api/v1/friends | 获取好友列表 | ✅ 完成 |
| GET /api/v1/friends/requests | 获取好友申请列表 | ✅ 完成 |
| POST /api/v1/friends/requests | 发送好友申请 | ✅ 完成 |
| PUT /api/v1/friends/requests/:id | 处理好友申请 | ✅ 完成 |
| DELETE /api/v1/friends/:id | 删除好友 | ✅ 完成 |
| PUT /api/v1/friends/:id/remark | 修改好友备注 | ✅ 完成 |
| GET /api/v1/friends/blacklist | 获取黑名单 | ✅ 完成 |
| POST /api/v1/friends/blacklist | 添加到黑名单 | ✅ 完成 |
| DELETE /api/v1/friends/blacklist/:id | 从黑名单移除 | ✅ 完成 |

### WebSocket接口

| 接口 | 功能 | 完成状态 |
|------|------|----------|
| 通过Gateway订阅 | 好友变更实时推送 | ✅ 完成 |

### WebSocket通知

| 通知类型 | 主题 | 完成状态 |
|---------|------|----------|
| 好友请求通知 | notification.friend.request.{to_user_id} | ✅ 完成 |
| 好友请求处理结果通知 | notification.friend.request_handled.{from_user_id} | ✅ 完成 |
| 好友删除通知 | notification.friend.deleted.{user_id} | ✅ 完成 |
| 好友备注修改同步 | notification.friend.remark_updated.{user_id} | ✅ 完成 |
| 黑名单变更通知 | notification.friend.blacklist_changed.{user_id} | ✅ 完成 |

---

## 4. Group Service (群组服务)

### HTTP接口

| 接口路由 | 功能 | 完成状态 |
|---------|------|----------|
| POST /api/v1/groups | 创建群组 | ✅ 完成 |
| GET /api/v1/groups | 获取我的群组列表 | ✅ 完成 |
| GET /api/v1/groups/:id | 获取群组信息 | ✅ 完成 |
| PUT /api/v1/groups/:id | 更新群组信息 | ✅ 完成 |
| DELETE /api/v1/groups/:id | 解散群组 | ✅ 完成 |
| GET /api/v1/groups/:id/members | 获取群成员列表 | ✅ 完成 |
| POST /api/v1/groups/:id/members | 邀请成员入群 | ✅ 完成 |
| DELETE /api/v1/groups/:id/members/:userId | 移除群成员 | ✅ 完成 |
| PUT /api/v1/groups/:id/members/:userId/role | 更新成员角色 | ✅ 完成 |
| PUT /api/v1/groups/:id/nickname | 修改群昵称 | ✅ 完成 |
| POST /api/v1/groups/:id/quit | 退出群组 | ✅ 完成 |
| POST /api/v1/groups/:id/transfer | 转让群主 | ✅ 完成 |
| POST /api/v1/groups/:id/join | 申请入群 | ✅ 完成 |
| GET /api/v1/groups/:id/requests | 获取入群申请列表 | ✅ 完成 |
| PUT /api/v1/groups/:id/requests/:requestId | 处理入群申请 | ✅ 完成 |
| GET /api/v1/groups/:id/qrcode | 获取群二维码 | ✅ 完成 |
| POST /api/v1/groups/:id/qrcode/refresh | 刷新群二维码 | ✅ 完成 |

### WebSocket接口

| 接口 | 功能 | 完成状态 |
|------|------|----------|
| 通过Gateway订阅 | 群组变更实时推送 | ✅ 完成 |

### WebSocket通知

| 通知类型 | 主题 | 完成状态 |
|---------|------|----------|
| 群组邀请通知 | notification.group.invited.{user_id} | ✅ 完成 |
| 成员加入通知 | notification.group.member_joined.{group_id} | ✅ 完成 |
| 成员退出/被移除通知 | notification.group.member_left.{group_id} | ✅ 完成 |
| 群组信息更新通知 | notification.group.info_updated.{group_id} | ✅ 完成 |
| 成员角色变更通知 | notification.group.role_changed.{group_id} | ✅ 完成 |
| 群组禁言通知 | notification.group.muted.{group_id} | ✅ 完成 |
| 群组解散通知 | notification.group.disbanded.{group_id} | ✅ 完成 |

---

## 5. Message Service (消息服务)

### HTTP接口

| 接口路由 | 功能 | 完成状态 |
|---------|------|----------|
| GET /api/v1/messages/offline | 获取离线消息 | ✅ 完成 |
| GET /api/v1/messages/history | 获取历史消息 | ✅ 完成 |
| POST /api/v1/messages/ack | 消息已读确认 | ✅ 完成 |
| GET /api/v1/groups/:id/messages/:msgId/reads | 获取群消息已读状态 | ✅ 完成 |
| GET /api/v1/messages/search | 搜索消息 | ✅ 完成 |

### WebSocket接口

| 接口 | 功能 | 完成状态 |
|------|------|----------|
| ws消息发送 | 实时消息发送 | ✅ 完成 |
| ws消息撤回 | 消息撤回 | ✅ 完成 |
| ws消息删除 | 消息删除 | ✅ 完成 |
| ws消息编辑 | 消息编辑 | ✅ 完成 |

### WebSocket通知

| 通知类型 | 主题 | 完成状态 |
|---------|------|----------|
| 新消息通知 | notification.message.new.{to_user_id} | ✅ 完成 |
| 消息已读回执通知 | notification.message.read_receipt.{from_user_id} | ✅ 完成 |
| 消息撤回通知 | notification.message.recalled.{conversation_id} | ✅ 完成 |
| 消息删除通知 | notification.message.deleted.{user_id} | ✅ 完成 |
| 消息编辑通知 | notification.message.edited.{user_id} | ✅ 完成 |
| 正在输入提示 | notification.message.typing.{to_user_id} | ✅ 完成 |
| @提及通知 | notification.message.mentioned.{user_id} | ✅ 完成 |

---

## 6. Conversation Service (会话服务)

### HTTP接口

| 接口路由 | 功能 | 完成状态 |
|---------|------|----------|
| GET /api/v1/conversations | 获取会话列表 | ✅ 完成 |
| GET /api/v1/conversations/unread/total | 获取总未读数 | ✅ 完成 |
| GET /api/v1/conversations/:conversationId | 获取会话详情 | ✅ 完成 |
| DELETE /api/v1/conversations/:conversationId | 删除会话 | ✅ 完成 |
| PUT /api/v1/conversations/:conversationId/pin | 置顶会话 | ✅ 完成 |
| PUT /api/v1/conversations/:conversationId/mute | 静音会话 | ✅ 完成 |
| PUT /api/v1/conversations/:conversationId/burn | 设置阅后即焚 | ✅ 完成 |
| PUT /api/v1/conversations/:conversationId/auto_delete | 设置自动删除 | ✅ 完成 |
| POST /api/v1/conversations/:conversationId/read-all | 标记全部已读 | ✅ 完成 |
| POST /api/v1/conversations/:conversationId/messages/read | 批量标记消息已读 | ✅ 完成 |
| GET /api/v1/conversations/:conversationId/messages/unread-count | 获取未读数 | ✅ 完成 |
| GET /api/v1/conversations/:conversationId/messages/read-receipts | 获取已读回执 | ✅ 完成 |
| GET /api/v1/conversations/:conversationId/messages/sequence | 获取消息序列号 | ✅ 完成 |

### WebSocket接口

| 接口 | 功能 | 完成状态 |
|------|------|----------|
| 通过Gateway订阅 | 会话变更推送 | ✅ 完成 |

### WebSocket通知

| 通知类型 | 主题 | 完成状态 |
|---------|------|----------|
| 会话未读数更新通知 | notification.conversation.unread_updated.{user_id} | ✅ 完成 |
| 会话置顶状态同步 | notification.conversation.pin_updated.{user_id} | ✅ 完成 |
| 会话删除同步 | notification.conversation.deleted.{user_id} | ✅ 完成 |
| 会话免打扰设置同步 | notification.conversation.mute_updated.{user_id} | ✅ 完成 |
| 会话阅后即焚设置同步 | notification.conversation.burn_updated.{user_id} | ✅ 完成 |
| 会话自动删除设置同步 | notification.conversation.auto_delete_updated.{user_id} | ✅ 完成 |

---

## 7. File Service (文件服务)

### HTTP接口

| 接口路由 | 功能 | 完成状态 |
|---------|------|----------|
| POST /api/v1/files/upload-token | 生成上传Token | ✅ 完成 |
| POST /api/v1/files/:fileId/complete | 完成文件上传 | ✅ 完成 |
| GET /api/v1/files/:fileId/download | 生成下载URL | ✅ 完成 |
| GET /api/v1/files/:fileId | 获取文件信息 | ✅ 完成 |
| DELETE /api/v1/files/:fileId | 删除文件 | ✅ 完成 |
| GET /api/v1/files | 获取文件列表 | ✅ 完成 |

### WebSocket接口

无WebSocket接口

### WebSocket通知

| 通知类型 | 主题 | 完成状态 |
|---------|------|----------|
| 文件上传完成通知 | notification.file.upload_completed.{user_id} | ✅ 完成 |
| 文件处理进度通知 | notification.file.processing.{user_id} | ✅ 完成 |
| 文件过期提醒 | notification.file.expiring.{user_id} | ✅ 完成 |

---

## 8. RTC Service (音视频服务)

### HTTP接口

| 接口路由 | 功能 | 完成状态 |
|---------|------|----------|
| POST /api/v1/calls | 发起呼叫 | ✅ 完成 |
| GET /api/v1/calls | 获取通话记录 | ✅ 完成 |
| GET /api/v1/calls/:callId | 获取通话会话 | ✅ 完成 |
| POST /api/v1/calls/:callId/join | 加入通话 | ✅ 完成 |
| POST /api/v1/calls/:callId/reject | 拒绝通话 | ✅ 完成 |
| POST /api/v1/calls/:callId/end | 结束通话 | ✅ 完成 |
| POST /api/v1/meetings | 创建会议室 | ✅ 完成 |
| GET /api/v1/meetings | 获取会议室列表 | ✅ 完成 |
| GET /api/v1/meetings/:roomId | 获取会议室详情 | ✅ 完成 |
| POST /api/v1/meetings/:roomId/join | 加入会议室 | ✅ 完成 |
| POST /api/v1/meetings/:roomId/end | 结束会议室 | ✅ 完成 |

### WebSocket接口

| 接口 | 功能 | 完成状态 |
|------|------|----------|
| ws通话状态推送 | 实时通话状态通知 | ✅ 完成 |

### WebSocket通知

| 通知类型 | 主题 | 完成状态 |
|---------|------|----------|
| 音视频通话邀请 | notification.livekit.call_invite.{user_id} | ✅ 完成 |
| 通话状态变更通知 | notification.livekit.call_status.{call_id} | ✅ 完成 |
| 通话拒绝通知 | notification.livekit.call_rejected.{from_user_id} | ✅ 完成 |

---

## 9. Sync Service (数据同步服务)

> 已移除：原 Sync Service 已不再作为独立服务维护。数据同步功能已分散至各业务服务中，由各服务自行处理消息同步、离线拉取等操作。

---

## 10. Realtime Service (连接服务)

### HTTP接口

| 接口路由 | 功能 | 完成状态 |
|---------|------|----------|
| GET /api/v1/ws | WebSocket接入点 | ✅ 完成 |

### WebSocket接口

| 接口 | 功能 | 完成状态 |
|------|------|----------|
| /ws?token=xxx | WebSocket连接认证 | ✅ 完成 |
| ping/pong | 心跳保活 | ✅ 完成 |
| 消息发送 | 客户端消息发送 | ✅ 完成 |
| 消息推送 | 服务端消息推送 | ✅ 完成 |
| 在线状态 | 用户在线状态管理 | ✅ 完成 |

### WebSocket通知

Gateway作为核心枢纽，负责接收NATS消息并推送到客户端：
- 订阅所有 `notification.*.*.{user_id}` 主题
- 消息格式转换和优先级队列管理
- 推送失败时触发离线推送

---

## 11. Push Service (推送服务)

### HTTP接口

| 接口路由 | 功能 | 完成状态 |
|---------|------|----------|
| 无独立HTTP接口 | 集成在其他服务中 | - |

### WebSocket接口

无WebSocket接口

### WebSocket通知

| 通知类型 | 主题 | 完成状态 |
|---------|------|----------|
| 离线推送发送结果通知 | notification.push.delivery_status.{user_id} | ✅ 完成 |
| 推送Token失效通知 | notification.push.token_invalid.{user_id} | ✅ 完成 |

---

## 12. Admin Service (管理后台服务)

> 已移除：原 Admin Service 已不再作为独立服务维护。管理后台功能暂缓开发，待后续规划。

---

## 13. Version Service (版本服务)

### HTTP接口

| 接口路由 | 功能 | 完成状态 |
|---------|------|----------|
| GET /api/v1/versions/check | 检查版本更新 | ✅ 完成 |
| GET /api/v1/versions/latest | 获取最新版本 | ✅ 完成 |
| GET /api/v1/versions/list | 获取版本列表 | ✅ 完成 |
| POST /api/v1/versions/report | 上报版本信息 | ✅ 完成 |

### WebSocket接口

无WebSocket接口

### WebSocket通知

无通知

---

## 功能统计汇总

| 模块 | HTTP接口数 | WebSocket接口数 | 通知类型数 | 完成率 |
|------|-----------|-----------------|-----------|--------|
| User Service (用户与认证) | 22 | 1 | 6 | 100% |
| Friend Service | 9 | - | 5 | 100% |
| Group Service | 16 | - | 7 | 100% |
| Message Service | 5 | 4 | 7 | 100% |
| Conversation Service | 13 | - | 6 | 100% |
| File Service | 6 | - | 3 | 100% |
| RTC Service | 11 | 1 | 3 | 100% |
| Realtime Service | 1 | 5 | - | 100% |
| Push Service | - | - | 2 | 100% |
| Version Service | 4 | - | - | 100% |
| **总计** | **87** | **11** | **39** | **100%** |

> 注：所有设计文档中标记的功能均已完成。WebSocket通知列表示例性列出，实际实现时由Gateway统一订阅并转发。