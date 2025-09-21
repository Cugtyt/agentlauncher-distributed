#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Agentlauncher Distributed Quick Start ===${NC}\n"

# Build images
echo -e "${YELLOW}Step 1: Building Docker images...${NC}"
./scripts/build-images.sh

# Deploy
echo -e "\n${YELLOW}Step 2: Deploying to Kubernetes...${NC}"
./scripts/deploy-k8s.sh

# Port forward
echo -e "\n${YELLOW}Step 3: Setting up port forwarding...${NC}"
echo -e "${GREEN}The API will be available at http://localhost:8080${NC}"
./scripts/port-forward.sh