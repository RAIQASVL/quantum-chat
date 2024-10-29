#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

check_service() {
    local name=$1
    local port=$2
    local endpoint=$3
    
    echo -n "Checking $name... "
    if curl -s -f "http://localhost:$port$endpoint" > /dev/null; then
        echo -e "${GREEN}OK${NC}"
        return 0
    else
        echo -e "${RED}FAILED${NC}"
        return 1
    fi
}

show_logs() {
    local service=$1
    echo -e "\n${GREEN}Last 10 lines of $service logs:${NC}"
    docker logs quantum-chat-$service --tail 10
}

while true; do
    clear
    echo "Quantum Chat Service Monitor"
    echo "=========================="
    date
    echo

    # Check services
    check_service "Go Server" "8080" "/health"
    check_service "Python API" "8000" "/health"
    check_service "Nginx" "80" "/health"

    # Show container status
    echo -e "\n${GREEN}Container Status:${NC}"
    docker-compose ps

    # Show resource usage
    echo -e "\n${GREEN}Resource Usage:${NC}"
    docker stats --no-stream

    # Show abbreviated logs
    show_logs "go-server"
    show_logs "python-api"
    show_logs "postgres"

    sleep 10
done