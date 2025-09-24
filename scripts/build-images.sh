#!/bin/bash

set -e

echo "Building Docker images for agentlauncher-distributed..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Generate version tag based on timestamp
VERSION=${1:-$(date +%Y%m%d-%H%M%S)}
echo -e "${YELLOW}Using version: $VERSION${NC}"

services=("agent-launcher" "agent-runtime" "llm-runtime" "tool-runtime" )

for service in "${services[@]}"; do
    echo -e "${YELLOW}Building $service:$VERSION...${NC}"
    docker build --no-cache -f deployments/docker/Dockerfile.$service -t agentlauncher/$service:$VERSION .
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $service:$VERSION built successfully${NC}"
    else
        echo -e "${RED}✗ Failed to build $service:$VERSION${NC}"
        exit 1
    fi
done

echo -e "${GREEN}All images built successfully!${NC}"
echo -e "${GREEN}Version: $VERSION${NC}"

# Load images into minikube for local development
echo -e "${YELLOW}Loading images into minikube...${NC}"
for service in "${services[@]}"; do
    minikube image load agentlauncher/$service:$VERSION
done

echo -e "\n${YELLOW}Built images:${NC}"
docker images | grep agentlauncher

# Save version to file for deployment script
echo $VERSION > .image-version