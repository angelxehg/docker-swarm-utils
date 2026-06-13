# docker-swarm-utils

A lightweight CLI tool designed to assist with Docker Secrets and Configs in Docker Swarm environments.

## Local Development

### Prerequisites

- Go 1.21 or higher

### Build

To build the binary locally:

```bash
go build -o docker-swarm-utils .
```

### Run

```bash
./docker-swarm-utils
```

### Test

```bash
go test ./...
```

## Docker Usage

This project supports multi-stage builds and provides two final image targets: `slim` (Debian-based) and `alpine` (Alpine-based).

### Build Slim Image

```bash
docker build --target slim -t swarm-utils:slim .
```

### Run Slim Image

```bash
docker run --rm swarm-utils:slim
```

### Build Alpine Image

```bash
docker build --target alpine -t swarm-utils:alpine .
```

### Run Alpine Image

```bash
docker run --rm swarm-utils:alpine
```

## Docker Swarm Integration

The binary is designed to be used in an entrypoint script to load Docker Secrets or Configs before your main application starts.

Example `entrypoint.sh`:

```bash
#!/bin/sh
set -e

# Load secrets/configs using docker-swarm-utils (implementation pending)
docker-swarm-utils

# Start the main application
exec "$@"
```

## Multi-Architecture Support

The `Dockerfile` is configured to support cross-architecture builds using Docker Buildx.

```bash
docker buildx build --platform linux/amd64,linux/arm64 --target slim -t your-repo/swarm-utils:latest .
```
