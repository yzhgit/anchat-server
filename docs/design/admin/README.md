# Admin Service (管理服务)

> ⚠️ **已废弃**: Admin Service 已暂时移除。以下文档保留仅供历史参考。相关管理功能将在后续版本中评估是否添加。
> 最后更新日期: 2026-08-19

## 1. 服务概述

**职责**: 系统管理、用户管理、数据统计

**核心功能**:
- 管理员管理
- 用户管理（搜索、禁用、封禁）
- 群组管理（查看、解散）
- 消息管理（敏感监控、撤回）
- 系统配置
- 数据统计
- 日志审计

## 2. 文档导航

| 功能 | 文档 | 说明 |
|------|------|------|
| 管理后台 | [admin.md](admin.md) | 完整管理功能设计 |

## 3. 数据模型

- **AdminUser**: 管理员用户
- **AdminRole**: 管理员角色
- **AdminPermission**: 权限
- **AuditLog**: 审计日志
- **SystemConfig**: 系统配置

## 4. 推送通知

- `notification.admin.announcement.broadcast` - 系统公告通知
- `notification.admin.user_banned.{user_id}` - 用户封禁通知
- `notification.admin.maintenance.broadcast` - 系统维护通知

## 5. 依赖服务

- **所有其他服务** (通过gRPC调用)
- **PostgreSQL**: 配置存储
- **Redis**: 配置缓存

---

返回: [后端总体设计](../backend-design.md)
