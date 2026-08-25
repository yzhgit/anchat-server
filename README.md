# AnyChat

A microservice-based instant messaging (IM) backend built with Go and the Kratos framework.

## Features

- Private & group chat
- Audio/video calls via LiveKit
- File transfer via MinIO
- Message read receipts
- Multi-device sync
- Offline push notifications

## Architecture

AnyChat consists of **10 services** communicating via gRPC (internal) and NATS (async):

| Layer | Service | Port |
|-------|---------|------|
| Gateway | **gateway** | HTTP 8080 |
| Gateway | **realtime** | HTTP/WS 8081 |
| Microservice | **user** | gRPC 50051 |
| Microservice | **friend** | gRPC 50052 |
| Microservice | **group** | gRPC 50053 |
| Microservice | **message** | gRPC 50054 |
| Microservice | **conversation** | gRPC 50055 |
| Microservice | **file** | gRPC 50056 |
| Microservice | **rtc** | gRPC 50057 |
| Microservice | **push** | gRPC 50058 |

Clients reach the system via HTTP (gateway) or WebSocket (realtime). Services discover each other through Consul.

```
Client (HTTP)  ->  gateway (8080)  ->  Microservices (gRPC)
Client (WS)    ->  realtime (8081)  ->  Microservices (gRPC)
                                       <-->  NATS JetStream (async)
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.26 |
| Framework | Go Kratos |
| Database | PostgreSQL 18.0 |
| Cache | Redis 7.0+ |
| Message Queue | NATS with JetStream |
| Object Storage | MinIO |
| Audio/Video | LiveKit |
| Service Discovery | Consul |
| Observability | Prometheus + Grafana + Tempo + Loki |
| Build Tool | Mage |

## Quick Start

### Requirements

- Go 1.26+
- Docker & Docker Compose
- protoc (for gRPC code generation)
- Mage: `go install github.com/magefile/mage@latest`

### Local Development

```bash
# 1. Install dependencies
mage deps

# 2. Start infrastructure (PostgreSQL, Redis, NATS, MinIO, etc.)
mage docker:up

# 3. Run database migrations
mage db:up

# 4. Start services (each in a separate terminal)
GRPC_ADDR=0.0.0.0:50051 mage dev:user
GRPC_ADDR=0.0.0.0:50052 mage dev:friend
HTTP_ADDR=0.0.0.0:8081 mage dev:realtime
# ... see docs/development/service-startup.md for the full list
```

For full port allocation and per-service commands, see [port-allocation.md](docs/development/port-allocation.md) and [service-startup.md](docs/development/service-startup.md).

## Build & Test

```bash
# List all commands
mage -l

# Build all services
mage build:all

# Run tests
mage test:all

# Lint / format
mage lint && mage fmt
```

## Documentation

- **API docs**: [docs/api/README.md](docs/api/README.md)
- **Getting started**: [docs/development/getting-started.md](docs/development/getting-started.md)
- **Service startup**: [docs/development/service-startup.md](docs/development/service-startup.md)
- **System design**: [docs/design/backend-design.md](docs/design/backend-design.md)
- **Writing API docs**: [docs/development/writing-api.md](docs/development/writing-api.md)

## License

MIT License — see [LICENSE](LICENSE)
