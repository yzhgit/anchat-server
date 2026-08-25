# 好友关系管理设计

## 1. 概述

好友关系管理提供好友列表获取、好友添加、删除、备注等功能。

## 2. 功能列表

- [x] 获取好友列表（支持增量同步）
- [x] 删除好友
- [x] 设置好友备注
- [x] 批量检查好友关系

## 3. 数据模型

### 3.1 Friendship 表

```go
type Friendship struct {
    ID            int64     // 主键
    UserID        string    // 用户ID
    FriendID      string    // 好友ID
    Remark        string    // 好友备注
    IsFavorite    bool      // 是否收藏
    CreatedAt     time.Time // 好友添加时间
    UpdatedAt     time.Time
}
```

## 4. 业务流程

### 4.1 获取好友列表

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant FriendService
    participant UserService
    participant DB

    Client->>Gateway: GET /friend/list?lastUpdateTime=xxx<br/>Header: Authorization: Bearer {token}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>FriendService: gRPC GetFriendList(userId, lastUpdateTime)
    FriendService->>DB: 查询好友列表
    DB-->>FriendService: 好友列表
    FriendService->>UserService: 批量获取用户信息
    FriendService-->>Gateway: 返回好友列表和同步时间
    Gateway-->>Client: 200 OK
```

### 4.2 删除好友

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant FriendService
    participant DB
    participant NATS

    Client->>Gateway: DELETE /friend/{friendId}<br/>Header: Authorization: Bearer {token}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>FriendService: gRPC DeleteFriend(userId, friendId)
    FriendService->>DB: 删除好友关系(双向)
    DB-->>FriendService: 成功
    FriendService->>NATS: 发布好友删除事件
    FriendService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

### 4.3 设置好友备注

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant FriendService
    participant DB

    Client->>Gateway: PUT /friend/remark<br/>Header: Authorization: Bearer {token}<br/>Body: {friend_id, remark}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>FriendService: gRPC UpdateRemark(userId, friendId, remark)
    FriendService->>DB: 更新好友备注
    DB-->>FriendService: 成功
    FriendService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

## 5. API设计

### 5.1 获取好友列表

```protobuf
message GetFriendListRequest {
    string user_id = 1;
    int64 last_update_time = 2; // 增量同步用
}

message GetFriendListResponse {
    repeated FriendInfo friends = 1;
    int64 sync_time = 2;
}

message FriendInfo {
    string friend_id = 1;
    string nickname = 2;
    string avatar = 3;
    string remark = 4;
    bool is_favorite = 5;
    int64 created_at = 6;
}
```

### 5.2 删除好友

```protobuf
message DeleteFriendRequest {
    string user_id = 1;
    string friend_id = 2;
}
```

### 5.3 更新备注

```protobuf
message UpdateRemarkRequest {
    string user_id = 1;
    string friend_id = 2;
    string remark = 3;
}
```

## 6. 通知主题

- `notification.friend.deleted.{user_id}` - 好友删除通知
- `notification.friend.remark_updated.{user_id}` - 好友备注修改通知
