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
