#!/bin/bash
# Build script for DeathStarBench with OpenTelemetry support
# This script builds the dependency images and service images in the correct order

set -e  # Exit on error

REGISTRY="${DOCKER_REGISTRY:-}"
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
    
    if [ -n "$REGISTRY" ]; then
        image_name="$REGISTRY/$image_name"
    fi
    
    if docker build -t "$image_name:$TAG" -f "$context/$dockerfile" "$context"; then
        log_info "Successfully built $image_name:$TAG"
        return 0
    else
        log_error "Failed to build $image_name:$TAG"
        return 1
    fi
}

# Parse command line arguments
BUILD_DEPS=true
BUILD_SOCIAL=true
BUILD_MEDIA=true

while [[ $# -gt 0 ]]; do
    case $1 in
        --deps-only)
            BUILD_SOCIAL=false
            BUILD_MEDIA=false
            shift
            ;;
        --social-only)
            BUILD_DEPS=false
            BUILD_MEDIA=false
            shift
            ;;
        --media-only)
            BUILD_DEPS=false
            BUILD_SOCIAL=false
            shift
            ;;
        --skip-deps)
            BUILD_DEPS=false
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --deps-only      Build only dependency images"
            echo "  --social-only    Build only social network service (skip deps and media)"
            echo "  --media-only     Build only media microservices (skip deps and social)"
            echo "  --skip-deps      Skip building dependency images"
            echo "  --help, -h       Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  DOCKER_REGISTRY  Docker registry prefix (optional)"
            echo "  DOCKER_TAG       Tag for built images (default: otel)"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Build everything"
            echo "  $0 --deps-only                        # Build only dependencies"
            echo "  $0 --skip-deps                        # Build services assuming deps exist"
            echo "  DOCKER_REGISTRY=10.10.10.240/library $0  # Build with registry prefix"
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
log_info "Registry: ${REGISTRY:-<none>}"
log_info "Tag: $TAG"

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Build dependency images
if [ "$BUILD_DEPS" = true ]; then
    log_info "========================================="
    log_info "Building Dependency Images"
    log_info "========================================="
    log_warn "This step may take 30-60 minutes on first build..."
    
    # Build social network dependencies
    build_image "socialNetwork/docker/thrift-microservice-deps/cpp" \
                "deathstarbench/social-network-microservices-deps" || exit 1
    
    # Build media microservices dependencies  
    build_image "mediaMicroservices/docker/thrift-microservice-deps/cpp" \
                "deathstarbench/media-microservices-deps" || exit 1
    
    log_info "Dependency images built successfully!"
else
    log_warn "Skipping dependency image builds"
fi

# Build social network services
if [ "$BUILD_SOCIAL" = true ]; then
    log_info "========================================="
    log_info "Building Social Network Services"
    log_info "========================================="
    
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
if [ "$BUILD_DEPS" = true ]; then
    echo "  - ${REGISTRY:+$REGISTRY/}deathstarbench/social-network-microservices-deps:$TAG"
    echo "  - ${REGISTRY:+$REGISTRY/}deathstarbench/media-microservices-deps:$TAG"
fi
if [ "$BUILD_SOCIAL" = true ]; then
    echo "  - ${REGISTRY:+$REGISTRY/}social-network-microservices:$TAG"
fi
if [ "$BUILD_MEDIA" = true ]; then
    echo "  - ${REGISTRY:+$REGISTRY/}media-microservices:$TAG"
fi

echo ""
log_info "To push images to registry:"
if [ -n "$REGISTRY" ]; then
    if [ "$BUILD_SOCIAL" = true ]; then
        echo "  docker push $REGISTRY/social-network-microservices:$TAG"
    fi
    if [ "$BUILD_MEDIA" = true ]; then
        echo "  docker push $REGISTRY/media-microservices:$TAG"
    fi
else
    echo "  Set DOCKER_REGISTRY environment variable and rebuild"
fi
