# Group Service (群组服务)

## 1. 服务概述

**职责**: 群组创建、成员管理、群设置

**核心功能**:
- 群组创建与解散
- 成员管理（邀请、申请、移除、退出）
- 群主与管理员管理
- 群组设置（名称、头像、公告、群简介、权限）
- 群成员设置（群昵称、免打扰、置顶聊天）
- 群置顶消息
- 群组信息同步

## 2. 文档导航

| 功能 | 文档 | 说明 |
|------|------|------|
| 群组管理 | [group.md](group.md) | 群组创建、解散、信息、设置 |
| 群成员管理 | [member.md](member.md) | 成员邀请、移除、角色管理、禁言 |
| 更新群设置 | [settings.md](settings.md) | 入群验证、邀请权限、查看历史、加好友、成员修改权 |
| 群聊消息置顶 | [pinned-message.md](pinned-message.md) | 顶部置顶栏、权限控制、漫游同步 |
| 群备注 | [group-remark.md](group-remark.md) | 用户为群设置仅自己可见的别名 |
| 群二维码 | [group-qrcode.md](group-qrcode.md) | 生成/刷新/扫码加入群的二维码机制 |

## 3. 数据模型

- **Group**: 群组基本信息 (含群设置)
- **GroupMember**: 群成员关系 (含个人置顶、禁言)
- **GroupJoinRequest**: 入群申请
- **GroupPinnedMessage**: 群置顶消息

## 4. 推送通知

> 详细通知主题见各子文档：
> - 群组通用: group.md
> - 成员相关: member.md

- `notification.group.invited.{user_id}` - 群组邀请通知
- `notification.group.disbanded.{group_id}` - 群组解散通知

## 5. 依赖服务

- **User Service**: 用户信息
- **Message Service**: 群系统消息
- **File Service**: 群头像
- **NATS**: 群变更事件
- **Redis**: 群信息缓存、成员缓存
- **PostgreSQL**: 群数据持久化

---

返回: [后端总体设计](../backend-design.md)
