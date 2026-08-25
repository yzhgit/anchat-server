# API 测试文档

本目录包含 AnyChat 项目的 API 文档和测试脚本。

## 目录结构

```
docs/api/
└── README.md           # 本文件

tests/api/
├── test-all.sh         # 运行所有 API 测试（推荐入口）
├── common.sh           # 共享函数库
├── README.md           # 详细测试文档
├── user/
│   └── test-user-api.sh    # User Service API 测试（含认证功能）
└── friend/
    └── test-friend-api.sh  # Friend Service API 测试
```

> **注意**: 旧的测试脚本路径已更改：
> - `scripts/test-api.sh` → `tests/api/user/test-user-api.sh`
> - `scripts/test-friend-api.sh` → `tests/api/friend/test-friend-api.sh`
> - `scripts/debug-friend-api.sh` → 已删除
> - `tests/integration/*_service_test.go` → 已删除（统一使用 Shell API 测试）
> - `tests/e2e/test-e2e.sh` → 已删除（功能合并到模块化 API 测试）
>
> **Auth 与 User 合并**: 原独立的 Auth Service 已合并到 User Service 中，认证相关测试统一在 `tests/api/user/` 下进行。


## 测试脚本使用

### 前置条件

1. **安装必要工具**

```bash
# 安装 jq（JSON 处理工具）
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

2. **安装 grpcurl（用于 gRPC 测试）**

```bash
# 使用 Go 安装
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 或下载二进制文件
# https://github.com/fullstorydev/grpcurl/releases
```

3. **启动服务**

```bash
# 启动基础设施
mage docker:up

# 运行数据库迁移
mage db:up

# 启动服务（在不同终端窗口，需通过环境变量指定端口）
GRPC_ADDR=0.0.0.0:50051 mage dev:user      # 启动 user service（含认证功能）
GRPC_ADDR=0.0.0.0:50052 mage dev:friend     # 启动 friend service
HTTP_ADDR=0.0.0.0:8081 mage dev:realtime  # 启动 realtime service
```

### HTTP API 测试

**快速开始：**

```bash
# 运行所有 API 测试（推荐）
./tests/api/test-all.sh

# 运行单个模块测试
./tests/api/user/test-user-api.sh
./tests/api/friend/test-friend-api.sh
```

**自定义 Gateway 地址：**

```bash
# 测试远程服务器
GATEWAY_URL=http://your-server:8080 ./tests/api/test-all.sh
```

> Gateway 地址指向 gateway 服务（端口 8080），它负责将 HTTP 请求转码转发到后端微服务的 gRPC 端口。

**测试内容：**

**User Service API（含认证功能）** (`tests/api/user/test-user-api.sh`):
- ✓ 发送短信/邮箱验证码
- ✓ 目标格式校验与发送频率限制
- ✓ 错误验证码拒绝、固定验证码注册
- ✓ 重置密码场景验证码发送
- ✓ 用户注册
- ✓ 用户登录
- ✓ Token 刷新
- ✓ 修改密码
- ✓ 用户登出
- ✓ 获取个人资料
- ✓ 更新个人资料

> 验证码与认证流程统一在 User 测试脚本中，不再维护单独的 auth 测试脚本。
- ✓ 搜索用户
- ✓ 获取/更新用户设置
- ✓ 刷新二维码
- ✓ 更新推送Token

**Friend Service API** (`tests/api/friend/test-friend-api.sh`):
- ✓ 发送好友申请
- ✓ 获取好友申请列表
- ✓ 接受/拒绝好友申请
- ✓ 获取好友列表
- ✓ 更新好友备注
- ✓ 黑名单管理
- ✓ 删除好友

## 手动测试示例

### 使用 cURL 测试 HTTP API

```bash
# 1. 发送注册验证码
curl -X POST http://localhost:8080/api/v1/users/send-code \
  -H "Content-Type: application/json" \
  -d '{
    "target": "13800138000",
    "target_type": 1,
    "purpose": 1,
    "device_id": "device-001"
  }'

# 2. 用户注册（开发环境默认固定验证码为 123456，可通过 VERIFY_DEBUG_FIXED_CODE 覆盖）
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "13800138000",
    "password": "Test@123456",
    "verify_code": "123456",
    "nickname": "测试用户",
    "device_type": 1,
    "device_id": "device-001"
  }'

# 3. 用户登录
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "account": "13800138000",
    "password": "Test@123456",
    "device_type": 1,
    "device_id": "device-001"
  }'

# 4. 获取个人资料（需要替换 YOUR_ACCESS_TOKEN）
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

枚举值说明（HTTP 入参）：

- `device_type`: `1=ios` `2=android` `3=web` `4=pc` `5=h5`
- `target_type`: `1=sms` `2=email`
- `purpose`: `1=register` `2=login` `3=reset_password` `4=bind_phone` `5=change_phone` `6=bind_email` `7=change_email`
- `source`（好友申请来源）: `1=search` `2=qrcode` `3=group` `4=contacts`

### 邮箱验证码联调（SMTP）

如果要让 `/api/v1/users/send-code` 真实发送邮箱验证码，请先为 **user service** 配置 SMTP：

```bash
export EMAIL_HOST=smtp.qq.com
export EMAIL_PORT=465
export EMAIL_USERNAME=your_account@qq.com
export EMAIL_PASSWORD=your_smtp_auth_code
export EMAIL_FROM_NAME=AnyChat
export EMAIL_FROM_ADDRESS=your_account@qq.com
```

说明：

- `EMAIL_PORT=465` 表示 SMTP over SSL
- `EMAIL_PORT=587` 表示 SMTP + STARTTLS
- `EMAIL_PASSWORD` 应填写邮箱服务商提供的 SMTP 授权码，而不是登录密码
- `EMAIL_FROM_ADDRESS` 最好与 `EMAIL_USERNAME` 保持一致

然后重启 **user service**，再调用邮箱发码接口：

```bash
curl -X POST http://localhost:8080/api/v1/users/send-code \
  -H "Content-Type: application/json" \
  -d '{
    "target": "you@example.com",
    "target_type": 2,
    "purpose": 1,
    "device_id": "web-mail-test"
  }'
```

> 如果 `EMAIL_HOST` 未配置或仍为默认占位值，开发环境下接口仍会返回成功，但不会真正发送邮件。

### 使用 grpcurl 测试 gRPC API

> 注意：gRPC 中 `device_type` / `target_type` / `purpose` 使用 proto 枚举名（例如 `DEVICE_TYPE_IOS`）。

```bash
# 1. 列出所有服务
grpcurl -plaintext localhost:50051 list

# 2. 查看服务方法
grpcurl -plaintext localhost:50051 list user.v1.UserService

# 3. 查看方法详情
grpcurl -plaintext localhost:50051 describe user.v1.UserService.Login

# 4. 调用登录接口
grpcurl -plaintext -d '{
  "account": "13800138000",
  "password": "Test@123456",
  "device_type": "DEVICE_TYPE_IOS",
  "device_id": "device-001"
}' localhost:50051 user.v1.UserService/Login

# 5. 调用用户资料接口
grpcurl -plaintext -d '{
  "user_id": "user-id-from-login"
}' localhost:50051 user.v1.UserService/GetProfile
```

## 常见问题

### 1. 测试脚本权限错误

```bash
# 解决方法：添加执行权限
chmod +x tests/api/test-all.sh
chmod +x tests/api/user/test-user-api.sh
chmod +x tests/api/friend/test-friend-api.sh
```

### 2. jq 命令未找到

```bash
# 安装 jq
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq
```

### 3. 连接被拒绝

```bash
# 检查服务是否启动
mage docker:ps

# 检查端口是否被占用
lsof -i :8080   # Gateway HTTP
lsof -i :50051  # User gRPC
lsof -i :8081   # Realtime WebSocket
```

### 4. 数据库错误

```bash
# 确保数据库已启动
mage docker:up

# 运行迁移
mage db:up

# 检查数据库连接
psql -h localhost -U anychat -d anychat
```

## 持续集成 (CI)

可以将测试脚本集成到 CI/CD 流程中：

```yaml
# .github/workflows/test.yml 示例
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Install dependencies
        run: |
          sudo apt-get install -y jq
          go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

      - name: Start services
        run: |
          mage docker:up
          mage db:up
          GRPC_ADDR=0.0.0.0:50051 mage dev:user &
          GRPC_ADDR=0.0.0.0:50052 mage dev:friend &
          HTTP_ADDR=0.0.0.0:8081 mage dev:realtime &
          sleep 10  # 等待服务启动

      - name: Run API tests
        run: ./tests/api/test-all.sh
```

## 性能测试

对于负载测试，可以使用以下工具：

- **HTTP API**: [Apache Bench](https://httpd.apache.org/docs/2.4/programs/ab.html), [wrk](https://github.com/wg/wrk)
- **gRPC API**: [ghz](https://ghz.sh/)

示例：

```bash
# 安装 ghz
go install github.com/bojand/ghz/cmd/ghz@latest

# 压力测试登录接口
ghz --insecure \
  --proto api/user/v1/users.proto \
  --call user.v1.UserService/Login \
  -d '{
    "account": "13800138000",
    "password": "Test@123456",
    "device_type": "DEVICE_TYPE_IOS",
    "device_id": "device-001"
  }' \
  -c 10 \
  -n 1000 \
  localhost:50051
```

## 贡献指南

如需添加新的测试用例：

1. 在相应的测试脚本中添加新的测试函数
2. 在 `main()` 函数中调用新的测试函数
3. 更新本文档
4. 提交 Pull Request

## 反馈和问题

如果发现 API 问题或测试脚本错误，请提交 Issue 或联系开发团队。