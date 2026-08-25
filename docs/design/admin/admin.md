# 管理后台设计

## 1. 概述

Admin服务提供后台管理功能，包括管理员管理、用户管理、群组管理、系统配置、审计日志等。

## 2. 功能列表

### 2.1 管理员管理
- [x] 管理员登录
- [x] 管理员列表
- [x] 创建管理员
- [x] 禁用/启用管理员
- [x] 重置密码

### 2.2 用户管理
- [x] 搜索用户
- [x] 查看用户详情
- [x] 禁用/启用用户
- [x] 封禁/解封用户

### 2.3 群组管理
- [x] 查看群组信息
- [x] 解散群组

### 2.4 系统管理
- [x] 系统统计
- [x] 系统配置管理
- [x] 审计日志

### 2.5 日志管理
- [x] 上传日志文件
- [x] 下载日志文件
- [x] 获取日志列表
- [x] 删除日志文件

## 3. 管理员角色

| 枚举值 | 角色 | 说明 |
|--------|------|------|
| 1 | superadmin | 超级管理员 |
| 2 | admin | 普通管理员 |
| 3 | readonly | 只读管理员 |

## 4. 数据模型

### 4.1 AdminUser 表

```go
type AdminUser struct {
    ID        string    // 管理员ID
    Username  string    // 用户名
    Password  string    // 密码哈希
    Role      int16     // 角色: 1-superadmin 2-admin 3-readonly
    Status    int8      // 状态: 1-正常 2-禁用
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 4.2 AuditLog 表

```go
type AuditLog struct {
    ID          string    // 日志ID
    AdminID     string    // 管理员ID
    Action      string    // 操作类型
    TargetType  string    // 目标类型
    TargetID    string    // 目标ID
    Details     string    // 详情(JSON)
    IPAddress   string    // IP地址
    CreatedAt   time.Time
}
```

### 4.3 SystemConfig 表

```go
type SystemConfig struct {
    ID        string    // 配置Key
    Value     string    // 配置值
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## 5. 业务流程

### 5.1 管理员登录

```mermaid
sequenceDiagram
    participant Admin
    participant Gateway
    participant AdminService
    participant DB

    Admin->>Gateway: POST /admin/login<br/>Body: {username, password}
    Gateway->>AdminService: gRPC Login(username, password, ip)
    AdminService->>DB: 查询管理员
    AdminService->>AdminService: 验证密码
    AdminService->>AdminService: 生成JWT Token
    AdminService->>DB: 更新最后登录时间
    AdminService-->>Gateway: Token + 管理员信息
    Gateway-->>Admin: 登录成功
```

### 5.2 用户管理

```mermaid
sequenceDiagram
    participant Admin
    participant Gateway
    participant AdminService
    participant DB

    Admin->>Gateway: GET /admin/users?keyword=xxx<br/>Header: Authorization: Bearer {token}
    Gateway->>AdminService: gRPC SearchUsers(keyword, page, pageSize)
    AdminService->>DB: 搜索用户
    AdminService-->>Gateway: 用户列表
    Gateway-->>Admin: 200 OK
```

### 5.3 禁用/启用用户

```mermaid
sequenceDiagram
    participant Admin
    participant Gateway
    participant AdminService
    participant UserService
    participant DB
    participant NATS

    Admin->>Gateway: PUT /admin/user/{userId}/status<br/>Header: Authorization: Bearer {token}<br/>Body: {status: disabled}
    Gateway->>AdminService: gRPC UpdateUserStatus(userId, status)
    AdminService->>UserService: 禁用用户
    UserService->>DB: 更新用户状态
    UserService->>NATS: 发布强制下线通知
    AdminService-->>Gateway: 成功
    Gateway-->>Admin: 200 OK
```
