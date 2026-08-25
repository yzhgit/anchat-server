# 推送通知接口设计

本文档定义各微服务的推送通知接口规范，基于 **WebSocket Realtime + NATS** 架构实现实时通知推送。

## 架构说明

### 通知流程
```
服务层 --[NATS Publish]--> Realtime Service --[WebSocket Push]--> 客户端
```

### NATS主题命名规范
- 格式: `notification.{service}.{event_type}.{target_user_id}`
- 示例: `notification.friend.request.user-123`
- 通配符订阅: `notification.*.*.user-123` (Realtime为特定用户订阅所有通知)

### 通知消息通用格式
```json
{
  "notification_id": "uuid",
  "type": "friend.request",
  "timestamp": 1234567890,
  "from_user_id": "user-456",
  "to_user_id": "user-123",
  "priority": "high|normal|low",
  "payload": {
    // 业务数据
  },
  "metadata": {
    "ttl": 3600,
    "require_ack": true
  }
}
```

---

## 4.1 User Service - 推送令牌管理

> **说明**: 原Auth Service已被User Service取代，推送通知令牌（Push Token）的管理统一由User Service负责。

### 推送令牌管理接口

User Service提供推送令牌（用于离线推送）的注册和维护功能。客户端在启动时或令牌变更时应主动上报推送令牌。

**相关接口**:

| 接口 | HTTP路径 | 方法 | 说明 |
|------|---------|------|------|
| `UpdatePushToken` | `/api/v1/users/push_token` | POST | 更新/注册推送令牌 |

**请求格式** (`UpdatePushTokenRequest`):
```json
{
  "user_id": "user-123",
  "device_id": "device-456",
  "push_token": "xxxxx_token_value",
  "platform": 1
}
```

- `user_id`: 用户ID
- `device_id`: 设备唯一标识
- `push_token`: 平台推送令牌（iOS APNs Device Token / Android FCM Token）
- `platform`: 推送平台，`1`-iOS / `2`-Android

**数据存储**:
推送令牌存储在 `user_push_tokens` 表中，由 `UserPushTokenRepository` 负责数据访问，支持以下操作：
- `CreateOrUpdate`: 创建或更新推送令牌
- `GetByUserID`: 获取用户所有推送令牌
- `GetByUserIDAndDeviceID`: 获取指定设备的推送令牌
- `Delete`: 删除指定设备的推送令牌

**推送通知场景（由User Service触发）**:

1. **多端登录互踢通知**
   - NATS主题: `notification.user.force_logout.{user_id}`
   - 触发时机: 新设备登录触发互踢策略时
   - 消息格式:
   ```json
   {
     "type": "user.force_logout",
     "payload": {
       "reason": "new_device_login",
       "device_type": "iOS",
       "device_id": "device-789",
       "login_time": 1234567890,
       "login_location": "Beijing, China"
     }
   }
   ```

2. **密码修改通知**
   - NATS主题: `notification.user.password_changed.{user_id}`
   - 触发时机: 用户修改密码成功后，通知其他设备重新登录
   - 消息格式:
   ```json
   {
     "type": "user.password_changed",
     "payload": {
       "changed_at": 1234567890,
       "device_type": "Web",
       "requires_relogin": true
     }
   }
   ```

3. **推送令牌失效通知**
   - NATS主题: `notification.user.push_token_invalid.{user_id}`
   - 触发时机: Push Service检测到推送令牌失效时通知User Service
   - 消息格式:
   ```json
   {
     "type": "user.push_token_invalid",
     "payload": {
       "device_id": "device-123",
       "platform": "1|2",
       "reason": "token_expired|app_uninstalled",
       "detected_at": 1234567890
     }
   }
   ```

**实现要点**:
- 互踢通知需要高优先级推送（priority: high）
- 需要客户端ACK确认，确保通知送达
- 通知失败时触发离线推送
- 推送令牌更新为幂等操作，同一 `user_id + device_id` 组合下重复更新以最新令牌为准

---

## 4.2 User Service - 推送通知接口

**推送场景**:

1. **用户资料更新通知**
   - NATS主题: `notification.user.profile_updated.{user_id}`
   - 触发时机: 用户修改个人资料（头像、昵称等）
   - 消息格式:
   ```json
   {
     "type": "user.profile_updated",
     "payload": {
       "user_id": "user-123",
       "updated_fields": ["avatar", "nickname"],
       "avatar_url": "https://cdn.example.com/avatar/xxx.jpg",
       "nickname": "新昵称",
       "updated_at": 1234567890
     }
   }
   ```
   - 用途: 通知其他端同步更新用户信息

2. **好友资料变更通知**
   - NATS主题: `notification.user.friend_profile_changed.{user_id}`
   - 触发时机: 好友修改了个人资料
   - 消息格式:
   ```json
   {
     "type": "user.friend_profile_changed",
     "payload": {
       "friend_user_id": "user-456",
       "updated_fields": ["avatar", "nickname", "signature"],
       "avatar_url": "https://cdn.example.com/avatar/yyy.jpg",
       "nickname": "朋友新昵称",
       "signature": "新个性签名"
     }
   }
   ```

3. **在线状态变更通知**
   - NATS主题: `notification.user.status_changed.{user_id}`
   - 触发时机: 用户上线/下线/离开
   - 消息格式:
   ```json
   {
     "type": "user.status_changed",
     "payload": {
       "user_id": "user-456",
       "status": "online|offline|away",
       "last_active_at": 1234567890,
       "platform": "iOS|Android|Web"
     }
   }
   ```

**实现要点**:
- 用户资料更新需推送到用户所有在线设备
- 好友资料变更需推送到所有关注该好友的用户
- 在线状态变更采用低优先级，避免频繁推送影响性能

---

## 4.3 Friend Service - 推送通知接口

**推送场景**:

1. **好友请求通知**
   - NATS主题: `notification.friend.request.{to_user_id}`
   - 触发时机: 收到好友申请
   - 消息格式:
   ```json
   {
     "type": "friend.request",
     "payload": {
       "request_id": "req-123",
       "from_user_id": "user-456",
       "from_nickname": "张三",
       "from_avatar": "https://cdn.example.com/avatar/xxx.jpg",
       "message": "你好，我是张三",
       "source": "search|qrcode|group",
       "created_at": 1234567890
     }
   }
   ```

2. **好友请求处理结果通知**
   - NATS主题: `notification.friend.request_handled.{from_user_id}`
   - 触发时机: 好友请求被接受或拒绝
   - 消息格式:
   ```json
   {
     "type": "friend.request_handled",
     "payload": {
       "request_id": "req-123",
       "to_user_id": "user-789",
       "to_nickname": "李四",
       "to_avatar": "https://cdn.example.com/avatar/yyy.jpg",
       "status": "accepted|rejected",
       "handled_at": 1234567890
     }
   }
   ```

3. **好友删除通知**
   - NATS主题: `notification.friend.deleted.{user_id}`
   - 触发时机: 被好友删除
   - 消息格式:
   ```json
   {
     "type": "friend.deleted",
     "payload": {
       "friend_user_id": "user-456",
       "deleted_at": 1234567890
     }
   }
   ```

4. **好友备注修改同步**
   - NATS主题: `notification.friend.remark_updated.{user_id}`
   - 触发时机: 修改好友备注（多端同步）
   - 消息格式:
   ```json
   {
     "type": "friend.remark_updated",
     "payload": {
       "friend_user_id": "user-456",
       "remark": "老同学",
       "updated_at": 1234567890
     }
   }
   ```

5. **黑名单变更通知**
   - NATS主题: `notification.friend.blacklist_changed.{user_id}`
   - 触发时机: 添加/移除黑名单
   - 消息格式:
   ```json
   {
     "type": "friend.blacklist_changed",
     "payload": {
       "target_user_id": "user-789",
       "action": "add|remove",
       "changed_at": 1234567890
     }
   }
   ```

**实现要点**:
- 好友请求通知需要高优先级推送，确保及时送达
- 好友请求被接受时，双方都需要收到通知
- 备注修改仅推送到用户自己的其他设备，不推送给对方

---

## 4.4 Group Service - 推送通知接口

**推送场景**:

1. **群组邀请通知**
   - NATS主题: `notification.group.invited.{user_id}`
   - 触发时机: 被邀请加入群组
   - 消息格式:
   ```json
   {
     "type": "group.invited",
     "payload": {
       "group_id": "group-123",
       "group_name": "工作群",
       "group_avatar": "https://cdn.example.com/group/xxx.jpg",
       "inviter_user_id": "user-456",
       "inviter_nickname": "王五",
       "invited_at": 1234567890,
       "require_approval": false
     }
   }
   ```

2. **加入群组通知**
   - NATS主题: `notification.group.member_joined.{group_id}`
   - 触发时机: 新成员加入群组
   - 消息格式:
   ```json
   {
     "type": "group.member_joined",
     "payload": {
       "group_id": "group-123",
       "user_id": "user-789",
       "nickname": "新成员",
       "avatar": "https://cdn.example.com/avatar/zzz.jpg",
       "inviter_user_id": "user-456",
       "joined_at": 1234567890
     }
   }
   ```
   - 推送对象: 所有群成员

3. **成员退出/被移除通知**
   - NATS主题: `notification.group.member_left.{group_id}`
   - 触发时机: 成员主动退群或被管理员移除
   - 消息格式:
   ```json
   {
     "type": "group.member_left",
     "payload": {
       "group_id": "group-123",
       "user_id": "user-789",
       "nickname": "已退出成员",
       "reason": "self_quit|removed_by_admin",
       "operator_user_id": "user-456",
       "left_at": 1234567890
     }
   }
   ```
   - 推送对象: 所有群成员

4. **群组信息更新通知**
   - NATS主题: `notification.group.info_updated.{group_id}`
   - 触发时机: 群名称、群头像、群公告等信息变更
   - 消息格式:
   ```json
   {
     "type": "group.info_updated",
     "payload": {
       "group_id": "group-123",
       "updated_fields": ["name", "avatar", "announcement"],
       "group_name": "新群名称",
       "group_avatar": "https://cdn.example.com/group/new.jpg",
       "announcement": "新群公告内容",
       "operator_user_id": "user-456",
       "updated_at": 1234567890
     }
   }
   ```
   - 推送对象: 所有群成员

5. **成员角色变更通知**
   - NATS主题: `notification.group.role_changed.{group_id}`
   - 触发时机: 成员被设置为管理员或取消管理员
   - 消息格式:
   ```json
   {
     "type": "group.role_changed",
     "payload": {
       "group_id": "group-123",
       "user_id": "user-789",
       "old_role": "member",
       "new_role": "admin",
       "operator_user_id": "user-456",
       "changed_at": 1234567890
     }
   }
   ```
   - 推送对象: 所有群成员

6. **群组禁言通知**
   - NATS主题: `notification.group.muted.{group_id}`
   - 触发时机: 成员被禁言或全员禁言开启
   - 消息格式:
   ```json
   {
     "type": "group.muted",
     "payload": {
       "group_id": "group-123",
       "mute_type": "all_members|specific_member",
       "target_user_id": "user-789",
       "duration": 3600,
       "operator_user_id": "user-456",
       "muted_at": 1234567890,
       "unmute_at": 1234571490
     }
   }
   ```
   - 推送对象: 全员禁言推送给所有成员，单人禁言推送给被禁言者和管理员

7. **群组解散通知**
   - NATS主题: `notification.group.disbanded.{group_id}`
   - 触发时机: 群主解散群组
   - 消息格式:
   ```json
   {
     "type": "group.disbanded",
     "payload": {
       "group_id": "group-123",
       "group_name": "已解散群",
       "operator_user_id": "user-456",
       "disbanded_at": 1234567890
     }
   }
   ```
   - 推送对象: 所有群成员

**实现要点**:
- 群组通知需要支持批量推送到所有成员
- Realtime需要维护group_id到user_id列表的映射关系
- 群组通知优先级：解散>禁言>邀请>信息变更
- 大群（成员>500）考虑延迟推送或仅推送关键通知

---

## 4.5 Message Service - 推送通知接口

**推送场景**:

1. **新消息通知**
   - NATS主题: `notification.message.new.{to_user_id}`
   - 触发时机: 收到新的单聊或群聊消息
   - 消息格式:
   ```json
   {
     "type": "message.new",
     "payload": {
       "message_id": "msg-123",
       "conversation_id": "conv-456",
       "conversation_type": "1|2",
       "from_user_id": "user-789",
       "from_nickname": "发送者",
       "from_avatar": "https://cdn.example.com/avatar/xxx.jpg",
       "content_type": "1|2|3|4|5|6|7",
       "content": "消息内容或摘要",
       "sent_at": 1234567890,
       "seq": 12345
     }
   }
   ```
   - 推送对象: 单聊推送给接收者，群聊推送给所有成员（除发送者）

2. **消息已读回执通知**
   - NATS主题: `notification.message.read_receipt.{from_user_id}`
   - 触发时机: 对方已读消息
   - 消息格式:
   ```json
   {
     "type": "message.read_receipt",
     "payload": {
       "conversation_id": "conv-456",
       "conversation_type": "1|2",
       "reader_user_id": "user-789",
       "last_read_seq": 12345,
       "read_at": 1234567890
     }
   }
   ```

3. **消息撤回通知**
   - NATS主题: `notification.message.recalled.{conversation_id}`
   - 触发时机: 消息被撤回
   - 消息格式:
   ```json
   {
     "type": "message.recalled",
     "payload": {
       "message_id": "msg-123",
       "conversation_id": "conv-456",
       "conversation_type": "1|2",
       "operator_user_id": "user-789",
       "recalled_at": 1234567890
     }
   }
   ```
   - 推送对象: 会话中所有成员

4. **正在输入提示**
   - NATS主题: `notification.message.typing.{to_user_id}`
   - 触发时机: 对方正在输入（单聊）
   - 消息格式:
   ```json
   {
     "type": "message.typing",
     "payload": {
       "conversation_id": "conv-456",
       "from_user_id": "user-789",
       "typing": true,
       "timestamp": 1234567890
     }
   }
   ```
   - 特点: 低优先级，客户端节流发送（3秒一次）

5. **@提及通知**
   - NATS主题: `notification.message.mentioned.{user_id}`
   - 触发时机: 群聊中被@
   - 消息格式:
   ```json
   {
     "type": "message.mentioned",
     "payload": {
       "message_id": "msg-123",
       "group_id": "group-456",
       "group_name": "工作群",
       "from_user_id": "user-789",
       "from_nickname": "发送者",
       "content": "@你 的消息内容",
       "mention_type": "single|all",
       "sent_at": 1234567890
     }
   }
   ```
   - 特点: 高优先级推送

**实现要点**:
- 新消息通知是最高频的推送类型，需要优化性能
- 群聊消息需要批量推送，避免单个发送
- 正在输入通知采用临时订阅机制，不持久化
- 消息推送失败时触发离线推送

---

## 4.6 Conversation Service - 推送通知接口

**推送场景**:

1. **会话未读数更新通知**
   - NATS主题: `notification.conversation.unread_updated.{user_id}`
   - 触发时机: 会话未读数变化（新消息、已读）
   - 消息格式:
   ```json
   {
     "type": "conversation.unread_updated",
     "payload": {
       "conversation_id": "conv-456",
       "conversation_type": "1|2",
       "unread_count": 5,
       "total_unread_count": 20,
       "last_message": {
         "message_id": "msg-123",
         "content": "最新消息摘要",
         "sent_at": 1234567890
       }
     }
   }
   ```

2. **会话置顶状态同步**
   - NATS主题: `notification.conversation.pin_updated.{user_id}`
   - 触发时机: 会话置顶/取消置顶（多端同步）
   - 消息格式:
   ```json
   {
     "type": "conversation.pin_updated",
     "payload": {
       "conversation_id": "conv-456",
       "is_pinned": true,
       "pin_time": 1234567890
     }
   }
   ```

3. **会话删除同步**
   - NATS主题: `notification.conversation.deleted.{user_id}`
   - 触发时机: 删除会话（多端同步）
   - 消息格式:
   ```json
   {
     "type": "conversation.deleted",
     "payload": {
       "conversation_id": "conv-456",
       "deleted_at": 1234567890
     }
   }
   ```

4. **会话免打扰设置同步**
   - NATS主题: `notification.conversation.mute_updated.{user_id}`
   - 触发时机: 设置/取消免打扰（多端同步）
   - 消息格式:
   ```json
   {
     "type": "conversation.mute_updated",
     "payload": {
       "conversation_id": "conv-456",
       "is_muted": true,
       "muted_until": 1234567890
     }
   }
   ```

**实现要点**:
- 会话通知主要用于多端同步，优先级为normal
- 未读数更新需要及时推送，影响用户体验
- 会话操作需要保证幂等性，避免重复推送

---

## 4.7 File Service - 推送通知接口

**推送场景**:

1. **文件上传完成通知**
   - NATS主题: `notification.file.upload_completed.{user_id}`
   - 触发时机: 文件上传到MinIO成功后
   - 消息格式:
   ```json
   {
     "type": "file.upload_completed",
     "payload": {
       "file_id": "file-123",
       "file_name": "document.pdf",
       "file_size": 1024000,
       "file_type": "1|2|3|4|5", // 1-image/2-video/3-audio/4-file/5-log
       "mime_type": "application/pdf",
       "download_url": "https://cdn.example.com/files/xxx",
       "thumbnail_url": "https://cdn.example.com/thumbnails/xxx",
       "uploaded_at": 1234567890
     }
   }
   ```
   - 用途: 通知用户其他端文件上传成功

2. **文件处理进度通知**
   - NATS主题: `notification.file.processing.{user_id}`
   - 触发时机: 大文件上传/压缩/转码处理中
   - 消息格式:
   ```json
   {
     "type": "file.processing",
     "payload": {
       "file_id": "file-123",
       "file_name": "video.mp4",
       "status": "processing|completed|failed",
       "progress": 75,
       "message": "视频转码中..."
     }
   }
   ```

3. **文件过期提醒**
   - NATS主题: `notification.file.expiring.{user_id}`
   - 触发时机: 临时文件即将过期（提前24小时）
   - 消息格式:
   ```json
   {
     "type": "file.expiring",
     "payload": {
       "file_id": "file-123",
       "file_name": "temp_file.zip",
       "expires_at": 1234567890,
       "hours_remaining": 24
     }
   }
   ```

**实现要点**:
- 文件上传完成通知优先级为normal
- 文件处理进度通知采用低优先级，避免频繁推送
- 文件过期提醒通过定时任务触发

---

## 4.8 Push Service - 推送通知接口

**推送场景**:

1. **离线推送发送结果通知**
   - NATS主题: `notification.push.delivery_status.{user_id}`
   - 触发时机: 离线推送发送成功/失败
   - 消息格式:
   ```json
   {
     "type": "push.delivery_status",
     "payload": {
       "push_id": "push-123",
       "platform": "1|2",
       "status": "sent|failed|clicked",
       "error_message": "APNs rejected",
       "sent_at": 1234567890
     }
   }
   ```

2. **推送Token失效通知**
   - NATS主题: `notification.push.token_invalid.{user_id}`
   - 触发时机: 推送Token过期或无效
   - 消息格式:
   ```json
   {
     "type": "push.token_invalid",
     "payload": {
       "device_id": "device-123",
       "platform": "1|2",
       "old_token": "expired_token_xxx",
       "reason": "token_expired|app_uninstalled",
       "detected_at": 1234567890
     }
   }
   ```

**实现要点**:
- 这些通知主要用于内部监控和调试
- 优先级为low，不影响业务核心流程

---

## 4.9 Realtime Service - 推送通知接口

Realtime Service是推送通知的核心枢纽，负责：

**核心职责**:

1. **NATS订阅管理**
   - 为每个连接的用户订阅: `notification.*.*.{user_id}`
   - 为用户所在的群组订阅: `notification.group.*.{group_id}`
   - 订阅系统广播: `notification.system.broadcast.*`

2. **WebSocket推送**
   - 接收NATS消息后，通过WebSocket推送给客户端
   - 消息格式转换（NATS → WebSocket）
   - 消息优先级队列管理

3. **推送失败处理**
   - 用户不在线: 触发Push Service发送离线推送
   - 推送失败重试（3次，指数退避）
   - 记录推送失败日志

4. **连接状态管理**
   - 维护user_id到WebSocket连接的映射
   - 维护group_id到成员列表的映射（用于群组推送）
   - 心跳检测和自动重连

**WebSocket推送消息格式**:
```json
{
  "type": "notification",
  "data": {
    "notification_id": "uuid",
    "notification_type": "friend.request",
    "timestamp": 1234567890,
    "payload": {
      // 具体通知内容
    }
  }
}
```

**实现要点**:
- Realtime需要高性能，单实例支持10万+并发连接
- 使用Redis存储连接状态，支持多Realtime实例负载均衡
- 群组推送采用批量发送，降低网络开销

---

## 4.10 LiveKit Service - 推送通知接口

**推送场景**:

1. **音视频通话邀请**
   - NATS主题: `notification.livekit.call_invite.{user_id}`
   - 触发时机: 收到音视频通话邀请
   - 消息格式:
   ```json
   {
     "type": "livekit.call_invite",
     "payload": {
       "call_id": "call-123",
       "room_name": "room-456",
       "call_type": "0|1", // 0-audio/1-video
       "from_user_id": "user-789",
       "from_nickname": "张三",
       "from_avatar": "https://cdn.example.com/avatar/xxx.jpg",
       "invited_users": ["user-123", "user-456"],
       "livekit_token": "eyJhbGci...",
       "expires_at": 1234567890
     }
   }
   ```
   - 特点: 最高优先级推送

2. **通话状态变更通知**
   - NATS主题: `notification.livekit.call_status.{call_id}`
   - 触发时机: 通话开始/结束/有人加入/退出
   - 消息格式:
   ```json
   {
     "type": "livekit.call_status",
     "payload": {
       "call_id": "call-123",
       "status": "started|ended|user_joined|user_left",
       "user_id": "user-789",
       "timestamp": 1234567890
     }
   }
   ```
   - 推送对象: 所有通话参与者

3. **通话拒绝通知**
   - NATS主题: `notification.livekit.call_rejected.{from_user_id}`
   - 触发时机: 对方拒接通话
   - 消息格式:
   ```json
   {
     "type": "livekit.call_rejected",
     "payload": {
       "call_id": "call-123",
       "rejector_user_id": "user-456",
       "rejector_nickname": "李四",
       "rejected_at": 1234567890,
       "reason": "busy|declined"
     }
   }
   ```

**实现要点**:
- 通话邀请需要最高优先级推送，确保及时送达
- 通话邀请超时（30秒）后自动取消
- 推送失败时触发电话振铃式离线推送

---

## 4.11 Sync Service - 推送通知接口

> **说明**: Sync Service（同步服务）已从项目中移除。原有的数据同步通知功能现已分散到各相关微服务中处理。

**功能迁移说明**:

| 原Sync Service功能 | 迁移至 | 说明 |
|-------------------|--------|------|
| 数据同步请求通知 | 各业务服务各自处理 | 各服务在多端操作时直接通过NATS发布对应的通知事件 |
| 同步完成通知 | 各业务服务各自处理 | 无需集中同步，各服务独立维护数据一致性 |
| 好友数据同步 | Friend Service | 好友关系变更通过 `notification.friend.*` 主题推送 |
| 会话数据同步 | Conversation Service | 会话操作通过 `notification.conversation.*` 主题推送 |
| 消息数据同步 | Message Service | 消息相关变更通过 `notification.message.*` 主题推送 |
| 群组数据同步 | Group Service | 群组变更通过 `notification.group.*` 主题推送 |

原有的循环同步防护由各服务自行实现，通过消息中的 `device_id` 字段识别发送源设备，避免回环推送。

---

## 4.12 Admin Service - 推送通知接口

> **说明**: Admin Service（管理服务）已从项目中移除。原有的管理员通知功能不再单独维护。

**功能替代方案**:

| 原Admin Service功能 | 替代方案 |
|-------------------|---------|
| 系统公告推送 | 如需系统公告功能，可通过 Push Service 的广播能力直接实现 |
| 用户封禁通知 | 由 User Service 的 `UpdateStatus` 能力配合 Realtime 强制断连实现 |
| 系统维护通知 | 通过系统级别的广播通知或客户端配置中心实现 |

如需恢复管理后台相关功能，建议以独立微服务或现有服务的扩展接口方式实现，并通过统一的 NATS 通知机制进行推送。

---

## 附录: NATS JetStream配置示例

### Stream配置

```go
// Notification Stream
streamConfig := &nats.StreamConfig{
    Name:        "NOTIFICATIONS",
    Subjects:    []string{"notification.>"},
    Retention:   nats.WorkQueuePolicy,
    MaxAge:      24 * time.Hour,  // 保留24小时
    Storage:     nats.FileStorage,
    Replicas:    3,  // 高可用
    Discard:     nats.DiscardOld,
}
```

### Consumer配置（Realtime订阅）

```go
// Realtime为每个用户创建Consumer
consumerConfig := &nats.ConsumerConfig{
    Durable:       fmt.Sprintf("realtime-user-%s", userID),
    FilterSubject: fmt.Sprintf("notification.*.*.%s", userID),
    AckPolicy:     nats.AckExplicitPolicy,
    MaxDeliver:    3,
    AckWait:       30 * time.Second,
}
```

### 发布通知示例代码

```go
// Friend Service发布好友请求通知
func (s *FriendService) PublishFriendRequestNotification(ctx context.Context, req *FriendRequest) error {
    notification := Notification{
        NotificationID: uuid.New().String(),
        Type:          "friend.request",
        Timestamp:     time.Now().Unix(),
        FromUserID:    req.FromUserID,
        ToUserID:      req.ToUserID,
        Priority:      "high",
        Payload: map[string]interface{}{
            "request_id":   req.ID,
            "from_user_id": req.FromUserID,
            "message":      req.Message,
            "source":       req.Source,
        },
        Metadata: map[string]interface{}{
            "ttl":         3600,
            "require_ack": true,
        },
    }

    data, _ := json.Marshal(notification)
    subject := fmt.Sprintf("notification.friend.request.%s", req.ToUserID)

    _, err := s.natsConn.Publish(subject, data)
    return err
}
```

### Realtime订阅和推送示例代码

```go
// Realtime订阅用户通知
func (g *Realtime) SubscribeUserNotifications(userID string, conn *websocket.Conn) error {
    subject := fmt.Sprintf("notification.*.*.%s", userID)

    sub, err := g.natsConn.Subscribe(subject, func(msg *nats.Msg) {
        var notification Notification
        json.Unmarshal(msg.Data, &notification)

        // 推送到WebSocket
        wsMsg := WebSocketMessage{
            Type: "notification",
            Data: notification,
        }

        if err := conn.WriteJSON(wsMsg); err != nil {
            log.Errorf("Failed to push notification to user %s: %v", userID, err)
            // 触发离线推送
            g.pushService.SendOfflinePush(userID, notification)
        } else {
            // 确认消息
            msg.Ack()
        }
    })

    return err
}
```

---

## 性能优化建议

### 1. 批量推送优化
- 群组消息推送采用批量发送，单次最多推送1000个用户
- 大群（>1000人）分批推送，避免阻塞

### 2. 优先级队列
- 高优先级: 音视频邀请、好友请求、系统封禁
- 普通优先级: 新消息、好友资料变更
- 低优先级: 正在输入、在线状态变更

### 3. 推送节流
- 正在输入通知: 客户端3秒发送一次
- 在线状态变更: 5秒合并一次推送
- 未读数更新: 1秒内合并多次更新

### 4. 缓存策略
- Realtime缓存用户在线状态（Redis）
- 群组成员列表缓存（Redis，TTL 10分钟）
- 通知模板缓存，减少序列化开销

### 5. 监控指标
- 通知发送成功率
- 推送延迟（P50、P95、P99）
- WebSocket连接数
- NATS消息堆积情况

---

## 总结

本设计文档定义了各微服务的推送通知接口，基于 **WebSocket Realtime + NATS JetStream** 架构实现实时通知推送。当前项目包含 User、Friend、Group、Message、Conversation、File、Push、Realtime、RTC（Real-Time Communication）等微服务。原Auth Service已由User Service取代，Sync Service和Admin Service已移除，相关功能已分散到各对应服务中。

**核心设计原则**:
1. **单一职责**: 每个服务只发布自己领域的通知事件
2. **解耦**: 服务不直接调用Realtime，通过NATS解耦
3. **可靠性**: 通知推送失败时触发离线推送，确保送达
4. **性能**: 批量推送、优先级队列、节流策略
5. **可扩展**: NATS JetStream支持水平扩展

**关键指标**:
- 通知推送延迟 < 100ms (P95)
- 通知送达率 > 99.9%
- 单Realtime实例支持 10万+ 并发连接
- NATS消息吞吐 > 100万 msg/s
