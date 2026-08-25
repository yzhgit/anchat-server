# Token 管理设计

## 1. 概述

Token 管理负责生成、验证、刷新访问令牌，支持多设备登录场景。

## 2. 功能列表

- [x] AccessToken 生成与验证
- [x] RefreshToken 生成与验证
- [x] Token 刷新机制
- [x] Token 有效期管理

## 3. Token 规格

| Token 类型 | 有效期 | 用途 |
|-----------|--------|------|
| AccessToken | 2小时 | API 访问授权 |
| RefreshToken | 7天 | 刷新 AccessToken |

## 4. Token 结构

### 4.1 AccessToken Claims

```go
type Claims struct {
    UserID    string `json:"user_id"`
    DeviceID  string `json:"device_id"`
    DeviceType int16  `json:"device_type"` // 1-ios 2-android 3-web 4-pc 5-h5
    Exp       int64  `json:"exp"`
    Iat       int64  `json:"iat"`
}
```

### 4.2 Token 生成流程

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant UserService
    participant JWTManager
    participant DB

    Client->>Gateway: POST /user/login<br/>Body: {account, password, device_id, device_type}
    Gateway->>UserService: gRPC Login
    UserService->>JWTManager: GenerateAccessToken(userID, deviceID, deviceType)
    JWTManager->>JWTManager: 生成JWT签名(RS256)
    JWTManager-->>UserService: token字符串
    UserService->>JWTManager: GenerateRefreshToken(userID, deviceID, deviceType)
    JWTManager-->>UserService: refreshToken字符串
    UserService->>DB: 保存会话
    UserService-->>Gateway: 返回Token
    Gateway-->>Client: 200 OK
```

## 5. Token 刷新

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant UserService
    participant JWTManager
    participant DB

    Client->>Gateway: POST /user/refresh<br/>Body: {refresh_token}
    Gateway->>UserService: gRPC RefreshToken(refreshToken)
    UserService->>JWTManager: 验证RefreshToken
    JWTManager-->>UserService: claims
    UserService->>DB: 查询会话
    DB-->>UserService: 会话信息
    UserService->>UserService: 检查RefreshToken过期
    UserService->>JWTManager: 生成新AccessToken + RefreshToken
    UserService->>DB: 更新会话
    DB-->>UserService: 成功
    UserService-->>Gateway: 返回新Tokens
    Gateway-->>Client: 200 OK
```

## 6. Token 验证

Gateway 本地验证 JWT，流程如下：

1. Client 携带 `Authorization: Bearer {access_token}` 请求 API
2. Gateway 本地验证 JWT 签名与有效期
3. Gateway 解析 claims，提取 `userId`
4. Gateway 将 `userId` 放置到请求上下文，后续 gRPC 调用自动携带

## 7. 错误码

| 错误码 | 说明 |
|--------|------|
| 10107 | RefreshToken无效 |
| 10108 | RefreshToken已过期 |

## 8. 安全考虑

1. **签名算法**: RS256 (非对称加密)
2. **密钥管理**: 配置文件或密钥管理系统
3. **Token 存储**: 会话表存储 Token 映射
4. **登出处理**: 删除会话记录使 Token 失效
