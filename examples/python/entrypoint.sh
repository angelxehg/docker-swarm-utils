#!/bin/sh
set -e

# List contents of /run/secrets
ls /run/secrets

# Run the swarm utils binary
docker-swarm-utils

# Execute the passed command
exec "$@"
