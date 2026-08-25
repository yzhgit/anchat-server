#!/bin/bash
#
# Start all microservices
# Starts all services in dependency order
#

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}➜ $1${NC}"
}

# Check if port is available
check_port_available() {
    local port=$1
    local result=$(lsof -i :$port 2>/dev/null | grep LISTEN || echo "")
    if [ -z "$result" ]; then
        return 0
    else
        return 1
    fi
}

# Wait for service to start (check gRPC or HTTP port)
wait_for_service() {
    local port=$1
    local service=$2
    local max_attempts=30
    local attempt=0

    print_info "Waiting for $service to start (port $port)..."

    while [ $attempt -lt $max_attempts ]; do
        if lsof -i :$port 2>/dev/null | grep LISTEN > /dev/null; then
            print_success "$service started"
            return 0
        fi
        sleep 1
        ((attempt++))
        echo -n "."
    done

    echo ""
    print_error "$service start timeout"
    return 1
}

# Generic function to start a single service
# Usage: start_service <service_name> <mage_target> <check_port> <pid_file>
start_service() {
    local name=$1
    local mage_target=$2
    local port=$3
    local pid_file="/tmp/${name}.pid"

    if check_port_available "$port"; then
        print_info "Starting ${name}..."
        nohup mage "$mage_target" 2>&1 &
        local pid=$!
        echo $pid > "$pid_file"

        if wait_for_service "$port" "$name"; then
            print_success "${name} running on PID: $pid"
        else
            print_error "${name} failed to start, check logs: /tmp/anychat-logs/${name}.log"
            exit 1
        fi
    else
        print_success "${name} is already running"
    fi
}

# Check infrastructure
check_infrastructure() {
    print_header "Checking Infrastructure"

    local failed=0

    if ! docker ps | grep anychat-postgres | grep -q "healthy\|Up"; then
        print_error "PostgreSQL not running, please run: mage docker:up"
        ((failed++))
    else
        print_success "PostgreSQL is running"
    fi

    if ! docker ps | grep anychat-redis | grep -q "healthy\|Up"; then
        print_error "Redis not running, please run: mage docker:up"
        ((failed++))
    else
        print_success "Redis is running"
    fi

    if ! docker ps | grep anychat-nats | grep -q "Up"; then
        print_error "NATS not running, please run: mage docker:up"
        ((failed++))
    else
        print_success "NATS is running"
    fi

    if ! docker ps | grep anychat-minio | grep -q "healthy\|Up"; then
        print_error "MinIO not running, please run: mage docker:up"
        ((failed++))
    else
        print_success "MinIO is running"
    fi

    if ! docker ps | grep anychat-livekit | grep -q "Up"; then
        print_error "LiveKit not running, please run: mage docker:up"
        ((failed++))
    else
        print_success "LiveKit is running"
    fi

    if [ $failed -gt 0 ]; then
        echo ""
        print_error "Infrastructure not ready, please start: mage docker:up"
        exit 1
    fi

    echo ""
    print_success "All infrastructure ready"
}

# Check database migrations
check_migrations() {
    print_header "Checking Database Migrations"

    if docker exec anychat-postgres psql -U anychat -d anychat -c "\dt" 2>/dev/null | grep -q users; then
        print_success "Database migrations completed"
    else
        print_info "Database migrations not completed, running migrations..."
        mage db:up
        print_success "Database migrations completed"
    fi
}

# Start core domain services (first layer, no inter-dependencies)
start_core_services() {
    print_header "Starting Core Domain Services"

    start_service "user"    "dev:user"    50051
    start_service "friend"  "dev:friend"  50052
    start_service "group"   "dev:group"   50053
    start_service "file"    "dev:file"    50056
}

# Start application layer services (second layer, message/conversation)
start_app_services() {
    print_header "Starting Application Layer Services"

    start_service "message" "dev:message"  50054
    start_service "conversation" "dev:conversation"  50055
}

# Start auxiliary services (third layer, push/rtc)
start_auxiliary_services() {
    print_header "Starting Auxiliary Services"

    start_service "push"  "dev:push"  50058
    start_service "rtc"   "dev:rtc"   50057
}

# Start gateway service
start_gateway() {
    print_header "Starting Gateway Service"

    start_service "gateway" "dev:gateway" 8080
}

# Start realtime service
start_realtime() {
    print_header "Starting Realtime Service"

    start_service "realtime" "dev:realtime" 8081
}

# Show service status
show_status() {
    print_header "Service Status"

    echo -e "${YELLOW}Core Domain Services:${NC}"
    echo "  user:     grpc://localhost:50051  /tmp/anychat-logs/user.log"
    echo "  friend:   grpc://localhost:50052  /tmp/anychat-logs/friend.log"
    echo "  group:    grpc://localhost:50053  /tmp/anychat-logs/group.log"
    echo "  file:     grpc://localhost:50056  /tmp/anychat-logs/file.log"

    echo -e "\n${YELLOW}Application Layer Services:${NC}"
    echo "  message:  grpc://localhost:50054  /tmp/anychat-logs/message.log"
    echo "  conversation:  grpc://localhost:50055  /tmp/anychat-logs/conversation.log"

    echo -e "\n${YELLOW}Auxiliary Services:${NC}"
    echo "  push:     grpc://localhost:50058  /tmp/anychat-logs/push.log"
    echo "  rtc: grpc://localhost:50057  /tmp/anychat-logs/rtc.log"

    echo -e "\n${YELLOW}Gateway Service:${NC}"
    echo "  gateway:  http://localhost:8080  /tmp/anychat-logs/gateway.log"
    echo "  Swagger UI:       http://localhost:8080/swagger/index.html"

    echo -e "\n${YELLOW}Realtime Service:${NC}"
    echo "  realtime:  http://localhost:8081  /tmp/anychat-logs/realtime.log"

    echo -e "\n${YELLOW}Stop All Services:${NC}"
    echo "  ./scripts/stop-services.sh"
}

# Main function
main() {
    echo -e "${GREEN}"
    echo "╔═══════════════════════════════════════════╗"
    echo "║   AnyChat Service Startup Script          ║"
    echo "╚═══════════════════════════════════════════╝"
    echo -e "${NC}"

    check_infrastructure
    check_migrations

    start_core_services
    sleep 1
    start_app_services
    sleep 1
    start_auxiliary_services
    sleep 1
    start_gateway
    sleep 1
    start_realtime

    show_status

    echo -e "\n${GREEN}✓ All services started successfully!${NC}\n"
}

main "$@"