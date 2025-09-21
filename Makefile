# Variables
DOCKER_REGISTRY := agentlauncher
NAMESPACE := agentlauncher
SERVICES := agent-launcher agent-runtime llm-runtime tool-runtime message-runtime

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

.PHONY: all build build-images deploy clean logs help

# Default target
all: build-images deploy

# Help target
help:
    @echo "$(GREEN)Available targets:$(NC)"
    @echo "  $(YELLOW)make build$(NC)          - Build Go binaries locally"
    @echo "  $(YELLOW)make build-images$(NC)   - Build Docker images"
    @echo "  $(YELLOW)make deploy$(NC)         - Deploy to Kubernetes"
    @echo "  $(YELLOW)make clean$(NC)          - Clean up Kubernetes resources"
    @echo "  $(YELLOW)make logs$(NC)           - View logs from all services"
    @echo "  $(YELLOW)make port-forward$(NC)   - Port forward API to localhost:8080"
    @echo "  $(YELLOW)make test-client$(NC)    - Run example client"
    @echo "  $(YELLOW)make status$(NC)         - Check deployment status"
    @echo "  $(YELLOW)make redeploy$(NC)       - Rebuild and redeploy everything"
    @echo "  $(YELLOW)make setup-secrets$(NC)  - Set up OpenAI API key"

# Build Go binaries locally
build:
    @echo "$(YELLOW)Building Go binaries...$(NC)"
    @mkdir -p bin
    @for service in $(SERVICES); do \
        echo "Building $$service..."; \
        go build -o bin/$$service ./cmd/$$service || exit 1; \
    done
    @echo "$(GREEN)✓ All binaries built successfully$(NC)"

# Build Docker images
build-images:
    @echo "$(YELLOW)Building Docker images...$(NC)"
    @for service in $(SERVICES); do \
        echo "Building $(DOCKER_REGISTRY)/$$service:latest..."; \
        docker build -f deployments/docker/Dockerfile.$$service -t $(DOCKER_REGISTRY)/$$service:latest . || exit 1; \
        echo "$(GREEN)✓ $$service image built$(NC)"; \
    done
    @echo "$(GREEN)✓ All images built successfully$(NC)"
    @docker images | grep $(DOCKER_REGISTRY)

# Deploy to Kubernetes
deploy: check-kubectl
    @echo "$(YELLOW)Deploying to Kubernetes...$(NC)"
    @echo "Creating namespace..."
    @kubectl apply -f deployments/k8s/00-namespace.yaml
    @echo "Applying configurations..."
    @kubectl apply -f deployments/k8s/01-configmap.yaml
    @kubectl apply -f deployments/k8s/02-secrets.yaml
    @echo "$(YELLOW)Deploying infrastructure...$(NC)"
    @kubectl apply -f deployments/k8s/infrastructure/
    @echo "Waiting for infrastructure to be ready..."
    @kubectl wait --for=condition=ready pod -l app=nats -n $(NAMESPACE) --timeout=60s 2>/dev/null || true
    @kubectl wait --for=condition=ready pod -l app=redis -n $(NAMESPACE) --timeout=60s 2>/dev/null || true
    @sleep 5
    @echo "$(YELLOW)Deploying services...$(NC)"
    @kubectl apply -f deployments/k8s/services/
    @echo "$(YELLOW)Setting up auto-scaling...$(NC)"
    @kubectl apply -f deployments/k8s/hpa/
    @echo "$(GREEN)✓ Deployment complete!$(NC)"
    @$(MAKE) wait-ready

# Wait for all deployments to be ready
wait-ready:
    @echo "$(YELLOW)Waiting for all deployments to be ready...$(NC)"
    @for service in $(SERVICES); do \
        kubectl rollout status deployment/$$service -n $(NAMESPACE) --timeout=120s || true; \
    done
    @echo "$(GREEN)✓ All services are ready$(NC)"

# Clean up Kubernetes resources
clean: check-kubectl
    @echo "$(YELLOW)Cleaning up Kubernetes resources...$(NC)"
    @kubectl delete namespace $(NAMESPACE) --ignore-not-found=true
    @echo "$(GREEN)✓ Cleanup complete$(NC)"

# View logs from all services
logs: check-kubectl
    @echo "$(YELLOW)Viewing logs (Ctrl+C to exit)...$(NC)"
    @kubectl logs -n $(NAMESPACE) -l app=agent-runtime --tail=100 -f

# View logs for specific service
logs-%: check-kubectl
    @echo "$(YELLOW)Viewing logs for $* (Ctrl+C to exit)...$(NC)"
    @kubectl logs -n $(NAMESPACE) -l app=$* --tail=100 -f

# Port forward for local access
port-forward: check-kubectl
    @echo "$(GREEN)Port forwarding API to http://localhost:8080$(NC)"
    @echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
    @kubectl port-forward -n $(NAMESPACE) service/agent-launcher 8080:8080

# Port forward for monitoring
port-forward-nats: check-kubectl
    @echo "$(GREEN)Port forwarding NATS monitor to http://localhost:8222$(NC)"
    @kubectl port-forward -n $(NAMESPACE) service/nats 8222:8222

port-forward-redis: check-kubectl
    @echo "$(GREEN)Port forwarding Redis to localhost:6379$(NC)"
    @kubectl port-forward -n $(NAMESPACE) service/redis 6379:6379

# Run example client
test-client:
    @echo "$(YELLOW)Running example client...$(NC)"
    @go run examples/client/main.go

# Check status
status: check-kubectl
    @echo "$(YELLOW)Deployment status:$(NC)"
    @kubectl get all -n $(NAMESPACE)
    @echo "\n$(YELLOW)Pod details:$(NC)"
    @kubectl get pods -n $(NAMESPACE) -o wide
    @echo "\n$(YELLOW)HPA status:$(NC)"
    @kubectl get hpa -n $(NAMESPACE)

# Quick rebuild and redeploy
redeploy: build-images deploy
    @echo "$(GREEN)✓ Redeploy complete$(NC)"

# Setup secrets interactively
setup-secrets: check-kubectl
    @echo "$(YELLOW)Setting up secrets...$(NC)"
    @read -p "Enter your OpenAI API Key: " api_key; \
    kubectl create secret generic agentlauncher-secrets \
        --from-literal=OPENAI_API_KEY="$$api_key" \
        --namespace=$(NAMESPACE) \
        --dry-run=client -o yaml | kubectl apply -f -
    @echo "$(GREEN)✓ Secrets configured$(NC)"

# Development mode - build and run locally
dev: build
    @echo "$(YELLOW)Starting services locally (requires NATS and Redis)...$(NC)"
    @echo "$(RED)Make sure NATS and Redis are running locally$(NC)"
    @trap 'kill %1 %2 %3 %4 %5' SIGINT; \
    ./bin/agent-launcher & \
    ./bin/agent-runtime & \
    ./bin/llm-runtime & \
    ./bin/tool-runtime & \
    ./bin/message-runtime & \
    wait

# Run tests
test:
    @echo "$(YELLOW)Running tests...$(NC)"
    @go test -v ./...

# Format code
fmt:
    @echo "$(YELLOW)Formatting code...$(NC)"
    @go fmt ./...
    @echo "$(GREEN)✓ Code formatted$(NC)"

# Lint code
lint:
    @echo "$(YELLOW)Linting code...$(NC)"
    @golangci-lint run ./...

# Generate Go modules
mod:
    @echo "$(YELLOW)Updating Go modules...$(NC)"
    @go mod tidy
    @go mod vendor
    @echo "$(GREEN)✓ Modules updated$(NC)"

# Check if kubectl is available
check-kubectl:
    @which kubectl > /dev/null || (echo "$(RED)kubectl not found. Please install kubectl.$(NC)" && exit 1)

# Check if cluster is reachable
check-cluster: check-kubectl
    @kubectl cluster-info > /dev/null 2>&1 || (echo "$(RED)Cannot connect to Kubernetes cluster$(NC)" && exit 1)

# Scale services
scale-%: check-kubectl
    @read -p "Enter number of replicas for $*: " replicas; \
    kubectl scale deployment/$* --replicas=$$replicas -n $(NAMESPACE)
    @echo "$(GREEN)✓ Scaled $* to $$replicas replicas$(NC)"

# Show resource usage
resources: check-kubectl
    @echo "$(YELLOW)Resource usage:$(NC)"
    @kubectl top pods -n $(NAMESPACE)
    @echo "\n$(YELLOW)Node resource usage:$(NC)"
    @kubectl top nodes

# Tail events
events: check-kubectl
    @echo "$(YELLOW)Kubernetes events:$(NC)"
    @kubectl get events -n $(NAMESPACE) --sort-by='.lastTimestamp'

# Watch pods
watch: check-kubectl
    @watch -n 2 kubectl get pods -n $(NAMESPACE)

# Debug a specific pod
debug-%: check-kubectl
    @pod=$$(kubectl get pods -n $(NAMESPACE) -l app=$* -o jsonpath='{.items[0].metadata.name}'); \
    echo "$(YELLOW)Debugging pod: $$pod$(NC)"; \
    kubectl describe pod $$pod -n $(NAMESPACE)

# Execute into a pod
exec-%: check-kubectl
    @pod=$$(kubectl get pods -n $(NAMESPACE) -l app=$* -o jsonpath='{.items[0].metadata.name}'); \
    echo "$(YELLOW)Executing into pod: $$pod$(NC)"; \
    kubectl exec -it $$pod -n $(NAMESPACE) -- /bin/sh

# Quick start for development
quickstart: build-images deploy wait-ready port-forward

# Install development tools
install-tools:
    @echo "$(YELLOW)Installing development tools...$(NC)"
    @go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    @echo "$(GREEN)✓ Development tools installed$(NC)"

# Clean Docker images
clean-images:
    @echo "$(YELLOW)Cleaning Docker images...$(NC)"
    @for service in $(SERVICES); do \
        docker rmi $(DOCKER_REGISTRY)/$$service:latest 2>/dev/null || true; \
    done
    @echo "$(GREEN)✓ Images cleaned$(NC)"

# Full clean (Kubernetes + Docker)
clean-all: clean clean-images
    @rm -rf bin/
    @echo "$(GREEN)✓ Full cleanup complete$(NC)"

# Show help by default
.DEFAULT_GOAL := help