# 文档系统使用指南

本指南说明如何使用 AnyChat 的文档系统。

## 快速开始

### 查看文档

#### 在线查看

访问部署在 GitHub Pages 的文档站点：
- https://yzhgit.github.io/anychat-server/

#### 本地查看

1. 生成 API 文档：
   ```bash
   mage docs:generate
   ```

2. 启动文档服务器：
   ```bash
   mage docs:serve
   ```

3. 在浏览器中打开: http://localhost:3000

### 编写 API 文档

为 Gateway HTTP 接口添加 Swagger 注释。详细说明请参考：
- [API 文档编写指南](development/writing-api.md)

### 更新文档

1. 修改代码或文档文件
2. 重新生成 API 文档（如果修改了 Gateway 代码）：
   ```bash
   mage docs:generate
   ```
3. 提交更改
4. 推送到 `main` 分支后，文档会自动部署到 GitHub Pages

## 文档结构

```
docs/
├── index.html              # Docsify 配置
├── .nojekyll              # 禁用 Jekyll
├── README.md              # 文档首页
├── _sidebar.md            # 侧边栏导航
├── api/
│   ├── QUICKSTART.md      # API 快速开始
│   ├── gateway-http-api.md # Gateway API 文档（包含交互式界面）
│   └── swagger/           # 自动生成的 OpenAPI 文件
│       └── openapi.json   # OpenAPI 3.0 规范（protoc-gen-openapi 生成）
├── development/
│   ├── getting-started.md
│   ├── writing-api.md
│   ├── service-startup.md
│   └── port-allocation.md
└── design/
    └── backend-design.md
```

## Mage 文档命令

### docs:generate

生成 API 文档（从 Go 代码注释生成 OpenAPI 规范）

```bash
mage docs:generate
```

**何时使用**:
- 添加或修改了 proto 文件中的 RPC 方法
- 更改了 `google/api/http` 注解
- 修改了消息类型定义

**输出**:
- `docs/api/swagger/openapi.json` - OpenAPI 3.0 规范

### docs:serve

启动本地文档服务器

```bash
mage docs:serve
```

**功能**:
- 在 http://localhost:3000 启动 Docsify 服务器
- 支持热重载
- 包含完整的文档站点和交互式 API 文档

**前置要求**:
- 需要安装 Node.js 和 npm
- 会自动安装 docsify-cli（如果未安装）

### docs:build

构建静态文档站点

```bash
mage docs:build
```

**功能**:
- 生成 API 文档
- 准备好用于部署的静态文件

**用途**:
- 手动部署前
- CI/CD 流程中

### docs:validate

验证 API 文档

```bash
mage docs:validate
```

**功能**:
- 生成文档
- 验证 openapi.json 是否存在
- 检查文档完整性

**用途**:
- CI/CD 流程中
- 提交前检查

## 技术栈

### Swagger 生成

- **工具**: [protoc-gen-openapi](https://github.com/go-kratos/kratos/tree/main/cmd/protoc-gen-openapi)（go-kratos 原生插件）
- **源文件**: `api/proto/**/*.proto`（proto 文件中的注释和 `google/api/http` 注解）
- **输出**: OpenAPI 3.0 规范 (`docs/api/swagger/openapi.json`)

### 文档站点

- **框架**: [Docsify](https://docsify.js.org/)
- **主题**: Vue 主题
- **插件**:
  - docsify-openapi - 渲染 OpenAPI 规范
  - docsify-search - 全文搜索
  - docsify-copy-code - 代码复制
  - docsify-pagination - 分页导航
  - docsify-tabs - 标签页
  - docsify-mermaid - 图表支持

### 部署

- **平台**: GitHub Pages
- **自动化**: GitHub Actions
- **触发**: 推送到 main 分支

## 自动化流程

### CI/CD 工作流

位置: `.github/workflows/docs.yml`

#### 触发条件

**Push 事件**:
- 分支: `main`, `develop`
- 路径:
  - `cmd/realtime/**`
  - `internal/realtime/**`
  - `docs/**`
  - `.github/workflows/docs.yml`

**Pull Request 事件**:
- 目标分支: `main`
- 相同的路径过滤

#### Job: generate

1. Checkout 代码
2. 安装 Go 和依赖
3. 安装 protoc 插件
4. 生成 API 文档
5. 验证 openapi.json
6. 上传文档 artifact
7. (PR only) 如果文档有变化，在 PR 中添加评论

#### Job: deploy (仅 main 分支)

1. Checkout 代码
2. 下载文档 artifact
3. 配置 GitHub Pages
4. 上传站点文件
5. 部署到 GitHub Pages
6. 添加部署摘要

### 本地开发流程

1. **修改代码**
   ```bash
   # 编辑 proto 文件，更新注释和 google/api/http 注解
   vim api/user/v1/user.proto
   ```

2. **生成文档**
   ```bash
   mage proto
   # 或仅生成 OpenAPI 文档
   mage docs:generate
   ```

3. **本地预览**
   ```bash
   mage docs:serve
   # 访问 http://localhost:3000
   ```

4. **验证**
   ```bash
   mage docs:validate
   ```

5. **提交**
   ```bash
   git add .
   git commit -m "docs(gateway): update API documentation"
   git push
   ```

6. **查看部署**
   - GitHub Actions 会自动运行
   - 文档会部署到 GitHub Pages

## 常见问题

### Q: 文档没有更新？

A: 确保：
1. 运行了 `mage proto` 或 `mage docs:generate`
2. 提交了 `docs/api/swagger/openapi.json` 文件
3. 推送到了正确的分支（main）
4. GitHub Actions 工作流运行成功

### Q: 本地预览报错？

A: 检查：
1. Node.js 是否安装
2. 运行 `npm install -g docsify-cli` 手动安装
3. 端口 3000 是否被占用

### Q: Swagger UI 显示空白？

A: 确保：
1. `docs/api/swagger/openapi.json` 文件存在
2. 文件路径在 `index.html` 中正确配置
3. 浏览器控制台没有错误

### Q: 如何添加新的文档页面？

A:
1. 在 `docs/` 目录下创建 Markdown 文件
2. 更新 `docs/_sidebar.md` 添加导航链接
3. 提交并推送

### Q: 如何修改文档主题？

A: 编辑 `docs/index.html`：
- 修改 CSS 主题链接
- 调整 Docsify 配置
- 添加/删除插件

## 最佳实践

### 1. 保持文档同步

- 修改 proto 接口后立即更新注释
- 在 PR 中包含文档更改
- 定期检查文档的准确性

### 2. 保持注释准确

- proto 方法注释用一句话概括功能
- 字段注释说明字段用途和约束
- 避免冗余信息

### 3. 测试交互式功能

在本地预览时测试：
- "Try it out" 按钮
- 参数输入
- 响应显示
- 认证功能

### 5. 版本控制

- 提交生成的文档文件
- 记录文档变更历史
- 使用语义化版本号

## 参考资料

- [go-kratos OpenAPI 指南](https://go-kratos.dev/zh-cn/docs/guide/openapi/)
- [Docsify 官方文档](https://docsify.js.org/)
- [OpenAPI 规范](https://swagger.io/specification/)
- [Docsify OpenAPI 插件](https://github.com/mikefarah/docsify-openapi)
- [API 文档编写指南](development/writing-api.md)
