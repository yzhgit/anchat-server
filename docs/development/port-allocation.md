# 端口分配规划

## 设计原则

本项目的端口规划遵循以下原则：

1. **容器内端口统一**：生产环境（Kubernetes）中每个 Pod 有独立的网络命名空间，所有业务微服务在容器内统一使用 gRPC `50051`。
2. **通过环境变量覆盖**：本地开发（`mage dev:*`）时，多个服务运行在同一台宿主机上，通过环境变量指定不同端口避免冲突（50051 起顺序递增）。
3. **中间件使用惯用端口**：PostgreSQL、Redis、NATS 等中间件使用各自惯用端口，无需映射。
4. **服务发现基于服务名**：内部通信通过 Consul 服务发现，不依赖硬编码端口。

## 端口策略概览

每个微服务暴露两类端口：
- **gRPC 端口**（`50051~`）：服务间内部通信，通过 Consul 发现
- **HTTP 端口**（`9101~`）：暴露 `/metrics` 供 Prometheus 采集；Gateway 和 Realtime 的 HTTP 端口同时承载业务与指标

| 环境 | 部署方式 | gRPC 端口 | HTTP 端口 | 说明 |
|------|---------|-----------|-----------|------|
| **生产 (Kubernetes)** | K8s Pod | 容器内 `50051`（统一） | 容器内 `9101`（业务） / Gateway `8080`、Realtime `8081` | 每 Pod 独立网络命名空间，端口不冲突 |
| **本地开发 (mage dev:*)** | 进程 | 50051~50058（顺序递增） | 9101~9108（业务） / Gateway `8080`、Realtime `8081` | 各服务不同端口，避免宿主机冲突 |

### 端口分配总表

| 类别 | 服务/组件 | gRPC 端口 | HTTP 端口 | 协议 | 说明 |
|------|----------|-----------|-----------|------|------|
| **外部网关** | gateway | — | 8080 | HTTP | 对外 HTTP API 统一入口，`/metrics` 同端口 |
| **微服务** | user | 50051 | 9101 | gRPC / HTTP | 用户与认证服务，HTTP 端口供 Prometheus 采集 |
| **微服务** | friend | 50052 | 9102 | gRPC / HTTP | 好友服务 |
| **微服务** | group | 50053 | 9103 | gRPC / HTTP | 群组服务 |
| **微服务** | message | 50054 | 9104 | gRPC / HTTP | 消息服务 |
| **微服务** | conversation | 50055 | 9105 | gRPC / HTTP | 会话服务 |
| **微服务** | file | 50056 | 9106 | gRPC / HTTP | 文件服务 |
| **微服务** | rtc | 50057 | 9107 | gRPC / HTTP | 音视频通话服务 |
| **微服务** | push | 50058 | 9108 | gRPC / HTTP | 离线推送服务 |
| **微服务** | realtime | — | 8081 | HTTP | WebSocket 长连接服务，`/metrics` 同端口 |
| **中间件** | PostgreSQL | 5432 | TCP | 主数据库 |
| **中间件** | Redis | 6379 | TCP | 缓存与会话存储 |
| **中间件** | NATS | 4222 | TCP | 消息队列客户端端口 |
| **中间件** | NATS Monitoring | 8222 | HTTP | NATS 监控接口 |
| **中间件** | MinIO API | 9000 | HTTP | 对象存储 API |
| **中间件** | MinIO Console | 9091 | HTTP | MinIO 管理控制台 |
| **中间件** | Consul | 8500 | HTTP | 服务发现与注册 |
| **中间件** | Prometheus | 9090 | HTTP | 指标采集和存储 |
| **中间件** | Grafana | 3000 | HTTP | 可视化监控面板 |
| **中间件** | Tempo UI | 3200 | HTTP | 分布式追踪 UI |
| **中间件** | Tempo OTLP gRPC | 4317 | gRPC | OTLP gRPC 接入端点（服务端 → Tempo） |
| **中间件** | Tempo OTLP HTTP | 4318 | HTTP | OTLP HTTP 接入端点（服务端 → Tempo） |
| **中间件** | Loki | 3100 | HTTP | 日志存储 |
| **中间件** | LiveKit | 7880 | WebSocket | LiveKit 音视频服务器 |
| **中间件** | LiveKit RTC TCP | 7881 | TCP | LiveKit 音视频 TCP |
| **中间件** | LiveKit RTC UDP | 7882 | UDP | LiveKit 音视频 UDP |

> 所有端口无冲突：微服务 gRPC 50051~50058 + HTTP 9101~9108 + Gateway 8080 + Realtime 8081 + 中间件各自惯用端口。

---

## 本地开发启动命令

本地开发时每个服务需要不同的端口，通过环境变量覆盖默认值（`50051`）：

```bash
# 启动基础设施
mage docker:up
mage docker:ps
mage db:up

# 启动微服务（在不同终端窗口）
# 每个服务需设置 gRPC 端口和 HTTP 端口
GRPC_ADDR=0.0.0.0:50051 HTTP_ADDR=0.0.0.0:9101 mage dev:user          # 用户与认证服务
GRPC_ADDR=0.0.0.0:50052 HTTP_ADDR=0.0.0.0:9102 mage dev:friend        # 好友服务
GRPC_ADDR=0.0.0.0:50053 HTTP_ADDR=0.0.0.0:9103 mage dev:group         # 群组服务
GRPC_ADDR=0.0.0.0:50054 HTTP_ADDR=0.0.0.0:9104 mage dev:message       # 消息服务
GRPC_ADDR=0.0.0.0:50055 HTTP_ADDR=0.0.0.0:9105 mage dev:conversation  # 会话服务
GRPC_ADDR=0.0.0.0:50056 HTTP_ADDR=0.0.0.0:9106 mage dev:file          # 文件服务
GRPC_ADDR=0.0.0.0:50057 HTTP_ADDR=0.0.0.0:9107 mage dev:rtc       # 音视频通话服务
GRPC_ADDR=0.0.0.0:50058 HTTP_ADDR=0.0.0.0:9108 mage dev:push          # 离线推送服务
HTTP_ADDR=0.0.0.0:8081 mage dev:realtime      # WebSocket 长连接服务

# 启动 gateway 服务（HTTP API 网关）
HTTP_ADDR=0.0.0.0:8080 mage dev:gateway
```

---

## 配置文件

每个服务的配置文件位于 `app/<service>/configs/config.yaml`：

**业务微服务（user, friend, group, message, conversation, file, rtc, push）**：

```yaml
server:
  http:
    addr: ${HTTP_ADDR:0.0.0.0:9101}  # HTTP 端口（/metrics 供 Prometheus 采集）
    timeout: 5s
  grpc:
    addr: ${GRPC_ADDR:0.0.0.0:50051}  # gRPC 端口（服务间通信）
    timeout: 5s
```

**Realtime 服务**：

```yaml
server:
  http:
    addr: ${HTTP_ADDR:0.0.0.0:8081}  # HTTP 端口（WebSocket + /metrics）
    timeout: 5s
```

环境变量通过 Kratos `env.NewSource("")` 自动解析，格式为 `${VAR:default}`。

---

## Kubernetes 部署说明

在 Kubernetes 中，每个业务微服务运行在独立的 Pod 中，每个 Pod 有独立的网络命名空间，因此所有服务 gRPC 统一 `50051`、HTTP 统一 `9101`：

```yaml
# 示例：User Service Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  template:
    spec:
      containers:
        - name: user
          ports:
            - containerPort: 50051   # gRPC（服务间通信）
            - containerPort: 9101     # HTTP（/metrics 供 Prometheus 采集）
          env:
            - name: GRPC_ADDR
              value: "0.0.0.0:50051"
            - name: HTTP_ADDR
              value: "0.0.0.0:9101"

---
apiVersion: v1
kind: Service
metadata:
  name: user-svc
spec:
  selector:
    app: user-service
  ports:
    - name: grpc
      port: 50051
      targetPort: 50051
    - name: http
      port: 9101
      targetPort: 9101
```

K8s Service 暴露 `50051`（gRPC）和 `9101`（HTTP），其他 Pod 通过 `user-svc:50051` 或 Consul 服务名 `user` 来访问，Prometheus 通过 `user-svc:9101/metrics` 采集指标。

**Realtime Service**（HTTP/WebSocket，独立 Pod）：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: realtime-service
spec:
  template:
    spec:
      containers:
        - name: realtime
          ports:
            - containerPort: 8081    # HTTP（WebSocket + /metrics）
          env:
            - name: HTTP_ADDR
              value: "0.0.0.0:8081"
```

---

## 端口检查命令

```bash
# 检查微服务端口占用
lsof -i :50051  # user gRPC
lsof -i :50058  # push gRPC
lsof -i :9101   # user HTTP (metrics)
lsof -i :9108   # push HTTP (metrics)
lsof -i :8081   # realtime WebSocket / HTTP (metrics)
lsof -i :8080   # Gateway HTTP (metrics)

# 查看基础设施状态
mage docker:ps

# 检查 Consul
curl http://localhost:8500/v1/status/leader
curl http://localhost:8500/v1/catalog/services
```

## 故障排查

### 端口已被占用

```bash
lsof -i :50051
kill -9 <PID>
```

### 服务无法启动

1. 检查端口是否被占用
2. 检查 `app/<service>/configs/config.yaml` 中的端口设置
3. 确认环境变量 `GRPC_ADDR` / `HTTP_ADDR` 是否正确设置
4. 检查 Docker 容器状态 `mage docker:ps`

### Consul 服务发现失败

```bash
# 检查 Consul 是否运行
curl http://localhost:8500/v1/status/leader

# 检查已注册的服务
curl http://localhost:8500/v1/catalog/services
```

---

## 更新记录

| 日期 | 变更 | 说明 |
|------|------|------|
| 2026-08-19 | 重新规划 | 微服务容器内统一 gRPC 50051，本地开发从 50051 顺序递增；Realtime HTTP 8081；中间件使用惯用端口 |
| 2026-08-23 | 新增 HTTP/Metrics 端口 | 业务微服务新增 HTTP 端口 9101~9108 暴露 /metrics；Gateway 8080、Realtime 8081 同端口兼作指标端口 |