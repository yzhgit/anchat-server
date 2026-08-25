# User Service (用户服务)

## 1. 服务概述

**职责**: 用户注册、登录、Token管理、多端登录策略; 用户资料管理、个人设置、二维码、推送Token管理

**核心功能**:
- 用户注册（手机号/邮箱，需验证码）
- 用户登录（账号密码/验证码）
- Token管理（JWT: AccessToken + RefreshToken）
- 多端登录策略与设备管理
- 用户资料管理（获取、修改、搜索）
- 个人设置（验证开关、通知设置、隐私设置）
- 二维码功能（生成、刷新、扫码）
- 推送Token管理
- 用户状态（在线状态、最后活跃时间）

## 2. 文档导航

| 功能 | 文档 | 说明 |
|------|------|------|
| 用户注册 | [register.md](register.md) | 手机号/邮箱注册 |
| 用户登录 | [login.md](login.md) | 登录方式与流程 |
| Token管理 | [token.md](token.md) | JWT令牌管理 |
| 会话管理 | [session.md](session.md) | 用户会话 |
| 设备管理 | [device.md](device.md) | 设备登录记录 |
| 密码管理 | [password.md](password.md) | 修改/重置密码 |
| 验证码 | [verification-code.md](verification-code.md) | 验证码发送与验证 |
| 绑定手机号 | [bind-phone.md](bind-phone.md) | 绑定手机号 |
| 更换手机号 | [change-phone.md](change-phone.md) | 更换手机号 |
| 绑定邮箱 | [bind-email.md](bind-email.md) | 绑定邮箱 |
| 更换邮箱 | [change-email.md](change-email.md) | 更换邮箱 |
| 用户资料 | [profile.md](profile.md) | 获取/修改/搜索用户资料 |
| 个人设置 | [settings.md](settings.md) | 用户设置管理 |
| 二维码 | [qrcode.md](qrcode.md) | 二维码生成与扫码 |
| 推送Token | [push-token.md](push-token.md) | 推送Token管理 |

## 3. 数据模型

- **User**: 用户基本信息
- **UserDevice**: 设备登录记录
- **UserSession**: 用户会话信息
- **UserProfile**: 用户详细资料
- **UserSettings**: 用户个人设置
- **UserQRCode**: 用户二维码记录
- **UserPushToken**: 推送Token

## 4. 推送通知

- `notification.user.force_logout.{user_id}` - 多端互踢通知
- `notification.user.unusual_login.{user_id}` - 异常登录提醒
- `notification.user.password_changed.{user_id}` - 密码修改通知
- `notification.user.profile_updated.{user_id}` - 用户资料更新通知
- `notification.user.friend_profile_changed.{user_id}` - 好友资料变更通知
- `notification.user.status_changed.{user_id}` - 在线状态变更通知

## 5. 依赖服务

- **Redis**: Token缓存、在线状态
- **PostgreSQL**: 设备登录记录
- **NATS**: 强制下线消息推送
- **File Service**: 头像上传
- **Redis**: 在线状态、二维码缓存
- **PostgreSQL**: 用户资料持久化

---

返回: [后端总体设计](../backend-design.md)
