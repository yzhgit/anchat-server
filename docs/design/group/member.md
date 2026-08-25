# 群成员管理设计

## 1. 概述

群成员管理处理成员邀请、移除、权限变更等操作。

## 2. 功能列表

- [x] 邀请成员
- [x] 移除成员
- [x] 退出群组
- [x] 设置/取消管理员
- [x] 群主转让
- [x] 更新成员昵称
- [x] **成员禁言** (本文档新增)

## 3. 业务流程

### 3.1 邀请成员

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant GroupService
    participant MessageService
    participant DB
    participant NATS

    Client->>Gateway: POST /group/{groupId}/members/invite<br/>Header: Authorization: Bearer {token}<br/>Body: {member_ids}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>GroupService: gRPC InviteMembers(userId, groupId, memberIds)
    GroupService->>GroupService: 验证权限(群主/管理员)
    GroupService->>DB: 查询已存在成员
    GroupService->>DB: 创建新成员记录
    GroupService->>MessageService: 发送入群系统消息
    GroupService->>NATS: 发布成员邀请事件
    GroupService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

### 3.2 移除成员

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant GroupService
    participant MessageService
    participant DB
    participant NATS

    Client->>Gateway: DELETE /group/{groupId}/members/{targetUserId}<br/>Header: Authorization: Bearer {token}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>GroupService: gRPC RemoveMember(userId, groupId, targetUserId)
    GroupService->>GroupService: 验证权限(群主/管理员)
    GroupService->>DB: 删除成员记录
    GroupService->>MessageService: 发送被移除系统消息
    GroupService->>NATS: 发布成员移除事件
    GroupService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

### 3.3 退出群组

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant GroupService
    participant MessageService
    participant DB
    participant NATS

    Client->>Gateway: DELETE /group/{groupId}/members/me<br/>Header: Authorization: Bearer {token}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>GroupService: gRPC QuitGroup(userId, groupId)
    GroupService->>GroupService: 检查是否为群主(群主不能退出)
    GroupService->>DB: 删除成员记录
    GroupService->>MessageService: 发送退群系统消息
    GroupService->>NATS: 发布退群事件
    GroupService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

### 3.4 设置/取消管理员

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant GroupService
    participant DB
    participant NATS

    Client->>Gateway: PUT /group/{groupId}/members/{targetUserId}/role<br/>Header: Authorization: Bearer {token}<br/>Body: {role: 0/1}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>GroupService: gRPC UpdateMemberRole(userId, groupId, targetUserId, role)
    GroupService->>GroupService: 验证权限(仅群主)
    GroupService->>DB: 更新成员角色
    GroupService->>NATS: 发布角色变更事件
    GroupService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

### 3.5 禁言/解除禁言

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant GroupService
    participant MessageService
    participant DB
    participant NATS

    Client->>Gateway: PUT /group/{groupId}/members/{targetUserId}/mute<br/>Header: Authorization: Bearer {token}<br/>Body: {type: permanent/temporary, duration_minutes: 60}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>GroupService: gRPC MuteMember(userId, groupId, targetUserId, type, durationMinutes)
    GroupService->>GroupService: 验证权限(群主/管理员)
    GroupService->>GroupService: 计算禁言截止时间
    GroupService->>DB: 更新 MutedUntil 字段
    GroupService->>MessageService: 发送禁言系统消息
    GroupService->>NATS: 发布禁言通知
    GroupService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

### 3.6 解除禁言

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant GroupService
    participant DB
    participant NATS

    Client->>Gateway: DELETE /group/{groupId}/members/{targetUserId}/mute<br/>Header: Authorization: Bearer {token}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>GroupService: gRPC UnmuteMember(userId, groupId, targetUserId)
    GroupService->>GroupService: 验证权限(群主/管理员)
    GroupService->>DB: 清空 MutedUntil 字段
    GroupService->>NATS: 发布解除禁言通知
    GroupService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

## 4. API设计

### 4.1 邀请成员

```protobuf
message InviteMembersRequest {
    string user_id = 1;
    string group_id = 2;
    repeated string member_ids = 3;
}
```

### 4.2 移除成员

```protobuf
message RemoveMemberRequest {
    string user_id = 1;
    string group_id = 2;
    string target_user_id = 3;
}
```

### 4.3 退出群组

```protobuf
message QuitGroupRequest {
    string user_id = 1;
    string group_id = 2;
}
```

### 4.4 更新成员角色

```protobuf
message UpdateMemberRoleRequest {
    string user_id = 1;
    string group_id = 2;
    string target_user_id = 3;
    GroupRole role = 4; // 2-admin 3-member
}
```

### 4.5 禁言/解除禁言 (新增)

```protobuf
// 禁言类型枚举
enum MuteType {
    PERMANENT = 0;  // 永久禁言
    TEMPORARY = 1;  // 限时禁言
}

// 禁言成员
message MuteMemberRequest {
    string user_id = 1;           // 操作者ID
    string group_id = 2;          // 群ID
    string target_user_id = 3;    // 被禁言用户ID
    MuteType type = 4;            // 禁言类型
    int32 duration_minutes = 5;  // 限时禁言时长（分钟），仅 TEMPORARY 有效
                                  // 常用值: 60(1小时), 1440(1天), 10080(7天), 43200(30天)
}

// 解除禁言
message UnmuteMemberRequest {
    string user_id = 1;           // 操作者ID
    string group_id = 2;          // 群ID
    string target_user_id = 3;    // 被解除禁言用户ID
}
```

### 4.6 客户端禁言时间判断 (新增)

```go
// 永久禁言时间常量
var PermMutedUntil = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

// 判断是否为永久禁言
func IsPermanentlyMuted(mutedUntil *time.Time) bool {
    if mutedUntil == nil {
        return false
    }
    return mutedUntil.Equal(PermMutedUntil) || mutedUntil.After(PermMutedUntil)
}

// 判断当前是否处于禁言状态
func IsMuted(mutedUntil *time.Time) bool {
    if mutedUntil == nil {
        return false
    }
    return mutedUntil.After(time.Now())
}
```

## 5. 权限规则

| 操作 | 群主 | 管理员 | 成员 |
|------|------|--------|------|
| 邀请成员 | ✓ | ✓ | ✗ |
| 移除成员 | ✓ | ✓(仅成员) | ✗ |
| 设置管理员 | ✓ | ✗ | ✗ |
| 转让群主 | ✓ | ✗ | ✗ |
| 修改群信息 | ✓ | ✗ | ✗ |
| 解散群组 | ✓ | ✗ | ✗ |
| 禁言成员 | ✓ | ✓ | ✗ |
| 解除禁言 | ✓ | ✓ | ✗ |

## 6. 通知主题

- `notification.group.member_muted.{group_id}` - 成员被禁言
- `notification.group.member_unmuted.{group_id}` - 成员解除禁言
- `notification.group.role_changed.{group_id}` - 成员角色变更
