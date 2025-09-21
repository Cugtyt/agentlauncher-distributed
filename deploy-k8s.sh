#!/bin/bash

set -e

echo "Deploying agentlauncher-distributed to Kubernetes..."

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl not found. Please install kubectl.${NC}"
    exit 1
fi

# Create namespace
echo -e "${YELLOW}Creating namespace...${NC}"
kubectl apply -f deployments/k8s/00-namespace.yaml

# Apply configurations
echo -e "${YELLOW}Applying configurations...${NC}"
kubectl apply -f deployments/k8s/01-configmap.yaml

# Check for secrets
if [ -z "$OPENAI_API_KEY" ]; then
    echo -e "${RED}Warning: OPENAI_API_KEY environment variable not set${NC}"
    echo -e "${YELLOW}Please update deployments/k8s/02-secrets.yaml with your API key${NC}"
fi
kubectl apply -f deployments/k8s/02-secrets.yaml

# Deploy infrastructure
echo -e "${YELLOW}Deploying infrastructure (NATS, Redis)...${NC}"
kubectl apply -f deployments/k8s/infrastructure/

# Wait for infrastructure
echo -e "${YELLOW}Waiting for infrastructure to be ready...${NC}"
kubectl wait --for=condition=ready pod -l app=nats -n agentlauncher --timeout=60s || true
kubectl wait --for=condition=ready pod -l app=redis -n agentlauncher --timeout=60s || true

sleep 10

# Deploy services
echo -e "${YELLOW}Deploying services...${NC}"
kubectl apply -f deployments/k8s/services/

# Deploy HPA
echo -e "${YELLOW}Setting up auto-scaling...${NC}"
kubectl apply -f deployments/k8s/hpa/

# Wait for deployments
echo -e "${YELLOW}Waiting for deployments to be ready...${NC}"
kubectl rollout status deployment/agent-launcher -n agentlauncher --timeout=120s
kubectl rollout status deployment/agent-runtime -n agentlauncher --timeout=120s
kubectl rollout status deployment/llm-runtime -n agentlauncher --timeout=120s
kubectl rollout status deployment/tool-runtime -n agentlauncher --timeout=120s
kubectl rollout status deployment/message-runtime -n agentlauncher --timeout=120s

echo -e "${GREEN}✓ Deployment complete!${NC}"

# Show status
echo -e "\n${YELLOW}Deployment status:${NC}"
kubectl get all -n agentlauncher

echo -e "\n${YELLOW}To access the API, run:${NC}"
echo "kubectl port-forward -n agentlauncher service/agent-launcher 8080:8080"