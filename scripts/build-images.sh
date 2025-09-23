#!/bin/bash

set -e

echo "Building Docker images for agentlauncher-distributed..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

services=("agent-launcher" "agent-runtime" "llm-runtime" "tool-runtime" )

for service in "${services[@]}"; do
    echo -e "${YELLOW}Building $service...${NC}"
    docker build -f deployments/docker/Dockerfile.$service -t agentlauncher/$service:latest .
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $service built successfully${NC}"
    else
        echo -e "${RED}✗ Failed to build $service${NC}"
        exit 1
    fi
done

echo -e "${GREEN}All images built successfully!${NC}"

echo -e "\n${YELLOW}Built images:${NC}"
docker images | grep agentlauncher