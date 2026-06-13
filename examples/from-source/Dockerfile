# Builder stage
FROM golang:1.26-alpine AS builder

# Install git to clone the repository
RUN apk add --no-cache git

WORKDIR /src

# Clone the specific tag from source
RUN git clone --branch v0.1.0 https://github.com/angelxehg/docker-swarm-utils.git .

# Build the binary
RUN CGO_ENABLED=0 go build -o /docker-swarm-utils .

# Final stage
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /docker-swarm-utils /usr/local/bin/docker-swarm-utils

# Use it as entrypoint
ENTRYPOINT ["/usr/local/bin/docker-swarm-utils", "entrypoint"]

# Your application command
CMD ["echo", "Docker Swarm Utils is ready!"]
