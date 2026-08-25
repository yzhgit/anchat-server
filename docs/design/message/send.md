# 消息发送设计（WebSocket + HTTP）

## 1. 概述

消息发送模块负责会话消息的生产、持久化、序列号分配与分发，统一覆盖 WebSocket 实时发送和 HTTP 发送兜底。

设计目标：

- **顺序性**：同一会话内消息按 `sequence` 单调递增；
- **可追溯**：消息全量落库，支持按序拉取与历史重放；
- **低耦合分发**：通过 NATS 通知总线向在线网关与推送链路分发；
- **多端一致**：在线走实时通知，离线走增量补齐；
- **多入口一致**：WebSocket 与 HTTP 发送共享同一套 `SendMessage` 业务链路。

## 2. 功能范围

- [x] 发送消息（单聊/群聊）
- [x] WebSocket 实时发送
- [x] HTTP 发送兜底
- [x] 会话内序列号生成
- [x] 基于 `local_id` 的发送幂等
- [x] 增量拉取会话消息
- [x] @提及通知
- [x] 消息过期策略（自动删除/阅后即焚）

## 3. 数据模型

### 3.1 Message

```go
type Message struct {
    MessageID                  string
    ConversationID             string
    ConversationType           int16  // 1-single/2-group
    TargetID                   string // single=对方userID, group=groupID
    SenderID                   string
    ContentType                int16  // 1-text/2-image/3-video/4-audio/5-file/6-location/7-card
    Content                    string // JSON string
    Sequence                   int64  // 会话内递增序号
    ReplyTo                    *string
    AtUsers                    []string
    Status                     int16  // 0-正常 1-撤回 2-删除

    BurnAfterReadingSeconds    int32
    AutoDeleteExpireTime       *time.Time
    BurnAfterReadingExpireTime *time.Time
    ExpireTime                 *time.Time

    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 3.2 ConversationSequence

```go
type ConversationSequence struct {
    ConversationID string
    CurrentSeq     int64
}
```

用于维护会话维度最新序列号，保证消息按会话递增。

## 4. API 设计

### 4.1 WebSocket：发送消息

客户端请求：

```json
{
  "type": "message.send",
  "payload": {
    "conversationId": "single_u1_u2",
    "contentType": 1,
    "content": "{\"text\":\"hello\"}",
    "replyTo": "",
    "atUsers": [],
    "localId": "local-001"
  }
}
```

成功回执：

```json
{
  "type": "message.sent",
  "payload": {
    "messageId": "2f3c...",
    "sequence": 101,
    "timestamp": 1712550000,
    "localId": "local-001"
  }
}
```

### 4.2 HTTP：发送消息

- `POST /api/v1/messages`

请求体：

```json
{
  "conversation_id": "conv_xxx",
  "content_type": 1,
  "content": "{\"text\":\"hello\"}",
  "reply_to": "msg_xxx",
  "at_users": ["u1", "u2"],
  "local_id": "local-001"
}
```

响应体（data）：

```json
{
  "message_id": "msg_xxx",
  "sequence": 1024,
  "timestamp": "2026-04-08T14:23:45Z"
}
```

### 4.3 gRPC：SendMessage

```protobuf
message SendMessageRequest {
  string sender_id = 1;
  string conversation_id = 2;
  ContentType content_type = 3; // 1-text/2-image/3-video/4-audio/5-file/6-location/7-card
  string content = 4;             // JSON string
  optional string reply_to = 5;
  repeated string at_users = 6;
  string local_id = 7;
}

message SendMessageResponse {
  string message_id = 1;
  int64 sequence = 2;
  google.protobuf.Timestamp timestamp = 3;
}
```

### 4.4 gRPC：GetMessages

```protobuf
message GetMessagesRequest {
  string conversation_id = 1;
  optional int64 start_seq = 2;
  optional int64 end_seq = 3;
  int32 limit = 4;
  bool reverse = 5;
}

message GetMessagesResponse {
  repeated Message messages = 1;
  int64 total = 2;
  bool has_more = 3;
}
```

## 5. 核心流程

### 5.1 WebSocket 发送消息

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant MessageService
    participant ConversationService
    participant GroupService
    participant DB
    participant NATS

    Client->>Gateway: GET /api/v1/ws?token={jwt}
    Client->>Gateway: WS message.send\npayload={conversationId, contentType, content, replyTo, atUsers, localId}
    Gateway->>MessageService: gRPC SendMessage(sender_id, conversation_id, content_type, content, local_id, ...)
    MessageService->>ConversationService: gRPC GetConversation(user_id, conversation_id)
    ConversationService-->>MessageService: conversation_type, target_id, auto_delete_duration, burn_after_reading
    opt 群聊
        MessageService->>GroupService: gRPC IsMember(group_id=target_id, user_id=sender_id)
        GroupService-->>MessageService: is_member
    end
    MessageService->>DB: 事务内分配 sequence 并保存消息
    MessageService->>NATS: 发布 message.new / message.mentioned（可选）
    MessageService-->>Gateway: SendMessageResponse
    Gateway-->>Client: WS message.sent\npayload={messageId, sequence, timestamp, localId}
```

### 5.2 HTTP 发送消息

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant MessageService
    participant ConversationService
    participant GroupService
    participant DB
    participant NATS

    Client->>Gateway: POST /api/v1/messages\nHeader: Authorization: Bearer {token}
    Gateway->>MessageService: gRPC SendMessage(sender_id, conversation_id, content_type, content, local_id, ...)
    MessageService->>ConversationService: gRPC GetConversation(user_id, conversation_id)
    opt 群聊
        MessageService->>GroupService: gRPC IsMember(group_id=target_id, user_id=sender_id)
    end
    MessageService->>DB: 事务内分配 sequence 并保存消息
    MessageService->>NATS: 发布 message.new / message.mentioned（可选）
    MessageService-->>Gateway: SendMessageResponse
    Gateway-->>Client: 200 OK
```

### 5.3 在线实时接收

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant MessageService
    participant NATS

    MessageService->>NATS: notification.message.new.{target}
    NATS-->>Gateway: notification.*.*.{userID}
    Gateway-->>Client: WS notification\npayload=Notification
```

### 5.4 离线消息补齐

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant SyncService
    participant MessageService
    participant DB

    Client->>Gateway: POST /api/v1/sync/messages?limit=50\nBody={conversationSeqs:[{conversationId, conversationType, lastSeq}]}
    Gateway->>SyncService: gRPC SyncMessages(userId, conversationSeqs, limit)

    loop 每个会话
        SyncService->>MessageService: gRPC GetMessages(conversation_id, start_seq=last_seq+1, reverse=false)
        MessageService->>DB: 按 sequence 查询消息
        DB-->>MessageService: messages
        MessageService-->>SyncService: messages + has_more
    end

    SyncService-->>Gateway: conversations[]
    Gateway-->>Client: 200 OK
```

## 6. 消息类型与状态

### 6.1 `content_type`

- `1`：text
- `2`：image
- `3`：video
- `4`：audio
- `5`：file
- `6`：location
- `7`：card

### 6.2 `status`

- `0`：正常
- `1`：撤回
- `2`：删除

## 7. 通知主题

- `notification.message.new.{receiver_user_id}`
- `notification.message.mentioned.{user_id}`

## 8. 设计约束

- 会话顺序由 `sequence` 保证；客户端应以 `sequence` 作为排序与去重基准；
- 发送请求以 `conversation_id` 作为主键，`target_id` 与 `conversation_type` 由服务端从会话推导；
- 发送前必须完成会话归属与成员权限校验；
- `local_id` 在发送场景必填，幂等作用域为 `(sender_id, conversation_id, local_id)`；
- 自动删除与阅后即焚时长以会话配置为准，在发送落库时生成策略快照；
- 单次拉取建议限制数量上限，避免大会话单次返回过大。
