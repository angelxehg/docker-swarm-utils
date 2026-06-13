# docker-swarm-utils

A lightweight CLI tool designed to assist with Docker Secrets and Configs in Docker Swarm environments.

## Usage

Build a image:

```bash
docker build -f examples/python/Dockerfile -t swarm-utils:app .
```

Run image:

```shell
docker run --rm \
  -v "$(pwd)/examples/python/simple_value.txt:/run/secrets/simple_value.txt:ro" \
  -v "$(pwd)/examples/python/environment_values.txt:/run/secrets/environment_values.txt:ro" \
  swarm-utils:app
```

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

## Multi-Architecture Support

The `Dockerfile` is configured to support cross-architecture builds using Docker Buildx.

```bash
docker buildx build --platform linux/amd64,linux/arm64 --target slim -t your-repo/swarm-utils:latest .
```
