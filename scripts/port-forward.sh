#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Setting up port forwarding for agentlauncher services...${NC}"

# Check if services are running
echo -e "${YELLOW}Checking service status...${NC}"
kubectl get pods -n agentlauncher

# Wait for agent-launcher to be ready
echo -e "${YELLOW}Waiting for agent-launcher to be ready...${NC}"
kubectl wait --for=condition=ready pod -l app=agent-launcher -n agentlauncher --timeout=60s

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ agent-launcher is ready${NC}"
else
    echo -e "${RED}✗ agent-launcher failed to become ready${NC}"
    exit 1
fi

# Start port forwarding
echo -e "${YELLOW}Starting port forwarding...${NC}"
echo -e "${GREEN}Agent Launcher API will be available at: http://localhost:8080${NC}"
echo -e "${GREEN}Tool Runtime API will be available at: http://localhost:8082${NC}"
echo -e "${GREEN}NATS Monitoring will be available at: http://localhost:8222${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop port forwarding${NC}"

# Port forward all services
kubectl port-forward -n agentlauncher service/agent-launcher 8080:8080 &
kubectl port-forward -n agentlauncher service/tool-runtime 8082:8082 &
kubectl port-forward -n agentlauncher nats-0 8222:8222 &

# Keep the script running and handle cleanup
trap 'echo -e "\n${YELLOW}Stopping port forwarding...${NC}"; kill $(jobs -p); exit 0' INT

echo -e "${GREEN}Port forwarding is active. Services are accessible at:${NC}"
echo -e "  • Agent Launcher: http://localhost:8080"
echo -e "  • Tool Runtime: http://localhost:8082"
echo -e "  • NATS Monitoring: http://localhost:8222"
echo -e "  • Health check: curl http://localhost:8080/health"
echo -e "  • NATS info: curl http://localhost:8222/varz"

# Wait for user to stop
wait