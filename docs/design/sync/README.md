# Sync Service (数据同步服务)

> ⚠️ **已废弃**: Sync Service 已暂时移除。同步功能将分散到各对应服务中（好友同步→Friend、群组同步→Group、消息同步→Message 等）。以下文档保留仅供历史参考。
> 最后更新日期: 2026-08-19

## 1. 服务概述

**职责**: 跨端数据同步、增量同步

**核心功能**:
- 好友数据同步
- 群组数据同步
- 会话数据同步
- 消息同步
- 设置同步

## 2. 文档导航

| 功能 | 文档 | 说明 |
|------|------|------|
| 数据同步 | [sync.md](sync.md) | 增量同步与全量同步 |

## 3. 同步范围

| 数据类型 | 同步内容 |
|----------|----------|
| 会话 | 会话列表、置顶、免打扰 |
| 消息 | 消息历史、未读状态 |
| 好友 | 好友列表、黑名单 |
| 群组 | 群组列表、成员 |
| 用户 | 用户资料 |

## 4. 推送通知

- `notification.sync.request.{user_id}` - 数据同步请求通知
- `notification.sync.completed.{user_id}` - 数据同步完成通知

## 5. 依赖服务

- **Friend Service**
- **Group Service**
- **Conversation Service**
- **Message Service**
- **Redis**: 同步状态缓存
- **PostgreSQL**: 数据版本

---

返回: [后端总体设计](../backend-design.md)
