# 🔹 Stage 1: Build Golang Application
FROM golang:1.22 AS builder
WORKDIR /app

# Force GOPATH AND GOBIN
ENV GOPATH=/go
ENV GOBIN=/go/bin

# Set Go module proxy (remove if unnecessary)
ENV GOPROXY=https://proxy.golang.org,direct

# Copy go.mod & go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# Copy application source code
COPY . .

# Copy migrations folder
COPY db/migrations ./db/migrations

# Build API service binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server-api cmd/server/server.go

# Build Worker service binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o kafka-worker cmd/kafka/worker.go

# Install migrate tool with MySQL driver
RUN go install -tags 'mysql' -ldflags="-s -w" github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Stage 2: Minimal Deployment Environment
FROM debian:bookworm-slim
WORKDIR /app

# Set timezone (optional)
RUN apt-get update && apt-get install -y tzdata && rm -rf /var/lib/apt/lists/*
ENV TZ=Asia/Taipei

# Copy API service binary
COPY --from=builder /app/server-api /app/server-api

# Copy Worker service binary
COPY --from=builder /app/kafka-worker /app/kafka-worker

# Copy the 'migrate' binary
COPY --from=builder /go/bin/migrate /app/migrate

# Copy migrations to final image
COPY --from=builder /app/db/migrations /app/db/migrations

# Grant execution permission
RUN chmod +x /app/server-api /app/kafka-worker /app/migrate

# Define exposed ports
EXPOSE 8080

# Default startup command
CMD ["/app/server-api"]
