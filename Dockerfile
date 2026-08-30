# ---------- Build stage ----------
FROM golang:alpine AS builder

WORKDIR /app

# Install certificates and git for dependencies
RUN apk add --no-cache ca-certificates git

# Copy dependency files first for better Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a static Linux binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server .

# ---------- Final stage ----------
FROM scratch

# Copy CA certificates in case your API makes HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy only the compiled binary
COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]