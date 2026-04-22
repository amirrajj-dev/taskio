# Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o taskio ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary and docs
COPY --from=builder /app/taskio .
COPY --from=builder /app/docs ./docs

EXPOSE 3000

CMD ["./taskio"]