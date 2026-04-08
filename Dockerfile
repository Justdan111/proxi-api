# Stage 1: Build
FROM golang:1.25.2-alpine AS builder

WORKDIR /app

# Install dependencies first (Docker cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server

# Stage 2: Run (tiny image)
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/server .

EXPOSE 8080

CMD ["./server"]