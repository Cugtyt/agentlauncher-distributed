#!/bin/bash

echo "Setting up port forwarding for agentlauncher API..."

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Port forwarding agent-launcher service to localhost:8080${NC}"
echo -e "${GREEN}API will be available at: http://localhost:8080${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop port forwarding${NC}\n"

kubectl port-forward -n agentlauncher service/agent-launcher 8080:8080