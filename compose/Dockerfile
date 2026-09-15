# ----- Stage 1 : Build Go Binary ------
FROM golang:1.27-alpine3.24 AS builder
WORKDIR /app

# Copy dependency files and download them
COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN ls -la /app

# Copy source code and build the binary
COPY internal/ ./internal/
COPY utils/ ./utils/
COPY cmd/ ./cmd/
RUN CGO_ENABLED=0 GOOS=linux go build -o health-checker ./cmd/health-checker/main.go


# ----- Stage 2 :  Final lightweight execution environment ------
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache docker-cli && \
    mkdir -p /app/logs /app/output

# Copy binary from build stage
COPY --from=builder /app/health-checker .
EXPOSE 8080

CMD ["./health-checker"]