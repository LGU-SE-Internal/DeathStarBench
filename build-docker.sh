#!/bin/bash
# Build script for DeathStarBench with OpenTelemetry support

set -e  # Exit on error

# Default registry
REGISTRY="${DOCKER_REGISTRY:-10.10.10.240/library}"
TAG="${DOCKER_TAG:-otel}"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to build an image
build_image() {
    local context=$1
    local image_name=$2
    local dockerfile=${3:-Dockerfile}
    
    log_info "Building $image_name from $context..."
    
    local full_image_name="$REGISTRY/$image_name:$TAG"
    
    if docker build -t "$full_image_name" -f "$context/$dockerfile" "$context"; then
        log_info "Successfully built $full_image_name"
        return 0
    else
        log_error "Failed to build $full_image_name"
        return 1
    fi
}

# Parse command line arguments
BUILD_SOCIAL=true
BUILD_MEDIA=true

while [[ $# -gt 0 ]]; do
    case $1 in
        --social-only)
            BUILD_MEDIA=false
            shift
            ;;
        --media-only)
            BUILD_SOCIAL=false
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --social-only    Build only social network service"
            echo "  --media-only     Build only media microservices"
            echo "  --help, -h       Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  DOCKER_REGISTRY  Docker registry prefix (default: 10.10.10.240/library)"
            echo "  DOCKER_TAG       Tag for built images (default: otel)"
            echo ""
            echo "Examples:"
            echo "  $0                                              # Build everything"
            echo "  $0 --social-only                                # Build only social network"
            echo "  DOCKER_TAG=v1.0 $0                              # Build with custom tag"
            echo "  DOCKER_REGISTRY=myregistry.com/myproject $0     # Use different registry"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

log_info "Starting build process..."
log_info "Registry: $REGISTRY"
log_info "Tag: $TAG"

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Build social network services
if [ "$BUILD_SOCIAL" = true ]; then
    log_info "========================================="
    log_info "Building Social Network Services"
    log_info "========================================="
    log_warn "This may take 30-60 minutes on first build (compiling dependencies from source)..."
    
    build_image "socialNetwork" "social-network-microservices" || exit 1
    
    log_info "Social network services built successfully!"
else
    log_warn "Skipping social network services build"
fi

# Build media microservices
if [ "$BUILD_MEDIA" = true ]; then
    log_info "========================================="
    log_info "Building Media Microservices"
    log_info "========================================="
    log_warn "This may take 30-60 minutes on first build (compiling dependencies from source)..."
    
    build_image "mediaMicroservices" "media-microservices" || exit 1
    
    log_info "Media microservices built successfully!"
else
    log_warn "Skipping media microservices build"
fi

log_info "========================================="
log_info "Build Complete!"
log_info "========================================="

echo ""
log_info "Built images:"
if [ "$BUILD_SOCIAL" = true ]; then
    echo "  - $REGISTRY/social-network-microservices:$TAG"
fi
if [ "$BUILD_MEDIA" = true ]; then
    echo "  - $REGISTRY/media-microservices:$TAG"
fi

echo ""
log_info "To push images to registry:"
if [ "$BUILD_SOCIAL" = true ]; then
    echo "  docker push $REGISTRY/social-network-microservices:$TAG"
fi
if [ "$BUILD_MEDIA" = true ]; then
    echo "  docker push $REGISTRY/media-microservices:$TAG"
fi

echo ""
log_info "To test the images:"
echo "  docker run --rm $REGISTRY/social-network-microservices:$TAG ls -la /usr/local/lib/ | grep opentelemetry"
