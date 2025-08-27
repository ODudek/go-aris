#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="aris"
APP_NAME="go-aris"
IMAGE_TAG="${IMAGE_TAG:-latest}"
REGISTRY="${REGISTRY:-localhost:5000}"
IMAGE_NAME="${REGISTRY}/${APP_NAME}:${IMAGE_TAG}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi

    # Check if docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "docker is not installed or not in PATH"
        exit 1
    fi

    # Check if we can connect to Kubernetes cluster
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    log_success "All prerequisites met"
}

build_image() {
    log_info "Building Docker image: ${IMAGE_NAME}"

    # Build the Docker image
    docker build -t "${IMAGE_NAME}" .

    if [ $? -eq 0 ]; then
        log_success "Docker image built successfully"
    else
        log_error "Failed to build Docker image"
        exit 1
    fi
}

push_image() {
    log_info "Pushing Docker image to registry: ${IMAGE_NAME}"

    # Push the image to registry
    docker push "${IMAGE_NAME}"

    if [ $? -eq 0 ]; then
        log_success "Docker image pushed successfully"
    else
        log_error "Failed to push Docker image"
        exit 1
    fi
}

deploy_to_kubernetes() {
    log_info "Deploying to Kubernetes cluster..."

    # Create namespace if it doesn't exist
    kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

    # Apply Kubernetes manifests
    log_info "Applying namespace..."
    kubectl apply -f k8s/namespace.yaml

    log_info "Applying ConfigMap..."
    kubectl apply -f k8s/configmap.yaml

    log_info "Applying Services..."
    kubectl apply -f k8s/service.yaml

    # Update the image in StatefulSet
    log_info "Updating StatefulSet with new image..."
    sed "s|image: go-aris:latest|image: ${IMAGE_NAME}|g" k8s/statefulset.yaml | kubectl apply -f -

    # Apply monitoring if the CRDs exist
    if kubectl get crd servicemonitors.monitoring.coreos.com &> /dev/null; then
        log_info "Applying monitoring configuration..."
        kubectl apply -f k8s/monitoring.yaml
    else
        log_warn "Prometheus CRDs not found, skipping monitoring setup"
    fi

    log_success "Kubernetes manifests applied successfully"
}

wait_for_deployment() {
    log_info "Waiting for deployment to be ready..."

    # Wait for StatefulSet to be ready
    kubectl wait --for=condition=ready pod -l app=${APP_NAME} -n ${NAMESPACE} --timeout=300s

    if [ $? -eq 0 ]; then
        log_success "Deployment is ready"
    else
        log_error "Deployment failed or timed out"
        show_logs
        exit 1
    fi
}

show_status() {
    log_info "Deployment status:"
    echo

    # Show pods
    kubectl get pods -n ${NAMESPACE} -l app=${APP_NAME}
    echo

    # Show services
    kubectl get services -n ${NAMESPACE}
    echo

    # Show StatefulSet
    kubectl get statefulset -n ${NAMESPACE}
    echo

    # Get external IP if LoadBalancer service exists
    EXTERNAL_IP=$(kubectl get service aris-loadbalancer -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    if [ -n "${EXTERNAL_IP}" ]; then
        log_success "External IP: ${EXTERNAL_IP}:6379"
    else
        log_info "Use port-forward to access the service:"
        log_info "kubectl port-forward service/aris-service -n ${NAMESPACE} 6379:8089"
    fi
}

show_logs() {
    log_info "Recent logs from Aris pods:"
    kubectl logs -l app=${APP_NAME} -n ${NAMESPACE} --tail=50
}

cleanup() {
    log_warn "Cleaning up deployment..."

    kubectl delete -f k8s/ --ignore-not-found=true
    kubectl delete namespace ${NAMESPACE} --ignore-not-found=true

    log_success "Cleanup completed"
}

scale_deployment() {
    local replicas=${1:-3}
    log_info "Scaling deployment to ${replicas} replicas..."

    kubectl scale statefulset aris-cluster -n ${NAMESPACE} --replicas=${replicas}
    kubectl wait --for=condition=ready pod -l app=${APP_NAME} -n ${NAMESPACE} --timeout=300s

    log_success "Scaled to ${replicas} replicas"
}

# Main execution
main() {
    case "${1:-deploy}" in
        "build")
            check_prerequisites
            build_image
            ;;
        "push")
            check_prerequisites
            push_image
            ;;
        "deploy")
            check_prerequisites
            build_image
            if [ "${REGISTRY}" != "localhost:5000" ]; then
                push_image
            fi
            deploy_to_kubernetes
            wait_for_deployment
            show_status
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs
            ;;
        "cleanup")
            cleanup
            ;;
        "scale")
            scale_deployment $2
            ;;
        "redeploy")
            log_info "Redeploying application..."
            kubectl rollout restart statefulset/aris-cluster -n ${NAMESPACE}
            wait_for_deployment
            show_status
            ;;
        *)
            echo "Usage: $0 {build|push|deploy|status|logs|cleanup|scale|redeploy}"
            echo
            echo "Commands:"
            echo "  build    - Build Docker image"
            echo "  push     - Push Docker image to registry"
            echo "  deploy   - Full deployment (build, push, deploy)"
            echo "  status   - Show deployment status"
            echo "  logs     - Show recent logs"
            echo "  cleanup  - Remove all resources"
            echo "  scale N  - Scale to N replicas"
            echo "  redeploy - Restart deployment"
            echo
            echo "Environment variables:"
            echo "  IMAGE_TAG - Docker image tag (default: latest)"
            echo "  REGISTRY  - Docker registry (default: localhost:5000)"
            exit 1
            ;;
    esac
}

# Execute main function with all arguments
main "$@"
