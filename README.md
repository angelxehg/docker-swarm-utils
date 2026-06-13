# docker-swarm-utils

A lightweight CLI tool designed to assist with Docker Secrets and Configs in Docker Swarm environments.

## Usage

### Build the Example Image

```bash
docker build -f examples/python/Dockerfile -t swarm-utils:app .
```

### Single-Value Secrets (Standard)

By default, the tool treats every file in `/run/secrets` or `/run/configs` as a single environment variable where the filename is the key and the content is the value.

```shell
docker run --rm \
  -v "$(pwd)/examples/python/simple_value.txt:/run/secrets/MY_SECRET:ro" \
  swarm-utils:app
```

### Dotenv Support (Multi-Variable)

If a secret file starts with (or contains) `# format: dotenv`, it will be parsed as a dotenv file, allowing multiple variables to be defined in a single file.

Example `secrets.env`:
```env
# format: dotenv
DB_USER=admin
DB_PASS=secret123
```

Variables from this file will be exported individually (`DB_USER` and `DB_PASS`).

Example run with dotenv file:
```shell
docker run --rm \
  -v "$(pwd)/examples/python/environment_values.txt:/run/secrets/vars.env:ro" \
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
