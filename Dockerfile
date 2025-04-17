# Stage 1: application build
FROM golang:1.25 AS builder

# Set working directory inside the container
WORKDIR /app

# Copy Go mod files and download deps
COPY go.mod go.sum ./
RUN go mod tidy

# Copy all source files
COPY . ./

# Build static tfs binary
RUN CGO_ENABLED=0 go build -o /tfs .

# Stage 2: final image
FROM alpine:latest

# Copy binary and entrypoint script
COPY --from=builder /tfs /usr/local/bin/tfs
COPY entrypoint.sh /entrypoint.sh

# Ensure entrypoint is executable
RUN chmod +x /entrypoint.sh

RUN apk add --no-cache git

ENV PATH="/root/.local/bin:$PATH"

ENTRYPOINT ["/entrypoint.sh"]
