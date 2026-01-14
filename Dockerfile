FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
# We might need gcc for CGO if SQLite was used, but for Postgres usually pure Go is fine.
# adding git just in case.
RUN apk add --no-cache git

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the binary
# Disable CGO for static binary usually suitable for scratch/alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Final Stage
FROM alpine:latest

WORKDIR /app

# Install certificates for HTTPS (if needed external APIs)
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .

# Expose port (default gin is 8080)
EXPOSE 8080

CMD ["./main"]
