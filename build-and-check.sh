#!/bin/bash

set -e

# Check if runtime image argument is provided
if [ -z "$1" ]; then
    echo "Error: RUNTIME_IMAGE argument is required"
    echo "Usage: $0 <runtime-image>"
    exit 1
fi

RUNTIME_IMAGE="$1"

# Phase 1: Detect the build image based on runtime image
detect_build_image() {
    local runtime_image="$1"
    
    # TODO: Implement logic to detect build image based on runtime image
    # For now, return the default build image
    echo "mcr.microsoft.com/oss/go/microsoft/golang:1.24-fips-azurelinux3.0"
}

BUILD_IMAGE=$(detect_build_image "$RUNTIME_IMAGE")

echo "==================================================================="
echo "Phase 1: Detected build image"
echo "==================================================================="
echo "BUILD_IMAGE: $BUILD_IMAGE"
echo "RUNTIME_IMAGE: $RUNTIME_IMAGE"
echo ""

# Check if runtime image is distroless
echo "==================================================================="
echo "Checking if runtime image is distroless-based"
echo "==================================================================="

# Pull the runtime image to inspect it
docker pull "$RUNTIME_IMAGE" >/dev/null 2>&1 || true

# Check if the image is distroless by inspecting common distroless characteristics:
# 1. No shell (/bin/sh, /bin/bash)
# 2. Very minimal file system
# 3. Image labels or repository name containing "distroless"

IS_DISTROLESS=false

# Check image config for distroless indicators
IMAGE_INFO=$(docker image inspect "$RUNTIME_IMAGE" 2>/dev/null || echo "[]")

# Check if image name contains "distroless"
if echo "$RUNTIME_IMAGE" | grep -iq "distroless"; then
    IS_DISTROLESS=true
    echo "✗ Image name contains 'distroless'"
fi

# Try to run a shell command to detect if shell exists
if [ "$IS_DISTROLESS" = false ]; then
    if ! docker run --rm --entrypoint /bin/sh "$RUNTIME_IMAGE" -c "echo test" >/dev/null 2>&1; then
        if ! docker run --rm --entrypoint /bin/bash "$RUNTIME_IMAGE" -c "echo test" >/dev/null 2>&1; then
            IS_DISTROLESS=true
            echo "✗ No shell found in image (likely distroless)"
        fi
    fi
fi

# Check for very small image size (distroless images are typically very small)
IMAGE_SIZE=$(echo "$IMAGE_INFO" | jq -r '.[0].Size // 0' 2>/dev/null || echo "0")
if [ "$IS_DISTROLESS" = false ] && [ "$IMAGE_SIZE" -gt 0 ] && [ "$IMAGE_SIZE" -lt 20000000 ]; then
    # Less than 20MB is a strong indicator of distroless
    echo "ℹ Image size is very small ($IMAGE_SIZE bytes), possible distroless indicator"
fi

echo ""

if [ "$IS_DISTROLESS" = true ]; then
    echo "==================================================================="
    echo "FIPS Compliance Check Result: NON-COMPLIANT"
    echo "==================================================================="
    echo "Reason: Runtime image is distroless-based"
    echo ""
    echo "Distroless images do not contain the necessary tools and libraries"
    echo "required for FIPS compliance verification. They lack:"
    echo "  - Shell utilities needed for runtime checks"
    echo "  - Standard system libraries"
    echo "  - FIPS-enabled OpenSSL or cryptographic libraries"
    echo "==================================================================="
    exit 1
fi

echo "✓ Runtime image is not distroless-based, proceeding with FIPS check"
echo ""

# Phase 2: Build the Docker image
IMAGE_TAG="fips-checker:${RUNTIME_IMAGE//\//-}"
IMAGE_TAG="${IMAGE_TAG//:/-}"

echo "==================================================================="
echo "Phase 2: Building Docker image"
echo "==================================================================="
echo "Image tag: $IMAGE_TAG"
echo ""

docker build \
    --build-arg BUILD_IMAGE="$BUILD_IMAGE" \
    --build-arg RUNTIME_IMAGE="$RUNTIME_IMAGE" \
    -t "$IMAGE_TAG" \
    .

echo ""
echo "==================================================================="
echo "Phase 3: Running the built image"
echo "==================================================================="
echo ""

# Phase 3: Run the built image
docker run --rm "$IMAGE_TAG"
