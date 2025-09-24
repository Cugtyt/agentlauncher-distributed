#!/bin/bash
set -e

echo "Deploying to Kubernetes..."

# Read version from build script
VERSION=${1:-latest}
if [ -f .image-version ]; then
    VERSION=$(cat .image-version)
fi
echo "Using image version: $VERSION"

# Create namespace
kubectl apply -f deployments/k8s/00-namespace.yaml

# Deploy ConfigMap
echo "Deploying configuration..."
kubectl apply -f deployments/k8s/01-configmap.yaml

# Deploy infrastructure (NATS, Redis)
echo "Deploying infrastructure..."
kubectl apply -f deployments/k8s/infrastructure/

# Wait for infrastructure to be ready
echo "Waiting for infrastructure to be ready..."
kubectl wait --for=condition=ready pod -l app=nats -n agentlauncher --timeout=60s
kubectl wait --for=condition=ready pod -l app=redis -n agentlauncher --timeout=60s

# Deploy services with version
echo "Deploying services with version: $VERSION..."

# Update image versions in deployment files
for service in agent-launcher agent-runtime llm-runtime tool-runtime; do
    if [ -f "deployments/k8s/services/${service}-deployment.yaml" ]; then
        # Create a temporary file with updated image version
        sed "s|{{VERSION}}|${VERSION}|g" \
            deployments/k8s/services/${service}-deployment.yaml > /tmp/${service}-deployment.yaml
        kubectl apply -f /tmp/${service}-deployment.yaml
        rm -f /tmp/${service}-deployment.yaml
    fi
done

# Apply only service definitions (not deployments)
for service_file in deployments/k8s/services/*-service.yaml; do
    if [ -f "$service_file" ]; then
        kubectl apply -f "$service_file"
    fi
done

# Deploy HPA
echo "Setting up auto-scaling..."
kubectl apply -f deployments/k8s/hpa/

echo "Deployment complete!"
kubectl get all -n agentlauncher