# Builder stage
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Copy go.mod
COPY go.mod ./

# Download dependencies (if any)
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /docker-swarm-utils .

# Slim target (Debian-based)
FROM python:3.11-slim AS slim
COPY --from=builder /docker-swarm-utils /usr/local/bin/docker-swarm-utils
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["python3", "--version"]

# Alpine target (Alpine-based)
FROM python:3.11-alpine AS alpine
COPY --from=builder /docker-swarm-utils /usr/local/bin/docker-swarm-utils
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["python3", "--version"]
