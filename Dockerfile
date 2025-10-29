# Use the official Golang image as the base for building
FROM golang:1.25.2-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app -ldflags "-w -s" .

# Use a minimal Alpine Linux image for the final production image
FROM alpine:latest
RUN apk update && \
    apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/*
WORKDIR /root/
COPY --from=builder /app/app .
CMD ["./app"]