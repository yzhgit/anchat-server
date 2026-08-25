# 编写 API 文档注释

本指南说明如何在 proto 文件中编写注释，以便通过 go-kratos 工具链自动生成 OpenAPI 3.0 规范文档。

## 概述

AnyChat 使用 go-kratos 框架的原生代码生成能力，通过 `protoc-gen-openapi` 插件直接从 `.proto` 文件生成 OpenAPI 3.0 规范。API 文档与接口定义保持单一数据源，proto 文件即文档。

**核心原则**: proto 文件中的注释和 `google/api/http` 注解驱动 OpenAPI 文档生成，无需在 Go handler 中额外编写文档注释。

## 生成流程

```
.proto 文件
  │  (protoc + protoc-gen-openapi)
  ▼
docs/api/swagger/openapi.json  (OpenAPI 3.0 规范)
  │  (Docsify + docsify-openapi 插件)
  ▼
交互式 API 文档页面
```

### 生成命令

```bash
# 生成 protobuf 代码 + OpenAPI 文档（一步完成）
mage proto

# 仅生成 OpenAPI 文档
mage docs:generate

# 本地预览文档
mage docs:serve
```

## Proto 文件注释规范

### 服务级别注释

在 `service` 关键字前添加服务描述：

```proto
// UserService 用户服务，提供用户注册、登录、Token 管理等接口
service UserService {}
```

### RPC 方法注释

在每个 `rpc` 方法前添加描述：

```proto
// Register 用户注册
rpc Register(RegisterRequest) returns (RegisterResponse);

// Login 用户登录
rpc Login(LoginRequest) returns (LoginResponse);
```

> 注释内容会映射到 OpenAPI 的 `summary` 字段。注释的第一行作为摘要，后续行作为详细描述。

### 消息类型注释

在 `message` 关键字前添加描述：

```proto
// LoginRequest 登录请求
message LoginRequest {
  // 账号（手机号或邮箱）
  string account = 1;

  // 密码
  string password = 2;

  // 设备类型
  DeviceType device_type = 3;

  // 设备 ID
  string device_id = 4;
}
```

字段注释会映射到 OpenAPI schema 的 `description` 属性。

## HTTP 路由注解

使用 `google.api.http` 注解将 gRPC 方法映射为 HTTP 端点。

### 引入依赖

```proto
import "google/api/annotations.proto";
```

### 基本用法

```proto
rpc Login(LoginRequest) returns (LoginResponse) {
  option (google.api.http) = {
    post: "/user/login"
    body: "*"
  };
}
```

### 路径参数

```proto
rpc GetProfile(GetProfileRequest) returns (GetProfileResponse) {
  option (google.api.http) = {
    get: "/users/{user_id}"
  };
}
```

### 其他 HTTP 方法

```proto
// PUT 请求（更新资源）
rpc UpdateProfile(UpdateProfileRequest) returns (UpdateProfileResponse) {
  option (google.api.http) = {
    put: "/users/me"
    body: "*"
  };
}

// DELETE 请求（删除资源）
rpc DeleteFriend(DeleteFriendRequest) returns (DeleteFriendResponse) {
  option (google.api.http) = {
    delete: "/friends/{friend_id}"
  };
}
```

### 支持的 HTTP 方法

| HTTP 方法 | 语义         | 示例                        |
|-----------|-------------|-----------------------------|
| `get`     | 获取资源     | `get: "/users/{user_id}"`   |
| `post`    | 创建资源     | `post: "/user/login"`       |
| `put`     | 更新资源     | `put: "/users/me"`          |
| `patch`   | 部分更新     | `patch: "/users/me"`        |
| `delete`  | 删除资源     | `delete: "/friends/{id}"`   |

## 认证配置

OpenAPI 中的 Bearer Token 认证信息在 proto 文件顶部通过全局注释配置：

```proto
syntax = "proto3";

package anychat.auth;

// AnyChat Gateway API 文档
//
// description: AnyChat 即时通讯系统网关 API 服务，提供用户认证、用户管理等 HTTP 接口。
// 所有需要认证的接口必须在 Header 中包含 Authorization: Bearer <token>。
//
// title: AnyChat Gateway API
// version: 1.0
// contact.name: AnyChat API Support
// contact.url: https://github.com/yzhgit/anychat-server
// contact.email: support@anychat.example.com
// license.name: MIT
// license.url: https://opensource.org/licenses/MIT
```

需要认证的接口会在 OpenAPI 中自动标记 `security: BearerAuth`。

## 枚举类型

枚举值和枚举项的注释会被正确映射：

```proto
// 设备类型
enum DeviceType {
  DEVICE_TYPE_UNSPECIFIED = 0;
  DEVICE_TYPE_IOS = 1;       // iOS
  DEVICE_TYPE_ANDROID = 2;   // Android
  DEVICE_TYPE_WEB = 3;       // Web
  DEVICE_TYPE_PC = 4;        // PC 客户端
  DEVICE_TYPE_H5 = 5;        // H5
}
```

## 完整示例

以下是一个完整的 proto 文件示例：

```proto
syntax = "proto3";

package anychat.auth;

import "google/api/annotations.proto";

// UserService 用户服务
service UserService {

  // Register 用户注册
  rpc Register(RegisterRequest) returns (RegisterResponse) {
    option (google.api.http) = {
      post: "/user/register"
      body: "*"
    };
  }

  // Login 用户登录
  rpc Login(LoginRequest) returns (LoginResponse) {
    option (google.api.http) = {
      post: "/user/login"
      body: "*"
    };
  }

  // RefreshToken 刷新 Token
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse) {
    option (google.api.http) = {
      post: "/user/refresh"
      body: "*"
    };
  }
}

// RegisterRequest 注册请求
message RegisterRequest {
  // 手机号
  string phone_number = 1;

  // 密码
  string password = 2;

  // 验证码
  string verify_code = 3;

  // 昵称
  string nickname = 4;

  // 设备类型
  DeviceType device_type = 5;

  // 设备 ID
  string device_id = 6;
}

// LoginRequest 登录请求
message LoginRequest {
  // 账号（手机号或邮箱）
  string account = 1;

  // 密码
  string password = 2;

  // 设备类型
  DeviceType device_type = 3;

  // 设备 ID
  string device_id = 4;
}

// LoginResponse 登录响应
message LoginResponse {
  // 用户 ID
  string user_id = 1;

  // 访问令牌
  string access_token = 2;

  // 刷新令牌
  string refresh_token = 3;

  // 过期时间（秒）
  int64 expires_in = 4;
}

// 设备类型
enum DeviceType {
  DEVICE_TYPE_UNSPECIFIED = 0;
  DEVICE_TYPE_IOS = 1;
  DEVICE_TYPE_ANDROID = 2;
  DEVICE_TYPE_WEB = 3;
  DEVICE_TYPE_PC = 4;
  DEVICE_TYPE_H5 = 5;
}
```

## 生成文档

### 本地生成

```bash
# 完整生成（protobuf 代码 + OpenAPI 文档）
mage proto

# 仅生成 OpenAPI 文档
mage docs:generate

# 验证文档完整性
mage docs:validate

# 本地预览（http://localhost:3000）
mage docs:serve
```

### CI/CD 自动生成

文档在以下情况自动生成：
- 推送到 `main` 分支时自动部署到 GitHub Pages
- Pull Request 时自动验证文档生成结果

## 最佳实践

### 1. 注释简洁准确

- 服务/方法注释用一句话概括功能
- 字段注释说明字段用途和约束
- 避免冗余信息

### 2. 保持 proto 注释与业务逻辑同步

修改接口时同步更新注释：

```bash
# 修改 proto 文件后
vim api/user/v1/user.proto

# 重新生成代码和文档
mage proto
```

### 3. 使用语义化路径

HTTP 路径遵循 RESTful 规范：

```proto
// 好的路径设计
get: "/users/{user_id}"        // 获取用户
post: "/friends"               // 添加好友
delete: "/friends/{friend_id}" // 删除好友

// 避免
post: "/getUserInfo"           // 动词风格
get: "/do-search-users"        // 不清晰
```

### 4. 枚举值使用有意义的名称

```proto
enum DeviceType {
  DEVICE_TYPE_UNSPECIFIED = 0;  // 明确的默认值
  DEVICE_TYPE_IOS = 1;          // 清晰命名
}
```

### 5. 及时重新生成文档

每次修改 `.proto` 文件后运行：

```bash
mage proto
```

## 常见问题

### Q: 修改了 proto 文件但文档没有更新？

运行 `mage proto` 重新生成 OpenAPI 文档。

### Q: 某个接口没有出现在文档中？

确保：
1. 该 RPC 方法添加了 `option (google.api.http)` 注解
2. 运行了 `mage proto` 生成文档

### Q: 如何调试 OpenAPI 生成？

```bash
# 查看生成的 openapi.json
cat docs/api/swagger/openapi.json | jq '.paths'

# 查看特定路径
cat docs/api/swagger/openapi.json | jq '.paths["/user/login"]'
```

### Q: 注释中的中文是否正常显示？

是的，OpenAPI 3.0 支持 UTF-8 编码，中文注释会被正确保留。

## 参考资料

- [go-kratos OpenAPI 指南](https://go-kratos.dev/zh-cn/docs/guide/openapi/)
- [protoc-gen-openapi](https://github.com/go-kratos/kratos/tree/main/cmd/protoc-gen-openapi)
- [google/api/http 注解](https://cloud.google.com/endpoints/docs/grpc/transcoding)
- [OpenAPI 3.0 规范](https://swagger.io/specification/)