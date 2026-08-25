# Friend Service (好友服务)

## 1. 服务概述

**职责**: 好友关系管理、好友申请、黑名单

**核心功能**:
- 好友搜索与添加
- 好友管理（列表、备注、分组）
- 好友申请（发送、处理）
- 黑名单管理
- 通讯录同步
- 好友状态同步

## 2. 文档导航

| 功能 | 文档 | 说明 |
|------|------|------|
| 好友关系 | [friendship.md](friendship.md) | 好友列表与关系管理 |
| 好友申请 | [request.md](request.md) | 好友请求发送与处理 |
| 黑名单 | [blacklist.md](blacklist.md) | 黑名单管理 |

## 3. 数据模型

- **Friendship**: 好友关系表（双向）
- **FriendRequest**: 好友申请记录
- **Blacklist**: 黑名单
- **FriendSettings**: 好友相关设置

## 4. 推送通知

- `notification.friend.request.{to_user_id}` - 好友请求通知
- `notification.friend.request_handled.{from_user_id}` - 好友请求处理结果通知
- `notification.friend.deleted.{user_id}` - 好友删除通知
- `notification.friend.remark_updated.{user_id}` - 好友备注修改同步
- `notification.friend.blacklist_changed.{user_id}` - 黑名单变更通知

## 5. 依赖服务

- **User Service**: 用户信息查询
- **Message Service**: 好友申请消息
- **NATS**: 好友变更事件推送
- **Redis**: 好友关系缓存
- **PostgreSQL**: 好友关系持久化

---

返回: [后端总体设计](../backend-design.md)
