# Dockerfile for gemini-cli-go
# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o gemini-cli-go .

# Final stage
FROM scratch
WORKDIR /

# Copy binary from builder
COPY --from=builder /app/gemini-cli-go /gemini-cli-go

# Entrypoint
ENTRYPOINT ["/gemini-cli-go"]
