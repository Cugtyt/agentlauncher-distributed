#!/bin/bash
set -e

echo "Deploying to Kubernetes..."

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

# Deploy services
echo "Deploying services..."
kubectl apply -f deployments/k8s/services/

# Deploy HPA
echo "Setting up auto-scaling..."
kubectl apply -f deployments/k8s/hpa/

echo "Deployment complete!"
kubectl get all -n agentlauncher