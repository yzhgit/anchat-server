# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AnyChat is a microservices-based instant messaging (IM) backend system written in Go, built on the **go-kratos** framework with a custom project structure (not standard kratos-layout). The system consists of **10 services** that communicate via gRPC internally, with NATS for async messaging. External HTTP requests are transcoded by **gateway** (port 8080), and WebSocket long connections are handled by the **realtime** service.

## Build System: Mage

This project uses **Mage** (not Make) as its build tool. Mage uses Go code to define build tasks.

### Essential Mage Commands

```bash
# Setup
mage deps                    # Install Go dependencies
mage install                 # Install dev tools (golangci-lint, migrate, mockgen, protoc-gen-go, etc.)

# Development
mage dev:message             # Run message service locally
mage dev:user                # Run user service locally
mage dev:realtime            # Run realtime service locally

# Building
mage build:all               # Build all services to bin/
mage proto                   # Generate protobuf code from api/**/*.proto
mage wire                    # Generate wire code

# Testing
mage test:all                # Run all tests with race detection and coverage
mage lint                    # Run golangci-lint
mage fmt                     # Format code with gofmt

# Infrastructure
mage docker:up               # Start PostgreSQL, Redis, NATS, MinIO, Prometheus, Grafana, Tempo, Loki, Promtail
mage docker:down             # Stop all containers
mage docker:logs             # Follow logs
mage docker:ps               # Show container status

# Database
mage db:up                   # Run migrations up
mage db:down                 # Rollback migrations

# Documentation
mage docs:generate           # Generate API documentation (OpenAPI)
mage docs:serve              # Start documentation server at http://localhost:3000
mage docs:build              # Build static documentation site
mage docs:validate           # Validate API documentation

# Other
mage clean                   # Remove bin/, coverage files
mage mock                    # Generate mock code
mage -l                      # List all available tasks
```

## Microservices Architecture

### The Services

1. **gateway**: HTTP API gateway, transcodes external HTTP to internal gRPC
2. **user**: Authentication and authorization (JWT tokens), user profile management
3. **friend**: Friend relationships and requests
4. **group**: Group chat management
5. **message**: Message storage and delivery
6. **conversation**: Session/conversation management
7. **file**: File uploads/downloads via MinIO
8. **rtc**: Real-Time Communication (audio/video) via LiveKit
9. **push**: Offline push notifications
10. **realtime**: WebSocket long-connection management

> **Note**: The Auth service has been merged into the User service. The Sync, Version, and Admin services have been removed.

### Service Structure Pattern

Each service lives under `app/<service-name>/` and follows this structure:

```
app/<service>/
├── cmd/             # Service entry point (main.go + Wire DI)
├── configs/         # Service-specific configuration (config.yaml)
└── internal/        # Business logic
    ├── handler/     # gRPC handlers (kratos framework)
    ├── server/      # Server registration
    ├── service/     # Business logic layer
    ├── repository/  # Data access layer (database operations)
    └── model/       # Database models (GORM)
```

The realtime service additionally contains `internal/websocket/` and `internal/notification/` packages.

### Service Communication

- **External HTTP**: Clients → gateway (port 8080) → transcode to backend microservice gRPC calls
- **External WebSocket**: Clients → realtime service (WebSocket long connections) → gRPC to backend
- **Internal**: Services communicate via gRPC, registered and discovered through Consul
- **Async**: Services publish/subscribe via NATS JetStream for event-driven operations
- **Config**: Each service reads from `app/<service-name>/configs/config.yaml` with environment variable overrides (`${VAR:default}`)

## Technology Stack

| Component | Technology | Port/Access |
|-----------|-----------|-------------|
| Language | Go 1.26 | — |
| Framework | Go Kratos | — |
| Database | PostgreSQL 18.0 | localhost:5432 (user: anychat, pass: anychat) |
| Cache | Redis 7.4+ | localhost:6379 |
| Message Queue | NATS with JetStream | localhost:4222 (client), :8222 (monitoring) |
| Object Storage | MinIO | localhost:9000 (API), :9091 (Console, admin/admin) |
| Audio/Video | LiveKit | ws://localhost:7880 |
| Service Discovery | Consul | localhost:8500 |
| API Gateway | gateway | :8080 |
| Metrics | Prometheus | localhost:9090 |
| Dashboards | Grafana | localhost:3000 (admin/admin) |
| Tracing | Tempo | localhost:3200, OTLP gRPC :4317, OTLP HTTP :4318 |
| Logging | Loki + Promtail | Loki localhost:3100 |
| Build Tool | Mage | — |

## Development Workflow

### Starting New Work

1. Start infrastructure: `mage docker:up`
2. Wait for health checks to pass: `mage docker:ps`
3. Run migrations: `mage db:up`
4. Start relevant service(s): `mage dev:user`, `mage dev:realtime`, etc.

### Adding a New Feature

1. **Database changes**: Create migration files in `migrations/` (`000XXX_<name>.up.sql` / `.down.sql`), then run `mage db:up`
2. **Proto changes**: Edit `.proto` files in `api/<service>/v1/`, then run `mage proto` to regenerate
3. **Code changes**: Follow the layered architecture (handler → service → repository → model)
4. **API documentation**: Add comments and `option (google.api.http)` annotations to proto files (see API Documentation section below)
5. **Testing**: Write tests, run `mage test:unit` or `mage test:all`
6. **Code quality**: `mage fmt && mage lint` before committing
7. **Documentation**: Run `mage docs:generate` if you added/modified microservice HTTP APIs

### Creating a New Service

**IMPORTANT**: When implementing a new service, you MUST complete ALL the following steps. Missing scripts, tests, or configuration updates is a common mistake.

#### Complete Checklist for New Service Implementation

##### 1. Database Layer
- [ ] Create migration files in `migrations/`
  - `000XXX_create_<name>_tables.up.sql`
  - `000XXX_create_<name>_tables.down.sql`
- [ ] Run `mage db:up` to apply migrations
- [ ] Verify tables created: `psql -U anychat -d im -c "\dt"`

##### 2. API Definition
- [ ] Create proto file in `api/<name>/v1/<name>.proto`
- [ ] Define gRPC service interface and message types
- [ ] Run `mage proto` to generate Go code
- [ ] Verify generated files: `api/<name>/v1/<name>.pb.go` and `<name>_grpc.pb.go`

##### 3. Service Implementation
- [ ] Create `app/<name>/cmd/main.go` — service entry point with Wire dependency injection
- [ ] Create `app/<name>/configs/config.yaml` — service configuration (gRPC addr, HTTP addr, DB, Redis, NATS, etc.)
- [ ] Create `app/<name>/internal/model/` — GORM models
- [ ] Create `app/<name>/internal/repository/` — data access layer with transaction support
- [ ] Create `app/<name>/internal/service/` — business logic layer
- [ ] Create `app/<name>/internal/handler/` — gRPC handler implementation
- [ ] Create `app/<name>/internal/server/` — server registration

##### 4. Gateway Integration (if HTTP API needed)
- [ ] Configure gateway routing to proxy to the new service's gRPC port

##### 5. Error Codes
- [ ] Update `pkg/errors/errors.go`:
  - Add error code constants following the pattern: user=10xxx, friend=20xxx, group=30xxx, etc.
  - Each service gets a dedicated 10xxx block

##### 6. Configuration
- [ ] Create `app/<name>/configs/config.yaml` following existing service templates

##### 7. Build System
- [ ] Update `magefile.go`:
  - Add service to `Build.All()` services list
  - Add `Build.<Name>()` method for individual builds
  - Add `Dev.<Name>()` method for local development

##### 8. Wire Dependency Injection
- [ ] Set up Wire in `app/<name>/cmd/` (wire.go, wire_gen.go)
- [ ] Run `wire gen ./app/<name>/cmd` to generate DI code

##### 9. Documentation
- [ ] Add comments and `option (google.api.http)` annotations to proto file
- [ ] Run `mage proto` to generate OpenAPI docs
- [ ] Verify OpenAPI UI shows all endpoints: `mage docs:serve`
- [ ] Update design docs under `docs/design/<name>/` if architectural changes were made

##### 10. Docker & Deployment (if needed)
- [ ] Update `deployments/docker/docker-compose.yml` if service needs to run in container

##### 11. Verification
- [ ] `mage build:<name>` — build succeeds
- [ ] `mage fmt && mage lint` — code quality passes
- [ ] `mage test:all` — tests pass
- [ ] Start infrastructure: `mage docker:up && mage db:up`
- [ ] `mage dev:<name>` — service starts successfully
- [ ] Check logs for errors
- [ ] Verify Consul registration

#### Common Mistakes to Avoid

1. ❌ **Missing proto comments / http annotations** — RPC methods won't appear in OpenAPI documentation
2. ❌ **Not updating error codes** — using wrong error code ranges; each service must use its own block
3. ❌ **Missing gateway routing configuration** — gateway won't proxy requests to the new service
4. ❌ **Wrong port allocation** — using conflicting ports; see `docs/development/port-allocation.md`
5. ❌ **Missing Consul registration** — other services won't be able to discover the new service

#### Port Allocation Guide

When creating a new service, assign ports following this pattern:
- gRPC API: `50051` (default in container), dev env uses sequential ports starting from 50051
- HTTP/metrics: `9101` (default in container), dev env uses sequential ports starting from 9101
- See `docs/development/port-allocation.md` for full port table

#### Reference Implementation

For a complete reference implementation, see the **friend** service:
- Database: `migrations/` (friend-related migration files)
- Proto: `api/friend/v1/friend.proto`
- Service: `app/friend/` (cmd, configs, internal/{handler,server,service,repository,model})
- Entry: `app/friend/cmd/main.go`
- Errors: `pkg/errors/errors.go` (codes 20101-20107)

## Shared Packages (pkg/)

The `pkg/` directory contains shared utilities used across services:

| Package | Purpose |
|---------|---------|
| `auth` | JWT authentication middleware and token management |
| `broker` | NATS JetStream publisher/subscriber abstraction |
| `cache` | Redis client wrapper with common operations |
| `config` | Configuration loading and validation |
| `consts` | Shared constants and enums |
| `crypto` | Encryption utilities (bcrypt, AES) |
| `database` | PostgreSQL/GORM database initialization |
| `errors` | Standardized error codes and response format |
| `grpc` | gRPC client/server configuration helpers |
| `jpush` | JPush offline push integration |
| `md` | Markdown/text processing utilities |
| `metrics` | Prometheus metrics registration and helpers |
| `oss` | MinIO object storage client wrapper |
| `registry` | Consul service registration/discovery |
| `sender` | SMS/email notification sender |
| `validator` | Input validation utilities |

## Configuration

Each service reads from `app/<service-name>/configs/config.yaml` with environment variable support using `${VAR:default}` syntax.

Example: `${NATS_URL:nats://127.0.0.1:4222}` uses the `NATS_URL` env var or defaults to `nats://127.0.0.1:4222`.

Key configuration sections in each service's config:
- `server.grpc.addr` / `server.http.addr` — gRPC and HTTP listen addresses
- `data.database` — PostgreSQL connection (host: localhost, port: 5432, user: anychat, password: anychat, dbname: im)
- `data.cache.addr` — Redis connection (127.0.0.1:6379)
- `data.broker.url` — NATS connection (nats://127.0.0.1:4222)
- `registry.consul` — Consul service registration (127.0.0.1:8500)
- `auth` — JWT secret and token expiry (user service)

When running in Docker, services use container names (postgres, redis, nats) as hostnames instead of localhost.

## Database Migrations

- Tool: golang-migrate
- Location: `migrations/` directory
- Connection string: `postgresql://anychat:anychat@localhost:5432/im?sslmode=disable`
- Files are sequential: `000001_<name>.up.sql` and `000001_<name>.down.sql`

## Testing Strategy

- Unit tests use `-short` flag to skip integration tests
- Integration tests require running infrastructure (`mage docker:up`)
- Mocks are generated with mockgen
- Coverage reports go to `coverage.out` and `coverage.html`

## Code Quality

The project uses golangci-lint with these enabled linters (see `.golangci.yml`):
- gofmt, golint, govet, errcheck, staticcheck
- unused, gosimple, ineffassign, deadcode
- goconst, gocyclo (complexity < 15), misspell

Generated protobuf files (`*.pb.go`) and test files are excluded from linting.

## Port Conventions

- **8080**: gateway (external HTTP entry point)
- **8081**: realtime service HTTP/WebSocket port
- **50051~50058**: Microservice gRPC ports (dev env)
- **9101~9108**: Microservice HTTP/metrics ports (dev env)
- **Full port table**: See `docs/development/port-allocation.md`

## Important Notes

- The codebase is in early stages — most services have skeleton main.go files with TODOs
- Services are designed to run both locally (via `mage dev:*`) and in Docker
- All services support graceful shutdown on SIGINT/SIGTERM
- Commit message format: `<type>(<scope>): <description>` (e.g., `feat(user): add user registration`)

## API Documentation

### Documentation System

The project uses **protoc-gen-openapi** (go-kratos native plugin) to generate OpenAPI 3.0 specifications directly from `.proto` files, and **Docsify** with **docsify-openapi** plugin to render an interactive documentation site.

**Core principle**: proto file comments and `google/api/http` annotations are the single source of truth for API documentation. No additional annotations needed in Go handler code.

### Writing API Documentation

When adding or modifying APIs, update the `.proto` file:

#### Proto Service/Method Comments

```proto
// UserService Authentication service
service UserService {

  // Login User login
  rpc Login(LoginRequest) returns (LoginResponse) {
    option (google.api.http) = {
      post: "api/v1/users/login"
      body: "*"
    };
  }
}
```

#### HTTP Route Annotations

Use `google.api.http` to map gRPC methods to HTTP endpoints:

```proto
import "google/api/annotations.proto";

// GET with path parameter
rpc GetProfile(GetProfileRequest) returns (GetProfileResponse) {
  option (google.api.http) = { get: "/users/{user_id}" };
}

// PUT with request body
rpc UpdateProfile(UpdateProfileRequest) returns (UpdateProfileResponse) {
  option (google.api.http) = {
    put: "/users/me"
    body: "*"
  };
}
```

Supported methods: `get`, `post`, `put`, `patch`, `delete`.

### Generating Documentation

```bash
# Generate protobuf code + OpenAPI docs (one step)
mage proto

# Generate OpenAPI docs only
mage docs:generate

# Preview documentation locally at http://localhost:3000
mage docs:serve

# Validate generated documentation
mage docs:validate
```

### Automatic Deployment

- Documentation is **automatically** generated and deployed to GitHub Pages on push to `main` branch
- CI validates documentation on Pull Requests
- Access deployed docs at: `https://yzhgit.github.io/anychat-server/`

### Documentation Structure

```
docs/
├── index.html              # Docsify configuration
├── .nojekyll              # Disable Jekyll processing
├── README.md              # Documentation homepage
├── _sidebar.md            # Sidebar navigation
├── api/
│   ├── QUICKSTART.md      # API quick start guide
│   ├── README.md          # API test documentation
│   └── swagger/           # Auto-generated OpenAPI files
│       └── openapi.json   # OpenAPI 3.0 specification
├── development/
│   ├── getting-started.md
│   ├── writing-api.md
│   └── ...
└── design/
    └── ...
```

### Best Practices

1. **Add comments to every proto service, rpc method, message, and field**
2. **Add `option (google.api.http)` to every rpc method that needs an HTTP endpoint**
3. **Run `mage proto`** before committing API changes
4. **Check documentation locally** with `mage docs:serve`

### Common Mistakes

- ❌ Forgetting `option (google.api.http)` on rpc methods → no HTTP endpoint in docs
- ❌ Not running `mage proto` after proto changes → stale documentation
- ❌ Missing field comments → less useful generated OpenAPI schemas

### Reference

- Full guide: `docs/development/writing-api.md`
- go-kratos OpenAPI guide: https://go-kratos.dev/docs/guide/openapi/
- OpenAPI spec: https://swagger.io/specification/