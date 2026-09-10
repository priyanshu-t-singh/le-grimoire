# Stage 1: Build
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

WORKDIR /app

# Cache Go modules
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code and build
COPY . .

ARG TARGETOS TARGETARCH

# Set default version to "docker-dev" if not provided
ARG VERSION=docker-dev

# Build the binary with CGO disabled for cross-compilation
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build \
      -trimpath \
      -ldflags="-w -s -X 'le-grimoire/internal/constants.Version=${VERSION}'" \
      -o le-grimoire main.go

# Stage 2: Runtime image
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata su-exec && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

RUN mkdir -p /home/appuser/.config/

# Copy compiled binary from the builder stage to system PATH
COPY --from=builder /app/le-grimoire /usr/local/bin/le-grimoire
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENV LE_GRIMOIRE_SERVER_HOST=0.0.0.0
ENV LE_GRIMOIRE_SERVER_PORT=8321

EXPOSE 8321

# Stay root here — entrypoint drops privileges
ENTRYPOINT ["/entrypoint.sh"]
