# AnyChat 测试策略

## 当前测试架构（已优化）

### 1. API 测试（HTTP/REST 通过 Gateway）
**位置**: `tests/api/`
**协议**: HTTP/REST (通过Gateway)
**语言**: Bash + curl + jq

**结构**:
```
tests/api/
├── test-all.sh           # 统一测试入口（推荐使用）
├── common.sh            # 共享函数库
├── README.md            # 详细测试文档
├── user/
│   └── test-user-api.sh    # User Service (含认证) API 测试（16个用例）
└── friend/
    └── test-friend-api.sh  # Friend Service API 测试（12个用例）
```

**特点**:
- ✅ 模块化组织，易于维护
- ✅ 共享函数库减少重复代码
- ✅ 独立测试每个 HTTP API 端点
- ✅ 快速执行、易于调试
- ✅ 无需 Go 环境，只需 bash/curl/jq
- ✅ 统一入口 `test-all.sh` 可运行所有测试

### 2. 单元测试
**位置**: `app/*/internal/service/*_test.go`
**语言**: Go + gomock
**特点**:
- ✅ 测试业务逻辑
- ✅ 使用 mock 隔离依赖
- ✅ 运行命令：`mage test:all`

---

## 历史改动说明

### 已删除的测试文件
以下测试文件因功能重复或已被替代而删除：

1. **`tests/integration/*_service_test.go`** - Go gRPC 集成测试
   - 原因：与 HTTP API 测试功能重复 90%
   - 替代方案：统一使用 `tests/api/` 下的 Shell 脚本测试

2. **`tests/e2e/test-e2e.sh`** - 端到端场景测试
   - 原因：场景测试可通过组合调用模块化 API 测试实现
   - 替代方案：使用 `tests/api/test-all.sh` 运行所有模块测试

3. **`scripts/debug-friend-api.sh`** - 调试脚本
   - 原因：测试脚本本身已提供详细日志
   - 替代方案：直接运行 `tests/api/friend/test-friend-api.sh`

4. **`scripts/test-api.sh`** - 旧的混合测试脚本
   - 原因：Auth 和 User API 测试混在一起，不便维护
   - 替代方案：合并为 `tests/api/user/test-user-api.sh`（User Service 含认证功能）

### 路径迁移对照表
| 旧路径 | 新路径 | 说明 |
|--------|--------|------|
| `scripts/test-api.sh` | `tests/api/user/test-user-api.sh` | Auth + User 合并为一个模块 |
| `tests/api/auth/*` | `tests/api/user/test-user-api.sh` | 认证功能已合并入 User Service，目录已删除 |
| `scripts/test-friend-api.sh` | `tests/api/friend/test-friend-api.sh` | 迁移到 tests/api |
| `scripts/debug-friend-api.sh` | 已删除 | 功能已集成到测试脚本 |
| `tests/integration/*_service_test.go` | 已删除 | 统一使用 Shell API 测试 |
| `tests/e2e/test-e2e.sh` | 已删除 | 功能由 `tests/api/test-all.sh` 替代 |
| N/A | `tests/api/test-all.sh` | 新增统一测试入口 |
| N/A | `tests/api/common.sh` | 新增共享函数库 |

---

## 采用的测试分层策略（方案A）

```
┌─────────────────────────────────────────────────────────────┐
│  Level 2: API Contract Tests (API 契约测试)                  │
│  文件: tests/api/test-all.sh + 各模块测试                     │
│  目的: 测试 Gateway HTTP API 的完整性和正确性                 │
│  频率: 每次提交、CI/CD                                        │
└─────────────────────────────────────────────────────────────┘
                             ▲
                             │
┌─────────────────────────────────────────────────────────────┐
│  Level 1: Unit Tests (单元测试)                              │
│  文件: app/*/internal/service/*_test.go                      │
│  目的: 测试业务逻辑正确性                                     │
│  频率: 开发过程中                                             │
└─────────────────────────────────────────────────────────────┘
```

**简化原因**:
- HTTP API 是对外契约，必须严格测试
- gRPC 是内部实现，通过 HTTP 测试间接覆盖
- 场景测试可通过组合 API 测试实现
- 简化维护，避免重复

---

## 当前实施方案（方案A）

### 已完成的改动

#### 1. 测试文件重组织
```bash
# 新的测试结构
tests/api/
├── test-all.sh                 # ✅ 统一测试入口
├── common.sh                   # ✅ 共享函数库
├── README.md                   # ✅ 详细测试文档
├── user/
│   └── test-user-api.sh       # ✅ User (含认证) API 测试
└── friend/
    └── test-friend-api.sh     # ✅ Friend API 测试

# 已删除的文件
❌ tests/api/auth/ （已合并到 tests/api/user/，认证功能归入 User Service）
❌ tests/integration/auth_service_test.go（已合并至 user）
❌ tests/integration/user_service_test.go
❌ tests/integration/friend_service_test.go
❌ tests/integration/helpers.go
❌ tests/e2e/test-e2e.sh
❌ scripts/test-api.sh
❌ scripts/debug-friend-api.sh
```

### 职责划分（最终版）
| 测试层级 | 文件 | 测试内容 | 何时运行 |
|---------|------|---------|---------|
| **API 测试** | `tests/api/*/test-*-api.sh` | 单个服务的所有 HTTP API | 本地开发、PR、CI |
| **统一入口** | `tests/api/test-all.sh` | 运行所有 API 测试 | PR、发布前 |
| **单元测试** | `app/*/internal/service/*_test.go` | 业务逻辑 | 本地开发 |

### 执行方式
```bash
# 开发阶段 - 测试单个模块
./tests/api/user/test-user-api.sh
./tests/api/friend/test-friend-api.sh

# 完整测试 - 运行所有 API 测试
./tests/api/test-all.sh

# CI/CD 流程
mage test:all                # 运行所有测试（含单元测试）
./tests/api/test-all.sh     # 运行所有 API 测试
```

---

## 测试覆盖矩阵

| 功能模块 | API 测试 | 单元测试 | 说明 |
|---------|---------|---------|------|
| **User Service（含认证）** | ✓ | ✓ | 注册、登录、Token 刷新、密码修改、登出、个人资料、设置、二维码、推送 Token |
| **Friend Service** | ✓ | ✓ | 好友申请、列表、黑名单、删除 |
| **跨模块场景** | ✓ | - | 通过组合运行多个模块测试实现 |
| **业务逻辑** | - | ✓ | Repository、Service 层单元测试 |

---

## 总结

### 采用的方案
**方案A**：简化测试架构，统一使用 HTTP API 测试
- ✅ 删除重复的 gRPC 集成测试
- ✅ 删除可通过组合实现的 E2E 场景测试
- ✅ 模块化组织 API 测试
- ✅ 提供统一测试入口
- ✅ 共享函数库减少重复代码

### 优势
1. **维护简单**: 只需维护一套 API 测试
2. **清晰分层**: API 测试 + 单元测试
3. **易于扩展**: 新增服务只需添加对应的测试脚本
4. **快速执行**: Shell 脚本测试执行快速，无需编译
5. **灵活组合**: 可独立运行单个模块，也可运行全套测试

### 使用指南
```bash
# 快速开始
./tests/api/test-all.sh

# 查看详细文档
cat tests/api/README.md

# 开发新服务时
# 1. 创建 tests/api/<service>/test-<service>-api.sh
# 2. 更新 tests/api/test-all.sh 添加新测试
# 3. 更新 tests/api/README.md 文档
```
