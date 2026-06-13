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

# Python image
FROM python:3.14-alpine AS app

# Copy app
WORKDIR /app
COPY examples/python/app.py .

# Setup entrypoint
COPY --from=builder /docker-swarm-utils /usr/local/bin/docker-swarm-utils
ENTRYPOINT ["/usr/local/bin/docker-swarm-utils", "entrypoint"]

# Setup command
CMD ["python", "app.py"]
