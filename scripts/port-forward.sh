#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Building Docker images for agentlauncher-distributed...${NC}"

# Docker registry
REGISTRY="agentlauncher"

# Build services
services=("agent-launcher" "agent-runtime" "llm-runtime" "tool-runtime" "message-runtime")

for service in "${services[@]}"; do
    echo -e "${YELLOW}Building $service...${NC}"
    docker build -f deployments/docker/Dockerfile.$service -t $REGISTRY/$service:latest .
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $service built successfully${NC}"
    else
        echo -e "${RED}✗ Failed to build $service${NC}"
        exit 1
    fi
done

echo -e "${GREEN}All images built successfully!${NC}"

# List built images
echo -e "\n${YELLOW}Built images:${NC}"
docker images | grep $REGISTRY