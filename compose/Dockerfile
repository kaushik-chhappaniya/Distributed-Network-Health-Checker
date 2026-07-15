FROM golang:1.26-alpine3.24
WORKDIR /app
COPY go.mod ./
RUN go mod download
RUN ls -la /app
COPY internal/ ./internal/
COPY utils/ ./utils/
COPY cmd/ ./cmd/
RUN apk add --no-cache docker-cli
RUN go build -o health-checker ./cmd/health-checker/main.go
CMD ["./health-checker"]