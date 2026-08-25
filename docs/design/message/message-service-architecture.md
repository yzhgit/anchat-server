# AnyChat 消息服务架构设计

## 1. 方案选型分析

### 1.1 方案对比

#### 方案A：NATS 直连客户端
```
客户端 <--NATS Protocol--> NATS Server <--NATS--> 微服务
```

**优点**：
- 简化架构，减少一层转发
- NATS 原生支持 Pub/Sub 和 Request/Reply

**缺点**：
- ❌ NATS 为服务器间通信设计，不适合大量客户端
- ❌ 客户端 SDK 支持有限（移动端、Web 兼容性）
- ❌ 难以实现细粒度权限控制和连接管理
- ❌ 网络穿透性差（防火墙、NAT）
- ❌ 无法在边界层做协议适配、消息压缩、限流
- ❌ 后端服务直接暴露，安全风险高
- ❌ 客户端升级困难（协议耦合）

#### 方案B（推荐）：WebSocket Gateway + NATS
```
客户端 <--WebSocket--> Gateway <--NATS--> 微服务
```

**优点**：
- ✅ 关注点分离：客户端协议 vs 服务器间通信
- ✅ WebSocket 广泛支持（浏览器、iOS、Android）
- ✅ Gateway 可做鉴权、限流、协议转换、消息压缩
- ✅ 后端服务不直接暴露，安全性高
- ✅ 灵活性：更换后端消息队列不影响客户端
- ✅ 易于优化：连接池管理、消息批处理、心跳检测
- ✅ 符合业界主流实践（微信、WhatsApp、Telegram）

**缺点**：
- 多一层转发（延迟增加 < 5ms，可接受）
- Gateway 需要管理连接状态（Redis + 内存）

### 1.2 业界最佳实践参考

| IM 产品 | 客户端协议 | 服务器间通信 | 架构特点 |
|---------|-----------|-------------|---------|
| 微信 | 自定义 TCP 长连接 | 消息队列 | Gateway + 业务服务分离 |
| WhatsApp | 自定义协议 | Erlang 消息传递 | 高并发连接管理 |
| Telegram | MTProto | 内部消息总线 | 多数据中心同步 |
| Slack | WebSocket | Kafka | 实时协作优化 |
| Discord | WebSocket | NATS/Kafka | Gateway 负载均衡 |

**共同点**：
1. 客户端使用 WebSocket 或自定义 TCP 协议
2. Gateway 层负责连接管理和协议转换
3. 后端使用消息队列处理异步消息流转
4. 分离在线状态管理和消息持久化

---

## 2. 推荐架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                         客户端层                             │
│  iOS / Android / Web / Desktop                              │
└─────────────────────────────────────────────────────────────┘
                          ↓ WebSocket
┌─────────────────────────────────────────────────────────────┐
│                      Realtime Service                       │
│  - WebSocket 连接管理                                        │
│  - 用户在线状态维护（Redis）                                 │
│  - 消息路由和转发                                            │
│  - 协议转换（WS ↔ Protobuf）                                │
│  - 认证和授权                                                │
│  - 限流和熔断                                                │
└─────────────────────────────────────────────────────────────┘
                          ↓ NATS JetStream
┌─────────────────────────────────────────────────────────────┐
│                     消息中间件层（NATS）                     │
│  - 消息分发和路由                                            │
│  - 消息持久化（离线消息）                                    │
│  - Pub/Sub 模式                                              │
│  - Stream 存储（7天）                                        │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌──────────────────┬──────────────────┬───────────────────────┐
│  Message Service │  Conversation Service │   Push Service        │
│  消息存储         │  会话管理        │   离线推送            │
└──────────────────┴──────────────────┴───────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│                      存储层                                  │
│  PostgreSQL (消息、会话)  Redis (缓存、在线状态)            │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 核心组件职责

#### Realtime Service（连接服务）
```go
// 核心职责
1. WebSocket 连接管理
   - 连接建立和保活（心跳检测）
   - 连接断开重连（指数退避）
   - 连接状态维护（内存 + Redis）

2. 用户在线状态管理
   - 登录状态维护（user_id -> connection）
   - 多端在线管理（同一用户多设备）
   - 在线状态广播（好友可见）
   - 离线状态标记

3. 消息路由
   - 单聊消息路由（点对点）
   - 群聊消息路由（1:N 扇出）
   - 系统通知路由（事件驱动）

4. 协议转换
   - WebSocket JSON/Protobuf ↔ NATS Message
   - 消息格式统一化
   - 消息加密解密（可选）

5. 质量保证
   - 消息去重（基于 message_id）
   - 消息顺序保证（序列号）
   - 消息可靠投递（ACK 机制）
   - 限流和熔断
```

#### Message Service（消息服务）
```go
// 核心职责
1. 消息持久化
   - 消息写入 PostgreSQL
   - 消息写入 NATS JetStream（作为消息队列）
   - 消息索引维护（ElasticSearch，可选）

2. 消息读取
   - 历史消息查询（分页）
   - 消息搜索（全文搜索）
   - 消息状态查询

3. 消息状态管理
   - 消息送达状态
   - 消息已读状态
   - 消息撤回
   - 消息编辑

4. 离线消息处理
   - 离线消息存储
   - 离线消息推送（通过 Push Service）
   - 离线消息拉取（增量同步）
```

#### Conversation Service（会话服务）
```go
// 核心职责
1. 会话列表管理
   - 会话创建和删除
   - 会话列表查询
   - 会话排序（置顶 + 时间）

2. 未读数管理
   - 单个会话未读数
   - 总未读数统计
   - 未读数清零
   - @我的消息未读数

3. 会话状态
   - 会话免打扰
   - 会话置顶
   - 最后消息显示
```

---

## 3. 消息类型设计

### 3.1 两种消息类型

#### A. 用户 IM 消息（User Messages）
**定义**：用户主动发送的聊天消息

**类型**：
- 文本消息
- 图片消息
- 视频消息
- 语音消息
- 文件消息
- 表情消息
- 名片消息
- 地理位置消息
- 合并转发消息

**流程**：
```
发送者 → Realtime → Message Service → NATS → Realtime → 接收者
                      ↓
                  PostgreSQL（持久化）
```

#### B. 系统通知消息（System Notifications）
**定义**：服务器主动推送给客户端的事件通知

**类型**：
1. **好友相关**
   - 收到好友申请
   - 好友申请被接受/拒绝
   - 好友信息更新
   - 好友删除

2. **群组相关**
   - 群邀请通知
   - 入群申请通知
   - 群成员加入/退出
   - 群信息变更
   - 群公告发布
   - 群禁言通知
   - 角色变更通知

3. **用户相关**
   - 账号登录异常
   - 密码修改通知
   - 被强制下线

4. **会话相关**
   - 消息已读回执
   - 正在输入提示
   - 消息撤回通知

**流程**：
```
业务服务 → NATS（Publish Event）→ Realtime（Subscribe）→ WebSocket → 客户端
```

### 3.2 消息格式统一化

#### WebSocket 消息格式
```json
{
  "type": "message|notification|ack|ping|pong",
  "version": "1.0",
  "messageId": "unique-message-id",
  "timestamp": 1234567890,
  "data": {
    // 根据 type 不同，data 结构不同
  }
}
```

#### 用户消息格式（type: message）
```json
{
  "type": "message",
  "messageId": "msg-123456",
  "timestamp": 1234567890,
  "data": {
    "conversationType": "private|group",
    "conversationId": "conv-123",
    "senderId": "user-123",
    "senderName": "张三",
    "senderAvatar": "https://...",
    "contentType": "text|image|video|audio|file",
    "content": {
      // 根据 contentType 不同
      "text": "你好",  // 文本消息
      "url": "...",    // 媒体消息
      "duration": 30   // 语音/视频时长
    },
    "atUsers": ["user-456"],  // @的用户（群聊）
    "replyTo": "msg-789",     // 回复的消息ID
    "localId": "local-abc"    // 客户端本地ID（用于关联）
  }
}
```

#### 系统通知格式（type: notification）
```json
{
  "type": "notification",
  "messageId": "notif-123456",
  "timestamp": 1234567890,
  "data": {
    "notificationType": "friend_request|group_invite|...",
    "title": "好友申请",
    "content": "张三请求添加您为好友",
    "actionType": "accept_reject|view|none",
    "actionData": {
      "requestId": "req-123",
      "userId": "user-123"
    },
    "priority": "high|normal|low",
    "sound": true,
    "badge": 1
  }
}
```

#### ACK 确认格式（type: ack）
```json
{
  "type": "ack",
  "messageId": "msg-123456",
  "timestamp": 1234567890,
  "data": {
    "originalMessageId": "msg-123456",
    "status": "received|delivered|read",
    "error": null
  }
}
```

---

## 4. NATS JetStream 配置

### 4.1 Stream 配置

#### 用户消息 Stream
```yaml
name: USER_MESSAGES
subjects:
  - msg.private.*    # 单聊消息：msg.private.{conversationId}
  - msg.group.*      # 群聊消息：msg.group.{groupId}
storage: file
retention: limits
max_msgs: 10000000   # 最多1000万条消息
max_bytes: 100GB     # 最多100GB
max_age: 168h        # 7天（604800秒）
max_msg_size: 10MB   # 单条消息最大10MB
discard: old         # 超过限制丢弃旧消息
```

#### 系统通知 Stream
```yaml
name: SYSTEM_NOTIFICATIONS
subjects:
  - notif.friend.*   # 好友通知
  - notif.group.*    # 群组通知
  - notif.user.*     # 用户通知
  - notif.system.*   # 系统通知
storage: file
retention: limits
max_msgs: 1000000
max_bytes: 10GB
max_age: 72h         # 3天
max_msg_size: 1MB
discard: old
```

#### 在线状态 Stream（临时）
```yaml
name: ONLINE_STATUS
subjects:
  - status.online.*
  - status.offline.*
  - status.typing.*
storage: memory      # 内存存储（不需要持久化）
retention: limits
max_msgs: 100000
max_age: 5m          # 5分钟
discard: old
```

### 4.2 Consumer 配置

#### Realtime 消费者（推送给在线用户）
```yaml
durable_name: gateway-consumer
deliver_policy: new          # 只消费新消息
ack_policy: explicit         # 显式ACK
ack_wait: 30s               # 30秒未ACK重发
max_deliver: 3              # 最多重试3次
filter_subject: msg.*        # 订阅所有消息
```

#### Message Service 消费者（持久化）
```yaml
durable_name: message-service-consumer
deliver_policy: all          # 消费所有消息
ack_policy: explicit
ack_wait: 60s
max_deliver: 5
filter_subject: msg.*
```

---

## 5. 消息流程设计

### 5.1 用户发送消息流程

```
┌─────────┐     ①     ┌─────────┐     ②     ┌─────────┐
│ 客户端A  │ ────────→ │ Realtime │ ────────→ │ Message │
│ (发送者) │  WebSocket │ Service │   gRPC    │ Service │
└─────────┘           └─────────┘           └─────────┘
                           │                      │
                           │ ③                    │ ④
                           ↓                      ↓
                      ┌─────────┐           ┌──────────┐
                      │  NATS   │←──────────│PostgreSQL│
                      │JetStream│   Publish  └──────────┘
                      └─────────┘
                           │
                           │ ⑤ Subscribe
                           ↓
                      ┌─────────┐
                      │ Realtime│
                      │ Service │
                      └─────────┘
                           │
                           │ ⑥ WebSocket
                           ↓
                      ┌─────────┐
                      │ 客户端B  │
                      │ (接收者) │
                      └─────────┘
```

#### 详细步骤

**① 客户端发送消息**
```json
{
  "type": "message",
  "localId": "local-abc-123",  // 客户端生成的本地ID
  "data": {
    "conversationType": 1,
    "conversationId": "conv-user123-user456",
    "receiverId": "user-456",
    "contentType": 1,
    "content": {"text": "你好"}
  }
}
```

**② Realtime 调用 Message Service (gRPC)**
```protobuf
message SendMessageRequest {
  string sender_id = 1;
  string conversation_id = 2;
  int32 conversation_type = 3;  // 1-single/2-group
  int32 content_type = 4;       // 1-text/2-image/3-video/4-audio/5-file/6-location/7-card
  string content = 5;  // JSON string
  repeated string at_users = 6;
  string reply_to = 7;
  string local_id = 8;
}

message SendMessageResponse {
  string message_id = 1;
  int64 timestamp = 2;
  int64 sequence = 3;
}
```

**③ Message Service 处理**
```go
func (s *MessageService) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
    // 1. 生成全局唯一消息ID
    messageId := generateMessageID()

    // 2. 生成会话内递增序列号
    sequence := s.getNextSequence(req.ConversationId)

    // 3. 写入 PostgreSQL（异步）
    go s.saveToDatabase(messageId, req, sequence)

    // 4. 发布到 NATS
    msg := &Message{
        MessageId:      messageId,
        ConversationId: req.ConversationId,
        SenderId:       req.SenderId,
        Content:        req.Content,
        Timestamp:      time.Now().Unix(),
        Sequence:       sequence,
    }

    subject := fmt.Sprintf("msg.%s.%s", req.ConversationType, req.ConversationId)
    s.nats.Publish(subject, msg)

    // 5. 更新会话最后消息（Conversation Service）
    s.sessionClient.UpdateLastMessage(req.ConversationId, messageId)

    return &SendMessageResponse{
        MessageId: messageId,
        Timestamp: msg.Timestamp,
        Sequence:  sequence,
    }, nil
}
```

**④ PostgreSQL 存储**
```sql
-- messages 表（按月分表）
CREATE TABLE messages_202602 (
    id BIGSERIAL PRIMARY KEY,
    message_id VARCHAR(64) NOT NULL UNIQUE,
    conversation_id VARCHAR(64) NOT NULL,
    conversation_type SMALLINT NOT NULL, -- 1-single/2-group
    sender_id VARCHAR(36) NOT NULL,
    content_type SMALLINT NOT NULL,      -- 1-text/2-image/3-video/4-audio/5-file/6-location/7-card
    content JSONB NOT NULL,
    sequence BIGINT NOT NULL,
    status SMALLINT DEFAULT 0,  -- 0-正常 1-撤回 2-删除
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_conversation_sequence (conversation_id, sequence),
    INDEX idx_sender_time (sender_id, created_at)
);
```

**⑤ Realtime 订阅 NATS 并路由**
```go
// Realtime 启动时订阅所有消息
func (g *Realtime) SubscribeMessages() {
    // 订阅私聊消息
    g.nats.Subscribe("msg.private.*", func(msg *nats.Msg) {
        message := parseMessage(msg.Data)
        g.routePrivateMessage(message)
    })

    // 订阅群聊消息
    g.nats.Subscribe("msg.group.*", func(msg *nats.Msg) {
        message := parseMessage(msg.Data)
        g.routeGroupMessage(message)
    })
}

// 路由私聊消息
func (g *Realtime) routePrivateMessage(msg *Message) {
    // 获取接收者的在线连接
    conns := g.getOnlineConnections(msg.ReceiverId)

    // 发送到所有在线设备
    for _, conn := range conns {
        conn.SendMessage(msg)
    }

    // 如果用户不在线，触发离线推送
    if len(conns) == 0 {
        g.pushService.SendOfflinePush(msg.ReceiverId, msg)
    }
}

// 路由群聊消息（扇出）
func (g *Realtime) routeGroupMessage(msg *Message) {
    // 获取群成员列表（缓存）
    members := g.getGroupMembers(msg.ConversationId)

    // 批量发送
    for _, memberId := range members {
        if memberId == msg.SenderId {
            continue  // 跳过发送者
        }

        conns := g.getOnlineConnections(memberId)
        for _, conn := range conns {
            conn.SendMessage(msg)
        }
    }
}
```

**⑥ 客户端接收消息并发送 ACK**
```json
// 接收消息
{
  "type": "message",
  "messageId": "msg-123456",
  "data": {...}
}

// 发送 ACK
{
  "type": "ack",
  "data": {
    "messageId": "msg-123456",
    "status": "received"
  }
}
```

### 5.2 系统通知流程

```
┌──────────┐     ①     ┌─────────┐     ②     ┌─────────┐
│ Business │ ────────→ │  NATS   │ ────────→ │ Realtime │
│ Service  │  Publish  │JetStream│ Subscribe │ Service │
└──────────┘           └─────────┘           └─────────┘
                                                   │
                                                   │ ③
                                                   ↓
                                              ┌─────────┐
                                              │  Client │
                                              └─────────┘
```

#### 详细步骤

**① 业务服务发布事件**
```go
// Friend Service - 发送好友申请后
func (s *FriendService) SendFriendRequest(...) error {
    // ... 业务逻辑

    // 发布好友申请事件
    event := &FriendRequestEvent{
        Type:       "friend_request",
        RequestId:  requestId,
        FromUserId: fromUserId,
        ToUserId:   toUserId,
        Message:    message,
        Timestamp:  time.Now().Unix(),
    }

    subject := fmt.Sprintf("notif.friend.request.%s", toUserId)
    s.nats.Publish(subject, event)

    return nil
}
```

**② Realtime 订阅并转换**
```go
func (g *Realtime) SubscribeNotifications() {
    // 订阅好友通知
    g.nats.Subscribe("notif.friend.*", func(msg *nats.Msg) {
        event := parseFriendEvent(msg.Data)
        notification := g.convertToNotification(event)
        g.sendToUser(event.ToUserId, notification)
    })

    // 订阅群组通知
    g.nats.Subscribe("notif.group.*", func(msg *nats.Msg) {
        event := parseGroupEvent(msg.Data)
        notification := g.convertToNotification(event)

        // 群通知可能需要发给多个用户
        for _, userId := range event.TargetUsers {
            g.sendToUser(userId, notification)
        }
    })
}

func (g *Realtime) convertToNotification(event interface{}) *Notification {
    // 根据事件类型转换为统一的通知格式
    switch e := event.(type) {
    case *FriendRequestEvent:
        return &Notification{
            Type:             "notification",
            NotificationType: "friend_request",
            Title:            "好友申请",
            Content:          fmt.Sprintf("%s请求添加您为好友", e.FromUserName),
            ActionType:       "accept_reject",
            ActionData:       map[string]interface{}{"requestId": e.RequestId},
            Priority:         "high",
            Sound:            true,
        }
    // ... 其他事件类型
    }
}
```

**③ 推送到客户端**
```json
{
  "type": "notification",
  "messageId": "notif-123456",
  "timestamp": 1234567890,
  "data": {
    "notificationType": "friend_request",
    "title": "好友申请",
    "content": "张三请求添加您为好友",
    "actionType": "accept_reject",
    "actionData": {
      "requestId": "req-123",
      "userId": "user-123",
      "userName": "张三",
      "userAvatar": "https://..."
    }
  }
}
```

---

## 6. 关键技术实现

### 6.1 连接管理

#### 连接状态存储（Redis）
```go
// 在线连接映射
type ConnectionManager struct {
    redis  *redis.Client
    local  sync.Map  // 本地缓存，加速查询
}

// 用户登录
func (cm *ConnectionManager) OnUserLogin(userId, deviceId, gatewayId string) {
    key := fmt.Sprintf("online:%s:%s", userId, deviceId)

    // Redis 存储（带过期时间）
    cm.redis.HSet(ctx, key, map[string]interface{}{
        "gateway_id": gatewayId,
        "login_time": time.Now().Unix(),
    })
    cm.redis.Expire(ctx, key, 24*time.Hour)

    // 本地缓存
    cm.local.Store(key, &Connection{...})

    // 发布在线状态事件
    cm.publishOnlineStatus(userId, true)
}

// 查询用户所有在线设备
func (cm *ConnectionManager) GetOnlineDevices(userId string) []string {
    pattern := fmt.Sprintf("online:%s:*", userId)
    keys, _ := cm.redis.Keys(ctx, pattern).Result()

    var devices []string
    for _, key := range keys {
        parts := strings.Split(key, ":")
        devices = append(devices, parts[2])  // 提取 deviceId
    }
    return devices
}

// 心跳保活
func (cm *ConnectionManager) Heartbeat(userId, deviceId string) {
    key := fmt.Sprintf("online:%s:%s", userId, deviceId)
    cm.redis.Expire(ctx, key, 24*time.Hour)
}
```

#### 连接路由表
```go
// 当前 Realtime 实例的连接映射
type LocalConnectionManager struct {
    connections sync.Map  // map[userId]map[deviceId]*WebSocketConn
}

func (lcm *LocalConnectionManager) AddConnection(userId, deviceId string, conn *websocket.Conn) {
    userConns, _ := lcm.connections.LoadOrStore(userId, &sync.Map{})
    userConns.(*sync.Map).Store(deviceId, conn)
}

func (lcm *LocalConnectionManager) GetConnections(userId string) []*websocket.Conn {
    userConns, ok := lcm.connections.Load(userId)
    if !ok {
        return nil
    }

    var conns []*websocket.Conn
    userConns.(*sync.Map).Range(func(key, value interface{}) bool {
        conns = append(conns, value.(*websocket.Conn))
        return true
    })
    return conns
}
```

### 6.2 消息去重

#### 基于 messageId 去重
```go
type DuplicateChecker struct {
    redis *redis.Client
}

func (dc *DuplicateChecker) IsDuplicate(messageId string) bool {
    key := fmt.Sprintf("msg:dup:%s", messageId)

    // SETNX 原子操作
    exists, _ := dc.redis.SetNX(ctx, key, "1", 10*time.Minute).Result()
    return !exists  // true 表示已存在（重复）
}

// 客户端发送消息时
func (g *Realtime) HandleMessage(conn *websocket.Conn, msg *Message) {
    if g.duplicateChecker.IsDuplicate(msg.MessageId) {
        // 重复消息，直接返回成功（幂等）
        conn.SendAck(msg.MessageId, "success")
        return
    }

    // 处理消息...
}
```

### 6.3 消息顺序保证

#### 会话序列号
```go
// 使用 Redis INCR 生成递增序列号
func (s *MessageService) GetNextSequence(conversationId string) int64 {
    key := fmt.Sprintf("seq:%s", conversationId)
    seq, _ := s.redis.Incr(ctx, key).Result()
    return seq
}

// 客户端接收消息时检查序列号
type MessageQueue struct {
    expectedSeq int64
    buffer      map[int64]*Message  // 乱序消息缓冲
}

func (mq *MessageQueue) OnMessage(msg *Message) {
    if msg.Sequence == mq.expectedSeq {
        // 按序到达，直接处理
        mq.handleMessage(msg)
        mq.expectedSeq++

        // 检查缓冲区是否有后续消息
        mq.processBufferedMessages()
    } else if msg.Sequence > mq.expectedSeq {
        // 乱序到达，缓存
        mq.buffer[msg.Sequence] = msg

        // 触发补齐（如果缺失太多）
        if msg.Sequence - mq.expectedSeq > 10 {
            mq.requestMissingMessages(mq.expectedSeq, msg.Sequence)
        }
    }
}
```

### 6.4 离线消息处理

#### 离线消息存储（NATS JetStream）
```go
// NATS Consumer 配置
consumer := &nats.ConsumerConfig{
    Durable:       "user-123-offline",
    FilterSubject: "msg.*.user-123",  // 只接收该用户的消息
    DeliverPolicy: nats.DeliverAllPolicy,  // 接收所有消息
    AckPolicy:     nats.AckExplicitPolicy,
}

// 用户上线后拉取离线消息
func (g *Realtime) OnUserOnline(userId string) {
    // 1. 获取用户最后已读序列号
    lastSeq := g.getLastReadSequence(userId)

    // 2. 从 NATS 拉取离线消息
    consumer, _ := g.js.PullSubscribe("", "", nats.Bind("USER_MESSAGES", userId))

    msgs, _ := consumer.Fetch(100)  // 每次拉取100条
    for _, msg := range msgs {
        message := parseMessage(msg.Data)
        if message.Sequence > lastSeq {
            g.sendToUser(userId, message)
        }
        msg.Ack()
    }
}
```

#### 离线推送（Push Service）
```go
// 用户离线时触发推送
func (ps *PushService) SendOfflinePush(userId string, message *Message) {
    // 1. 获取用户推送 Token
    tokens := ps.getUserPushTokens(userId)

    // 2. 构造推送内容
    notification := &PushNotification{
        Title: message.SenderName,
        Body:  message.ContentPreview,
        Badge: ps.getUnreadCount(userId),
        Sound: "default",
        Data: map[string]interface{}{
            "messageId":      message.MessageId,
            "conversationId": message.ConversationId,
        },
    }

    // 3. 发送推送（FCM/APNs）
    for _, token := range tokens {
        ps.sendPush(token, notification)
    }
}
```

---

## 7. 性能优化策略

### 7.1 消息批处理

#### 批量发送（客户端 → 服务器）
```go
// 客户端累积消息，批量发送
type MessageBatcher struct {
    messages []*Message
    timer    *time.Timer
}

func (mb *MessageBatcher) AddMessage(msg *Message) {
    mb.messages = append(mb.messages, msg)

    // 达到批次大小或超时后发送
    if len(mb.messages) >= 10 || mb.timer == nil {
        mb.flush()
    }
}

func (mb *MessageBatcher) flush() {
    if len(mb.messages) == 0 {
        return
    }

    // 批量发送
    sendBatch(mb.messages)
    mb.messages = nil
}
```

#### 批量推送（服务器 → 客户端）
```go
// 群聊消息批量推送
func (g *Realtime) BatchPushGroupMessage(groupId string, msg *Message) {
    members := g.getGroupMembers(groupId)

    // 分批推送（每批1000人）
    batchSize := 1000
    for i := 0; i < len(members); i += batchSize {
        end := i + batchSize
        if end > len(members) {
            end = len(members)
        }

        batch := members[i:end]
        g.pushToBatch(batch, msg)
    }
}
```

### 7.2 消息压缩

```go
// WebSocket 消息压缩
func (conn *WebSocketConn) SendCompressed(msg *Message) error {
    data, _ := json.Marshal(msg)

    // 超过1KB启用压缩
    if len(data) > 1024 {
        var buf bytes.Buffer
        writer := gzip.NewWriter(&buf)
        writer.Write(data)
        writer.Close()

        return conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
    }

    return conn.WriteMessage(websocket.TextMessage, data)
}
```

### 7.3 缓存策略

#### 群成员列表缓存
```go
// 缓存群成员列表（减少数据库查询）
func (g *Realtime) getGroupMembers(groupId string) []string {
    key := fmt.Sprintf("group:members:%s", groupId)

    // 先查 Redis
    members, err := g.redis.SMembers(ctx, key).Result()
    if err == nil && len(members) > 0 {
        return members
    }

    // Redis 没有，查数据库
    members = g.groupService.GetMembers(groupId)

    // 写入 Redis（1小时过期）
    g.redis.SAdd(ctx, key, members)
    g.redis.Expire(ctx, key, time.Hour)

    return members
}
```

#### 用户信息缓存
```go
// 缓存用户基本信息（减少 gRPC 调用）
type UserCache struct {
    cache *lru.Cache  // LRU 本地缓存
    redis *redis.Client
}

func (uc *UserCache) GetUserInfo(userId string) *UserInfo {
    // 1. 本地缓存
    if info, ok := uc.cache.Get(userId); ok {
        return info.(*UserInfo)
    }

    // 2. Redis 缓存
    key := fmt.Sprintf("user:info:%s", userId)
    data, _ := uc.redis.Get(ctx, key).Bytes()
    if len(data) > 0 {
        info := &UserInfo{}
        json.Unmarshal(data, info)
        uc.cache.Add(userId, info)
        return info
    }

    // 3. 调用 User Service
    info := uc.userService.GetUserInfo(userId)

    // 写入缓存
    data, _ = json.Marshal(info)
    uc.redis.Set(ctx, key, data, 10*time.Minute)
    uc.cache.Add(userId, info)

    return info
}
```

### 7.4 数据库优化

#### 消息表分表策略
```sql
-- 按月分表
CREATE TABLE messages_202601 (...);
CREATE TABLE messages_202602 (...);
CREATE TABLE messages_202603 (...);

-- 分表路由逻辑
func getMessageTable(timestamp int64) string {
    t := time.Unix(timestamp, 0)
    return fmt.Sprintf("messages_%s", t.Format("200601"))
}
```

#### 索引优化
```sql
-- 会话消息查询索引（最常用）
CREATE INDEX idx_conversation_sequence ON messages (conversation_id, sequence DESC);

-- 用户消息查询索引
CREATE INDEX idx_sender_time ON messages (sender_id, created_at DESC);

-- 消息搜索索引（GIN）
CREATE INDEX idx_content_search ON messages USING gin(content);
```

---

## 8. 监控和运维

### 8.1 关键指标

#### Realtime 指标
- WebSocket 连接数（当前/峰值）
- 消息吞吐量（发送/接收 QPS）
- 消息延迟（P50/P95/P99）
- 连接成功率
- 心跳丢失率

#### NATS 指标
- Stream 消息数量
- Stream 存储大小
- Consumer Lag（消费延迟）
- 消息发布速率
- 消息消费速率

#### Message Service 指标
- 消息写入 QPS
- 消息查询 QPS
- 数据库响应时间
- 离线消息堆积量

### 8.2 告警规则

```yaml
# Prometheus 告警规则
groups:
  - name: realtime
    rules:
      - alert: HighConnectionCount
        expr: realtime_connections > 100000
        for: 5m
        annotations:
          summary: "Realtime 连接数过高"

      - alert: HighMessageLatency
        expr: histogram_quantile(0.95, realtime_message_latency_seconds) > 1
        for: 5m
        annotations:
          summary: "消息延迟超过1秒"

  - name: nats
    rules:
      - alert: HighConsumerLag
        expr: nats_consumer_lag > 10000
        for: 5m
        annotations:
          summary: "NATS 消费延迟过高"
```

---

## 9. 部署架构

### 9.1 服务部署

```yaml
# Kubernetes 部署示例
apiVersion: apps/v1
kind: Deployment
metadata:
  name: realtime-service
spec:
  replicas: 3  # 至少3个实例（高可用）
  selector:
    matchLabels:
      app: realtime
  template:
    spec:
      containers:
      - name: realtime
        image: anychat/realtime:latest
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
        env:
        - name: NATS_URL
          value: "nats://nats:4222"
        - name: REDIS_ADDR
          value: "redis:6379"
```

### 9.2 网络拓扑

```
               ┌─────────────────┐
               │   LoadBalancer  │
               │   (Nginx/F5)    │
               └────────┬────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
   ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
   │Realtime-1│    │Realtime-2│    │Realtime-3│
   └────┬────┘    └────┬────┘    └────┬────┘
        │               │               │
        └───────────────┼───────────────┘
                        │
                ┌───────▼────────┐
                │  NATS Cluster  │
                │   (3 nodes)    │
                └───────┬────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
   ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
   │Message  │    │Conversation│    │  Push   │
   │Service  │    │Service  │    │ Service │
   └─────────┘    └─────────┘    └─────────┘
```

---

## 10. 总结与建议

### 10.1 推荐方案

✅ **推荐使用：WebSocket Gateway + NATS 混合架构**

**理由**：
1. 符合业界最佳实践
2. 关注点分离，便于扩展
3. 安全性和可控性更好
4. 客户端兼容性广泛
5. 运维成本可控

### 10.2 实施路线

**Phase 1: MVP（1-2周）**
- ✅ Realtime WebSocket 基础框架
- ✅ 单聊消息收发
- ✅ 基础 NATS Pub/Sub
- ✅ 消息持久化

**Phase 2: 核心功能（2-3周）**
- ✅ 群聊消息
- ✅ 系统通知
- ✅ 离线消息
- ✅ 消息状态管理

**Phase 3: 优化（1-2周）**
- ✅ 消息批处理
- ✅ 缓存优化
- ✅ 监控告警
- ✅ 压测调优

**Phase 4: 高级特性（按需）**
- 消息搜索（ElasticSearch）
- 多媒体消息优化
- 音视频通话集成
- 消息加密

### 10.3 关键决策总结

| 决策点 | 选择 | 原因 |
|--------|------|------|
| 客户端协议 | WebSocket | 广泛支持，易于调试 |
| 消息队列 | NATS JetStream | 轻量级，支持持久化 |
| 消息存储 | PostgreSQL + NATS | 可靠性高，查询灵活 |
| 在线状态 | Redis | 快速查询，支持过期 |
| 群消息扇出 | Realtime 层 | 降低后端压力 |
| 离线推送 | Push Service | 解耦，支持多平台 |

---

**文档版本**: v1.0
**最后更新**: 2026-02-16
**作者**: Claude + AnyChat Team
