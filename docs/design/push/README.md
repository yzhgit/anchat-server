# Push Service (推送服务)

## 1. 服务概述

**职责**: 离线推送、通知管理

**核心功能**:
- 离线推送（iOS APNs、Android FCM/极光/华为/小米）
- 推送类型（新消息、好友申请、@消息、音视频通话）
- 推送策略（免打扰时段、已读不推、折叠）
- 推送内容（标题、角标、声音）
- 推送统计

## 2. 文档导航

| 功能 | 文档 | 说明 |
|------|------|------|
| 推送服务 | [push.md](push.md) | 离线推送设计 |
| 推送接口 | [push-notification.md](../push-notification.md) | 完整接口定义 |

## 3. 数据模型

- **PushLog**: 推送日志
- **PushFailure**: 推送失败记录

## 4. 推送通知

- `notification.push.delivery_status.{user_id}` - 离线推送发送结果通知
- `notification.push.token_invalid.{user_id}` - 推送Token失效通知

## 5. 依赖服务

- **User Service**: 推送Token
- **Message Service**: 消息内容
- **Conversation Service**: 免打扰设置
- **APNs/FCM/极光/华为/小米**: 推送通道
- **Redis**: 推送队列
- **NATS**: 推送事件订阅

---

返回: [后端总体设计](../backend-design.md)
