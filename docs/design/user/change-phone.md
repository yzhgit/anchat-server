# 更换手机号设计

## 1. 概述

更换手机号功能允许用户将已绑定的手机号更换为新的手机号。更换手机号需要先验证旧手机号（或通过其他方式验证身份），确保操作安全性。

## 2. 功能列表

- [x] 更换手机号（需验证码）
- [x] 更换前检查新手机号是否已被占用
- [x] 更换后使旧手机号相关的会话失效

## 3. 业务流程

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant UserService
    participant UserService
    participant Redis
    participant DB
    participant NATS

    Client->>Gateway: POST /user/phone/change<br/>Header: Authorization: Bearer {token}<br/>Body: {old_phone_number, new_phone_number, new_verify_code, old_verify_code}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>UserService: gRPC ChangePhone(userId, oldPhoneNumber, newPhoneNumber, newVerifyCode, oldVerifyCode)
    UserService->>UserService: 验证旧手机号是否匹配
    UserService->>UserService: 检查新手机号格式
    UserService->>DB: 检查新手机号是否已被占用
    UserService->>UserService: 验证新手机号验证码(newPhoneNumber, verifyCode, change_phone)
    UserService->>Redis: 验证验证码
    Redis-->>UserService: 验证成功
    UserService-->>UserService: 验证成功
    UserService->>DB: 更新用户手机号
    UserService->>NATS: 发布手机号更换事件
    NATS-->>UserService: 触发强制下线
    UserService-->>UserService: 发布强制下线通知
    UserService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

## 4. 验证规则

| 字段 | 规则 |
|------|------|
| 新手机号 | 11位数字，以1开头 |
| 验证码 | 6位数字，有效期5分钟 |
| 绑定检查 | 新手机号未被其他用户占用 |
| 旧手机号验证 | 需验证旧手机号验证码或使用其他身份验证方式 |

## 5. API设计

### 5.1 请求

```protobuf
message ChangePhoneRequest {
    string user_id = 1;
    string old_phone_number = 2;    // 旧手机号
    string new_phone_number = 3;     // 新手机号
    string new_verify_code = 4;     // 新手机号验证码
    string old_verify_code = 5;      // 旧手机号验证码（可选，用于验证身份）
}
```

### 5.2 响应

```protobuf
message ChangePhoneResponse {
    string old_phone_number = 1;    // 旧手机号（脱敏）
    string new_phone_number = 2;    // 新手机号
}
```

### 5.3 错误码

| 错误码 | 说明 |
|--------|------|
| 10206 | 验证码错误 |
| 10207 | 验证码已过期 |
| 20106 | 新手机号格式错误 |
| 20107 | 新手机号已被占用 |
| 20110 | 旧手机号不匹配 |
| 10104 | 用户不存在 |

## 6. 安全考虑

1. **旧手机号验证**：必须验证旧手机号或使用其他身份验证方式
2. **验证码一次性使用**：验证成功后立即失效
3. **手机号唯一性**：同一手机号只能绑定一个账号
4. **强制下线**：更换手机号后，其他设备需要重新登录
5. **安全通知**：更换成功后发送通知到旧手机号

## 7. 依赖服务

- **User Service**: 验证码验证（认证功能已合并到 User 服务）
- **Verify Service**: 验证码校验与消费
- **Conversation Service**: 会话管理
- **PostgreSQL**: 用户信息持久化

---

返回: [用户服务总体设计](README.md)
