#!/bin/sh
set -e

# Run the swarm utils binary
docker-swarm-utils

# Execute the passed command
exec "$@"
