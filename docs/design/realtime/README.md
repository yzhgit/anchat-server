# Realtime Service (WebSocket 长连接服务)

## 1. 服务概述

**职责**: WebSocket 长连接管理、消息实时推送

> **注意**: 本目录名称从 `gateway` 重命名为 `realtime`，实际服务名为 `realtime`，代码位于 `app/realtime/`。

**核心功能**:
- 连接管理（建立、认证、保活、断线重连）
- 消息推送（实时推送、送达确认、重传）
- 在线状态（多端在线检测）
- 连接分布（多节点部署、负载均衡）
- 协议处理（接入、退出、强制下线、心跳）

## 2. 文档导航

| 功能 | 文档 | 说明 |
|------|------|------|
| WebSocket | [websocket.md](websocket.md) | 连接管理与消息推送 |

## 3. 推送通知架构

Realtime 作为推送通知的核心枢纽：
- **NATS 订阅**: 为每个用户订阅 `notification.*.*.{user_id}`
- **WebSocket 推送**: 将 NATS 消息转发到客户端
- **推送失败处理**: 用户不在线时触发 Push Service

## 4. 依赖服务

- **User Service**: Token 验证（原 Auth 服务已合并到 User 服务）
- **Message Service**: 消息处理
- **Redis**: 连接信息、在线状态
- **NATS**: 跨节点消息路由
- **Consul**: 服务发现

## 5. 通信方式

```
Client (WebSocket) ←→ Realtime (8081) ←gRPC→ Backend Microservices
                                           ↕
                                      NATS JetStream
```

---

返回: [后端总体设计](../backend-design.md)