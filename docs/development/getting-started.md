# 快速开始指南

## 环境准备

### 1. 安装Go环境

确保安装了Go 1.26或更高版本：

```bash
go version
```

### 2. 安装Docker和Docker Compose

```bash
docker --version
docker-compose --version
```

### 3. 安装Mage构建工具

Mage是基于Go的构建工具，类似于Make但使用Go代码定义任务。

```bash
go install github.com/magefile/mage@latest
mage -version
```

## 本地开发

### 1. 克隆项目

```bash
git clone https://github.com/yzhgit/anychat-server
cd anychat-server
```

### 2. 安装依赖

```bash
mage deps
mage install
```

### 3. 查看所有可用命令

```bash
mage -l
```

输出示例：
```
Targets:
  build:all          builds all services
  build:user         builds user service
  build:friend       builds friend service
  build:realtime    builds realtime service
  ...
  clean              removes build artifacts
  db:up              runs database migrations up
  db:down            runs database migrations down
  deps               installs dependencies
  dev:user           runs user service locally
  dev:friend         runs friend service locally
  dev:realtime      runs realtime service locally
  dev:wire           generates Wire dependency injection code
  docker:up          starts docker compose
  docker:down        stops docker compose
  docker:logs        shows docker compose logs
  docker:ps          shows docker compose status
  fmt                formats code
  install            installs required tools
  lint               runs linter
  proto              generates protobuf code
  test:all           runs all tests
  test:coverage      generates test coverage report
  docs:generate      generates API documentation
  docs:serve         starts documentation server
  ...
```

### 4. 启动基础设施

```bash
mage docker:up
mage docker:ps
```

启动后可以访问以下服务：

- **PostgreSQL**: localhost:5432（用户名: anychat, 密码: anychat, 数据库: im）
- **Redis**: localhost:6379
- **NATS**: localhost:4222 (客户端), localhost:8222 (管理)
- **MinIO**: API: localhost:9000, Console: http://localhost:9091（用户名: admin, 密码: admin）
- **Consul**: http://localhost:8500 (服务发现)
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000（用户名: admin, 密码: admin）
- **Tempo**: http://localhost:3200
- **Loki**: http://localhost:3100
- **LiveKit**: ws://localhost:7880

### 5. 数据库迁移

```bash
# 运行迁移
mage db:up

# 回滚迁移
mage db:down
```

### 6. 运行服务

```bash
# 运行用户与认证服务
mage dev:user

# 运行好友服务
mage dev:friend

# 运行WebSocket长连接服务
mage dev:realtime

# 生成Wire依赖注入代码（首次或修改后需要）
mage dev:wire
```

## 开发流程

### 构建服务

```bash
# 构建所有服务
mage build:all

# 构建特定服务
mage build:user
mage build:friend
mage build:realtime
```

构建的二进制文件将输出到 `bin/` 目录。

### 添加新功能

1. 在对应的 `app/<service>/internal/service/` 层添加业务逻辑
2. 在 `app/<service>/internal/handler/` 层添加 gRPC handler
3. 在 `api/proto/<service>/v1/` proto 文件中定义接口
4. 生成 protobuf 代码: `mage proto`
5. 生成 Wire 代码: `mage dev:wire`
6. 运行测试: `mage test:all`

### 代码规范

```bash
mage fmt
mage lint
mage test:all
mage test:coverage
```

### 提交代码

```bash
git commit -m "feat(user): 添加用户注册功能"
```

## Mage常用命令速查

### 开发相关
```bash
mage deps         # 安装依赖
mage install      # 安装开发工具
mage dev:user     # 运行用户服务
mage dev:realtime # 运行realtime服务
mage dev:wire     # 生成Wire代码
mage fmt          # 格式化代码
mage lint         # 代码检查
```

### 构建相关
```bash
mage build:all    # 构建所有服务
mage build:user   # 构建用户服务
mage clean        # 清理构建产物
mage proto        # 生成protobuf代码
```

### 测试相关
```bash
mage test:all           # 运行所有测试
mage test:coverage      # 生成覆盖率报告
```

### Docker相关
```bash
mage docker:up    # 启动所有容器
mage docker:down  # 停止所有容器
mage docker:logs  # 查看日志
mage docker:ps    # 查看容器状态
mage docker:build # 构建Docker镜像
```

### 数据库相关
```bash
mage db:up   # 运行迁移
mage db:down # 回滚迁移
```

### 文档相关
```bash
mage docs:generate  # 生成API文档
mage docs:serve     # 启动文档服务器
mage docs:build     # 构建文档站点
mage docs:validate  # 校验API文档
```

## 常见问题

### Q: Mage命令找不到？

确保已安装Mage且 `$GOPATH/bin` 在 PATH 中：
```bash
go install github.com/magefile/mage@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### Q: 端口已被占用怎么办？

```bash
# 查找占用端口的进程
sudo lsof -i :<port>
# 停止进程
sudo kill -9 <PID>
```

### Q: 数据库连接失败？

```bash
mage docker:ps
mage docker:logs
```

### Q: Consul服务发现失败？

```bash
curl http://localhost:8500/v1/status/leader
```

## 下一步

- 查看[设计文档](/design/backend-design.md)了解系统架构
- 查看[API文档](/api/QUICKSTART.md)了解接口定义
- 阅读 `magefile.go` 了解更多构建任务
- 开始实现你的第一个功能！