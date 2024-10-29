#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Starting Quantum Chat services...${NC}"

# Stop any running containers
echo "Stopping existing containers..."
docker-compose down -v

# Build and start services
echo "Building and starting services..."
docker-compose up -d --build

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
sleep 10

# Initialize database
echo "Initializing database..."
docker exec -i quantum-chat-postgres psql -U user -d chatdb << EOF
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    public_key BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER REFERENCES users(id),
    receiver_id INTEGER REFERENCES users(id),
    content BYTEA NOT NULL,
    timestamp BIGINT NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_receiver ON messages(receiver_id);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
EOF

echo -e "${GREEN}Database initialized successfully${NC}"

# Check service health
echo "Checking service health..."
docker-compose ps

# Function to test endpoint
test_endpoint() {
    local url=$1
    local name=$2
    echo -n "Testing $name... "
    if curl -s -f "$url" > /dev/null; then
        echo -e "${GREEN}OK${NC}"
    else
        echo -e "${RED}FAILED${NC}"
    fi
}

# Wait a bit for services to start
sleep 5

# Test endpoints
test_endpoint "http://localhost:8080/health" "Go server"
test_endpoint "http://localhost:8000/health" "Python API"
test_endpoint "http://localhost/health" "Nginx"

echo -e "\n${GREEN}Setup complete!${NC}"
echo "You can now access:"
echo " - Web interface: http://localhost"
echo " - API docs: http://localhost:8000/docs"
echo " - API endpoints: http://localhost/api/*"

# Show logs
echo -e "\nService logs:"
docker-compose logs