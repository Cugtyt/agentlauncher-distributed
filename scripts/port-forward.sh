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

# Check if we have any running agent-launcher pods
RUNNING_PODS=$(kubectl get pods -n agentlauncher -l app=agent-launcher --field-selector=status.phase=Running -o name 2>/dev/null | wc -l)

if [ $RUNNING_PODS -gt 0 ]; then
    echo -e "${GREEN}✓ Found running agent-launcher pod${NC}"
else
    echo -e "${YELLOW}Waiting for agent-launcher to be ready...${NC}"
    kubectl wait --for=condition=ready pod -l app=agent-launcher -n agentlauncher --timeout=60s
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}✗ agent-launcher failed to become ready${NC}"
        echo -e "${YELLOW}Checking pod status:${NC}"
        kubectl get pods -n agentlauncher -l app=agent-launcher
        kubectl describe pods -n agentlauncher -l app=agent-launcher | tail -10
        echo -e "${YELLOW}Continuing with port forwarding anyway...${NC}"
    else
        echo -e "${GREEN}✓ agent-launcher is ready${NC}"
    fi
fi

# Start port forwarding
echo -e "${YELLOW}Starting port forwarding...${NC}"
echo -e "${GREEN}Agent Launcher API will be available at: http://localhost:8080${NC}"
echo -e "${GREEN}Tool Runtime API will be available at: http://localhost:8082${NC}"
echo -e "${GREEN}NATS Monitoring will be available at: http://localhost:8222${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop port forwarding${NC}"

# Port forward all services - use specific pods if services aren't working
echo -e "${YELLOW}Starting port forwarding...${NC}"

# Try service first, fall back to pod if needed
if kubectl get service -n agentlauncher agent-launcher >/dev/null 2>&1; then
    kubectl port-forward -n agentlauncher service/agent-launcher 8080:8080 &
else
    AGENT_POD=$(kubectl get pods -n agentlauncher -l app=agent-launcher --field-selector=status.phase=Running -o name | head -1)
    if [ -n "$AGENT_POD" ]; then
        kubectl port-forward -n agentlauncher $AGENT_POD 8080:8080 &
    fi
fi

if kubectl get service -n agentlauncher tool-runtime >/dev/null 2>&1; then
    kubectl port-forward -n agentlauncher service/tool-runtime 8082:8082 &
else
    TOOL_POD=$(kubectl get pods -n agentlauncher -l app=tool-runtime --field-selector=status.phase=Running -o name | head -1)
    if [ -n "$TOOL_POD" ]; then
        kubectl port-forward -n agentlauncher $TOOL_POD 8082:8082 &
    fi
fi

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