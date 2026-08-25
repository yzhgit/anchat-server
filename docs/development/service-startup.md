# AnyChat 服务启动指南

## 启动前准备

确保基础设施已启动并健康：

```bash
mage docker:up
mage docker:ps
mage db:up
```

## 启动方式

### 方案 1: 使用独立终端（推荐用于开发调试）

根据需要，在多个终端窗口中分别启动服务。注意：**本地开发时每个服务需要不同的端口**，通过环境变量指定：

**终端 1 - 启动 gateway 服务:**
```bash
cd /home/mosee/flamingo/flamingo-server-kratos
HTTP_ADDR=0.0.0.0:8080 mage dev:gateway
```

**终端 2 - 启动 user service:**
```bash
cd /home/mosee/flamingo/flamingo-server-kratos
GRPC_ADDR=0.0.0.0:50051 mage dev:user
```

**终端 3 - 启动 friend service:**
```bash
cd /home/mosee/flamingo/flamingo-server-kratos
GRPC_ADDR=0.0.0.0:50052 mage dev:friend
```

**终端 4 - 启动 realtime service:**
```bash
cd /home/mosee/flamingo/flamingo-server-kratos
HTTP_ADDR=0.0.0.0:8081 mage dev:realtime
```

完整列表：

| 命令 | 服务 | 端口 |
|------|------|------|
| `GRPC_ADDR=0.0.0.0:50051 mage dev:user` | 用户与认证服务 | gRPC 50051 |
| `GRPC_ADDR=0.0.0.0:50052 mage dev:friend` | 好友服务 | gRPC 50052 |
| `GRPC_ADDR=0.0.0.0:50053 mage dev:group` | 群组服务 | gRPC 50053 |
| `GRPC_ADDR=0.0.0.0:50054 mage dev:message` | 消息服务 | gRPC 50054 |
| `GRPC_ADDR=0.0.0.0:50055 mage dev:conversation` | 会话服务 | gRPC 50055 |
| `GRPC_ADDR=0.0.0.0:50056 mage dev:file` | 文件服务 | gRPC 50056 |
| `GRPC_ADDR=0.0.0.0:50057 mage dev:rtc` | 音视频通话服务 | gRPC 50057 |
| `GRPC_ADDR=0.0.0.0:50058 mage dev:push` | 离线推送服务 | gRPC 50058 |
| `HTTP_ADDR=0.0.0.0:8081 mage dev:realtime` | WebSocket 长连接服务 | HTTP 8081 |

### 方案 2: 使用 tmux/screen（推荐用于管理多个服务）

```bash
# 安装 tmux
sudo apt-get install tmux  # Ubuntu/Debian
brew install tmux          # macOS

# 启动 tmux 会话
tmux new -s anychat

# 窗口 0: user
GRPC_ADDR=0.0.0.0:50051 mage dev:user

# 窗口 1: friend (Ctrl+B, C 新建窗口)
GRPC_ADDR=0.0.0.0:50052 mage dev:friend

# 窗口 2: realtime (Ctrl+B, C 新建窗口)
HTTP_ADDR=0.0.0.0:8081 mage dev:realtime

# Detach: Ctrl+B, D
# Re-attach: tmux attach -t anychat
```

## 端口验证

启动前检查端口：
```bash
# 检查所有端口占用情况
lsof -i :8080   # gateway HTTP
lsof -i :50051  # user gRPC
lsof -i :50052  # friend gRPC
lsof -i :8081   # realtime WebSocket
```

## 服务验证

所有服务启动后，验证：

```bash
# 检查 Gateway
curl http://localhost:8080/health

# 检查 gRPC 服务
grpcurl -plaintext localhost:50051 list       # user
grpcurl -plaintext localhost:50052 list       # friend
grpcurl -plaintext localhost:8081/ws          # realtime WebSocket

# 检查 Consul 注册的服务
curl http://localhost:8500/v1/catalog/services
```

## 常见问题

### Q1: 端口被占用
**症状**: `bind: address already in use`

**解决**:
```bash
# 查找占用端口的进程
sudo lsof -i :50051
# 杀掉进程
sudo kill -9 <PID>
```

### Q2: 服务无法连接
**症状**: `Failed to connect to backend services`

**解决**:
1. 确保基础设施已启动：`mage docker:up`
2. 确保数据库已迁移：`mage db:up`
3. 检查目标服务是否已启动并监听对应端口
4. 确认 Consul 服务发现已注册

### Q3: 数据库连接失败
**症状**: `Failed to connect database`

**解决**:
```bash
# 确保 Docker 基础设施运行
mage docker:up
mage docker:ps

# 运行数据库迁移
mage db:up
```

### Q4: Consul 未启动导致服务发现失败
**症状**: `service discovery error` / `connection refused`

**解决**:
```bash
# 检查 Consul 容器状态
docker ps | grep consul

# 检查 Consul 端口
curl http://localhost:8500/v1/status/leader
```

## 端口分配

| 服务 | 本地开发端口 | 容器内端口 | 说明 |
|------|-------------|-----------|------|
| gateway | HTTP 8080 | HTTP 8080 | HTTP API 网关 |
| user | gRPC 50051 | gRPC 50051 | 用户与认证服务 |
| friend | gRPC 50052 | gRPC 50051 | 好友服务 |
| group | gRPC 50053 | gRPC 50051 | 群组服务 |
| message | gRPC 50054 | gRPC 50051 | 消息服务 |
| conversation | gRPC 50055 | gRPC 50051 | 会话服务 |
| file | gRPC 50056 | gRPC 50051 | 文件服务 |
| rtc | gRPC 50057 | gRPC 50051 | 音视频通话服务 |
| push | gRPC 50058 | gRPC 50051 | 离线推送服务 |
| realtime | HTTP 8081 | HTTP 8081 | WebSocket 长连接服务 |

> 容器内所有业务微服务统一使用 gRPC 50051，本地开发通过环境变量 `GRPC_ADDR` 覆盖为 50051~50058 不同端口。
>
> 详细端口分配: `docs/development/port-allocation.md`