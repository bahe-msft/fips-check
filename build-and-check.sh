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
echo "Checking if runtime image has OpenSSL binary"
echo "==================================================================="

# Pull the runtime image to inspect it
docker pull "$RUNTIME_IMAGE" >/dev/null 2>&1 || true

# Check if the image has OpenSSL binary by trying to execute it
HAS_OPENSSL=false

# Try common OpenSSL binary locations
for openssl_path in /usr/bin/openssl /bin/openssl /usr/local/bin/openssl openssl; do
    if docker run --rm --entrypoint sh "$RUNTIME_IMAGE" -c "command -v $openssl_path" >/dev/null 2>&1; then
        HAS_OPENSSL=true
        echo "✓ Found OpenSSL binary at: $openssl_path"
        break
    fi
done

# If no shell, try direct execution
if [ "$HAS_OPENSSL" = false ]; then
    if docker run --rm --entrypoint /usr/bin/openssl "$RUNTIME_IMAGE" version >/dev/null 2>&1; then
        HAS_OPENSSL=true
        echo "✓ Found OpenSSL binary at: /usr/bin/openssl"
    elif docker run --rm --entrypoint /bin/openssl "$RUNTIME_IMAGE" version >/dev/null 2>&1; then
        HAS_OPENSSL=true
        echo "✓ Found OpenSSL binary at: /bin/openssl"
    fi
fi

echo ""

if [ "$HAS_OPENSSL" = false ]; then
    echo "==================================================================="
    echo "FIPS Compliance Check Result: NON-COMPLIANT"
    echo "==================================================================="
    echo "Reason: Runtime image does not contain OpenSSL binary"
    echo ""
    echo "Images without OpenSSL cannot support FIPS compliance because:"
    echo "  - FIPS mode requires OpenSSL with FIPS module"
    echo "  - Go systemcrypto depends on OpenSSL for cryptographic operations"
    echo "  - No OpenSSL means no FIPS cryptographic provider available"
    echo ""
    echo "This is common in:"
    echo "  - Distroless images"
    echo "  - Minimal/scratch-based images"
    echo "  - Images using alternative crypto libraries"
    echo "==================================================================="
    exit 1
fi

echo "✓ Runtime image has OpenSSL binary, proceeding with FIPS check"
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
