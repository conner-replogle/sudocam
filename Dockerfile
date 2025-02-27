# Use Go 1.23 bookworm as base image
FROM golang:1.24-bookworm AS base

FROM node:20-slim AS ui-builder
WORKDIR /build
COPY . .
RUN npm install -g pnpm
RUN cd server/ui && pnpm install --force
RUN cd server/ui && pnpm run build

FROM base AS builder
# Move to working directory /build
WORKDIR /build

# Install SQLite development dependencies
RUN apt-get update && apt-get install -y libsqlite3-dev

# Copy the entire source code into the container
COPY . .

# Build the application
RUN cd server &&  go mod download

# Build the application with static linking
RUN cd server && go build  -o sudocam-server

# Use stable-slim instead of bookworm-slim
FROM debian:stable-slim AS production

# Combine CA certificate installation and SQLite3 runtime into a single layer
RUN apt-get update && \
    apt-get install -y ca-certificates libsqlite3-0 && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

    
WORKDIR /app

# Create data directory for SQLite database
RUN mkdir -p /app/data

COPY --from=builder /build/server/sudocam-server ./
COPY --from=ui-builder /build/server/ui/dist ./ui/dist

# Document the port that may need to be published
EXPOSE 8080

# Define a volume for SQLite database persistence
VOLUME ["/app/data"]

# Start the application
CMD ["/app/sudocam-server"]
