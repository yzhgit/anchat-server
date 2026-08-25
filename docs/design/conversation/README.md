# Conversation Service (会话服务)

## 1. 服务概述

**职责**: 会话列表管理、会话状态、未读数管理

**核心功能**:
- 会话管理（创建、获取、删除）
- 会话状态（置顶、免打扰）
- 未读数管理
- 会话类型（单聊、群聊、系统消息）
- 会话同步

## 2. 文档导航

| 功能 | 文档 | 说明 |
|------|------|------|
| 会话管理 | [conversation.md](conversation.md) | 会话列表、置顶、免打扰 |

## 3. 数据模型

- **Conversation**: 会话表
- **SessionSettings**: 会话设置
- **SessionUnread**: 未读数记录

## 4. 推送通知

- `notification.conversation.unread_updated.{user_id}` - 会话未读数更新通知
- `notification.conversation.pin_updated.{user_id}` - 会话置顶状态同步
- `notification.conversation.deleted.{user_id}` - 会话删除同步
- `notification.conversation.mute_updated.{user_id}` - 会话免打扰设置同步

## 5. 依赖服务

- **Message Service**: 最后消息
- **User Service**: 会话对方信息
- **Group Service**: 群信息
- **Redis**: 会话缓存、未读数缓存
- **PostgreSQL**: 会话持久化

---

返回: [后端总体设计](../backend-design.md)
