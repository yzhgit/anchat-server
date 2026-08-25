#!/bin/bash
#
# Stop all microservices
#

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_info() {
    echo -e "${YELLOW}➜ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

echo -e "${YELLOW}Stopping all microservices...${NC}\n"

# Stop all microservice processes
pkill -f "user|friend|group|file|message|conversation|push|rtc|gateway|realtime" 2>/dev/null || true

# Wait for processes to end
sleep 2

# Check if any processes remain
remaining=$(ps aux | grep -E "user|friend|group|file|message|conversation|push|rtc|gateway|realtime" | grep -v grep || echo "")

if [ -z "$remaining" ]; then
    print_success "All microservices stopped"
else
    print_info "Some processes still running, forcing stop..."
    pkill -9 -f "user|friend|group|file|message|conversation|push|rtc|gateway|realtime" 2>/dev/null || true
    sleep 1
    print_success "Force stop completed"
fi

# Clean up PID files
rm -f \
    /tmp/user.pid \
    /tmp/friend.pid \
    /tmp/group.pid \
    /tmp/file.pid \
    /tmp/message.pid \
    /tmp/conversation.pid \
    /tmp/push.pid \
    /tmp/rtc.pid \
    /tmp/gateway.pid \
    /tmp/realtime.pid \
    2>/dev/null

# Show port status
echo ""
print_info "Port status check:"
for port in \
    50051 50052 50053 50054 50055 50056 50057 50058 \
    8080 8081; do
    if lsof -i :$port 2>/dev/null | grep LISTEN > /dev/null; then
        echo "  Port $port: still in use"
    else
        echo "  Port $port: free ✓"
    fi
done

echo -e "\n${GREEN}Done!${NC}"