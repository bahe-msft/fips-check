# Define build arguments at the top level (before any FROM)
ARG BUILD_IMAGE=mcr.microsoft.com/oss/go/microsoft/golang:1.24-fips-azurelinux3.0
ARG RUNTIME_IMAGE=debian:bookworm-slim

# Build stage
FROM ${BUILD_IMAGE} AS builder

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY internal internal
COPY cmd cmd

# Build the binary
# CGO is enabled by default, but we set it explicitly since the code requires it
RUN CGO_ENABLED=1  go build -o fips-checker ./cmd/fips-checker

# Runtime stage
FROM ${RUNTIME_IMAGE}

# Copy the binary from builder stage
COPY --from=builder /build/fips-checker /usr/local/bin/fips-checker

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/fips-checker"]
